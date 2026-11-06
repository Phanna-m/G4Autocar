package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"vehicle-commerce/backend/internal/domain"

	"github.com/google/uuid"
)

func isVehicleCategory(categoryType string) bool {
	return categoryType == "GASOLINE_VEHICLE" || categoryType == "ELECTRIC_VEHICLE"
}

func (db *DB) CreateOrder(req *domain.CreateOrderRequest, totalAmount float64) (*domain.Order, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Find or create customer
	var customerID string
	err = tx.QueryRow(`SELECT id FROM customers WHERE phone = $1 LIMIT 1`, req.CustomerPhone).Scan(&customerID)
	if err == sql.ErrNoRows {
		customerID = uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO customers (id, full_name, phone, email, created_at)
			VALUES ($1, $2, $3, $4, NOW())`,
			customerID, req.CustomerName, req.CustomerPhone, req.CustomerEmail)
		if err != nil {
			return nil, fmt.Errorf("failed to create customer: %w", err)
		}
	} else if err != nil {
		return nil, err
	}

	// 2. Generate Order Number
	orderNo := fmt.Sprintf("ORD-%s-%d", time.Now().Format("20060102"), time.Now().Unix()%100000)
	orderID := uuid.New().String()
	deliveryFeeStatus := "NOT_REQUIRED"
	if req.DeliveryMethod == "DELIVERY" {
		deliveryFeeStatus = "TO_CONFIRM"
	}

	_, err = tx.Exec(`
		INSERT INTO orders (id, order_no, branch_id, customer_id, payment_method, subtotal, discount_amount, tax_amount, total_amount, delivery_method, delivery_address, delivery_fee_status, status, payment_status, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0.00, 0.00, $7, $8, NULLIF($9, ''), $10, 'PENDING', 'UNPAID', $11, NOW())`,
		orderID, orderNo, req.BranchID, customerID, req.PaymentMethod, totalAmount, totalAmount,
		req.DeliveryMethod, req.DeliveryAddress, deliveryFeeStatus, req.Notes)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// 3. Process items
	for _, it := range req.Items {
		itemUUID := uuid.New().String()
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}

		var prodName, categoryType string
		var price float64
		err = tx.QueryRow(`
			SELECT p.name, p.base_price, c.type
			FROM products p
			JOIN categories c ON c.id = p.category_id
			WHERE p.id = $1`, it.ProductID).Scan(&prodName, &price, &categoryType)
		if err != nil {
			return nil, fmt.Errorf("product %s not found: %w", it.ProductID, err)
		}

		sub := price * float64(qty)
		_, err = tx.Exec(`
			INSERT INTO order_items (id, order_id, product_id, vehicle_item_id, item_name, unit_price, quantity, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			itemUUID, orderID, it.ProductID, it.VehicleItemID, prodName, price, qty, sub)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}

		// Enquiries without a VIN do not reserve a vehicle unit.
		if it.VehicleItemID != nil && *it.VehicleItemID != "" {
			result, err := tx.Exec(`
				UPDATE vehicle_items SET status = 'RESERVED'
				WHERE id = $1 AND branch_id = $2 AND product_id = $3 AND status = 'AVAILABLE'`,
				*it.VehicleItemID, req.BranchID, it.ProductID)
			if err != nil {
				return nil, fmt.Errorf("failed to reserve vehicle unit: %w", err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return nil, err
			}
			if affected == 0 {
				return nil, fmt.Errorf("vehicle unit is unavailable at the selected branch")
			}
			if err := addStockMovement(tx, req.BranchID, it.ProductID, "", "RESERVATION", -qty, ""); err != nil {
				return nil, err
			}
		} else if !isVehicleCategory(categoryType) {
			result, err := tx.Exec(`
				UPDATE branch_inventory 
				SET quantity = quantity - $1, reserved_qty = reserved_qty + $1
				WHERE branch_id = $2 AND product_id = $3 AND quantity >= $1`,
				qty, req.BranchID, it.ProductID)
			if err != nil {
				return nil, fmt.Errorf("failed to reserve product stock: %w", err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return nil, err
			}
			if affected == 0 {
				return nil, fmt.Errorf("insufficient stock for product %s", it.ProductID)
			}
			if err := addStockMovement(tx, req.BranchID, it.ProductID, "", "RESERVATION", -qty, ""); err != nil {
				return nil, err
			}
		}
	}

	// 4. Audit Log
	_, _ = tx.Exec(`
		INSERT INTO audit_logs (id, branch_id, action, entity_name, entity_id, created_at)
		VALUES ($1, $2, 'ORDER_CREATED', 'orders', $3, NOW())`,
		uuid.New().String(), req.BranchID, orderID)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &domain.Order{
		ID:                orderID,
		OrderNo:           orderNo,
		BranchID:          req.BranchID,
		CustomerID:        customerID,
		CustomerName:      req.CustomerName,
		CustomerPhone:     req.CustomerPhone,
		PaymentMethod:     req.PaymentMethod,
		DeliveryMethod:    req.DeliveryMethod,
		DeliveryAddress:   req.DeliveryAddress,
		DeliveryFeeStatus: deliveryFeeStatus,
		TotalAmount:       totalAmount,
		Status:            "PENDING",
		PaymentStatus:     "UNPAID",
		Notes:             req.Notes,
		CreatedAt:         time.Now(),
	}, nil
}

func (db *DB) GetOrders(branchID, status string) ([]domain.Order, error) {
	query := `
		SELECT o.id, o.order_no, o.branch_id, b.name as branch_name, o.customer_id, c.full_name, c.phone,
		       o.payment_method, o.delivery_method, COALESCE(o.delivery_address, ''), o.delivery_fee,
		       o.delivery_fee_status, o.subtotal, o.discount_amount, o.tax_amount, o.total_amount,
		       COALESCE((SELECT SUM(p.amount) FROM payments p WHERE p.order_id = o.id), 0),
		       GREATEST(o.total_amount - COALESCE((SELECT SUM(p.amount) FROM payments p WHERE p.order_id = o.id), 0), 0),
		       o.status, o.payment_status, COALESCE(o.notes, ''), o.created_at
		FROM orders o
		JOIN branches b ON o.branch_id = b.id
		JOIN customers c ON o.customer_id = c.id
		WHERE 1=1`

	var args []interface{}
	idx := 1
	if branchID != "" {
		query += fmt.Sprintf(" AND o.branch_id = $%d", idx)
		args = append(args, branchID)
		idx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND o.status = $%d", idx)
		args = append(args, status)
		idx++
	}
	query += " ORDER BY o.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.BranchID, &o.BranchName, &o.CustomerID, &o.CustomerName, &o.CustomerPhone,
			&o.PaymentMethod, &o.DeliveryMethod, &o.DeliveryAddress, &o.DeliveryFee, &o.DeliveryFeeStatus,
			&o.Subtotal, &o.DiscountAmount, &o.TaxAmount, &o.TotalAmount, &o.PaidAmount, &o.BalanceDue,
			&o.Status, &o.PaymentStatus, &o.Notes, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (db *DB) GetOrderByReference(reference, branchID string) (*domain.Order, error) {
	var order domain.Order
	err := db.QueryRow(`
		SELECT o.id, o.order_no, o.branch_id, b.name, o.customer_id, c.full_name, c.phone,
		       o.payment_method, o.delivery_method, COALESCE(o.delivery_address, ''), o.delivery_fee,
		       o.delivery_fee_status, o.subtotal, o.discount_amount, o.tax_amount, o.total_amount,
		       COALESCE(SUM(p.amount), 0), GREATEST(o.total_amount - COALESCE(SUM(p.amount), 0), 0),
		       o.status, o.payment_status, COALESCE(o.notes, ''), o.created_at
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		JOIN customers c ON c.id = o.customer_id
		LEFT JOIN payments p ON p.order_id = o.id
		WHERE (o.order_no = $1 OR o.id::text = $1)
		  AND ($2 = '' OR o.branch_id = $2)
		GROUP BY o.id, b.name, c.full_name, c.phone`, reference, branchID).Scan(
		&order.ID, &order.OrderNo, &order.BranchID, &order.BranchName, &order.CustomerID,
		&order.CustomerName, &order.CustomerPhone, &order.PaymentMethod, &order.DeliveryMethod,
		&order.DeliveryAddress, &order.DeliveryFee, &order.DeliveryFeeStatus, &order.Subtotal,
		&order.DiscountAmount, &order.TaxAmount, &order.TotalAmount, &order.PaidAmount,
		&order.BalanceDue, &order.Status, &order.PaymentStatus, &order.Notes, &order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (db *DB) ConfirmOrderDeliveryFee(id string, fee float64, branchID, actorID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var orderBranchID, deliveryMethod, paymentStatus string
	err = tx.QueryRow(`
		SELECT branch_id, delivery_method, payment_status
		FROM orders WHERE id = $1 FOR UPDATE`, id).
		Scan(&orderBranchID, &deliveryMethod, &paymentStatus)
	if err != nil {
		return err
	}
	if branchID != "" && orderBranchID != branchID {
		return sql.ErrNoRows
	}
	if deliveryMethod != "DELIVERY" {
		return fmt.Errorf("delivery fee can only be set for delivery orders")
	}
	if paymentStatus != "UNPAID" {
		return fmt.Errorf("delivery fee cannot be changed after payment has started")
	}
	if fee < 0 {
		return fmt.Errorf("delivery fee cannot be negative")
	}
	if _, err := tx.Exec(`
		UPDATE orders
		SET delivery_fee = $1,
		    delivery_fee_status = 'CONFIRMED',
		    total_amount = subtotal - discount_amount + tax_amount + $1,
		    status = CASE WHEN status = 'PENDING' THEN 'CONFIRMED' ELSE status END,
		    updated_at = NOW()
		WHERE id = $2`, fee, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO audit_logs (id, user_id, branch_id, action, entity_name, entity_id, created_at)
		VALUES ($1, NULLIF($2, '')::uuid, $3, 'DELIVERY_FEE_CONFIRMED', 'orders', $4, NOW())`,
		uuid.New().String(), actorID, orderBranchID, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) UpdateOrderStatus(id, status, branchID, actorID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var order domain.Order
	err = tx.QueryRow(`
		SELECT id, branch_id, status FROM orders WHERE id = $1 FOR UPDATE`, id).
		Scan(&order.ID, &order.BranchID, &order.Status)
	if err != nil {
		return err
	}
	if branchID != "" && order.BranchID != branchID {
		return sql.ErrNoRows
	}
	if order.Status == "CANCELLED" || order.Status == "COMPLETED" {
		return fmt.Errorf("cannot change status of a %s order", strings.ToLower(order.Status))
	}

	if status == "CANCELLED" {
		rows, err := tx.Query(`
			SELECT oi.product_id, c.type, oi.vehicle_item_id, oi.quantity
			FROM order_items oi
			JOIN products p ON p.id = oi.product_id
			JOIN categories c ON c.id = p.category_id
			WHERE oi.order_id = $1`, id)
		if err != nil {
			return err
		}
		type itemToRelease struct {
			productID    string
			categoryType string
			vehicleID    sql.NullString
			quantity     int
		}
		var items []itemToRelease
		for rows.Next() {
			var item itemToRelease
			if err := rows.Scan(&item.productID, &item.categoryType, &item.vehicleID, &item.quantity); err != nil {
				rows.Close()
				return err
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		for _, item := range items {
			if item.vehicleID.Valid {
				result, err := tx.Exec(`UPDATE vehicle_items SET status = 'AVAILABLE' WHERE id = $1 AND status = 'RESERVED'`, item.vehicleID.String)
				if err != nil {
					return err
				}
				affected, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if affected > 0 {
					if err := addStockMovement(tx, order.BranchID, item.productID, "", "RELEASE", item.quantity, ""); err != nil {
						return err
					}
				}
			} else if !isVehicleCategory(item.categoryType) {
				if _, err := tx.Exec(`
					UPDATE branch_inventory
					SET quantity = quantity + $1, reserved_qty = GREATEST(reserved_qty - $1, 0), updated_at = NOW()
					WHERE branch_id = $2 AND product_id = $3`, item.quantity, order.BranchID, item.productID); err != nil {
					return err
				}
				if err := addStockMovement(tx, order.BranchID, item.productID, "", "RELEASE", item.quantity, ""); err != nil {
					return err
				}
			}
		}
	} else if status == "COMPLETED" {
		if _, err := tx.Exec(`
			UPDATE vehicle_items vi SET status = 'SOLD'
			FROM order_items oi
			WHERE oi.order_id = $1 AND oi.vehicle_item_id = vi.id AND vi.status = 'RESERVED'`, id); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			UPDATE branch_inventory bi SET reserved_qty = GREATEST(bi.reserved_qty - oi.quantity, 0), updated_at = NOW()
			FROM order_items oi
			WHERE oi.order_id = $1 AND oi.vehicle_item_id IS NULL
			  AND bi.branch_id = $2 AND bi.product_id = oi.product_id`, id, order.BranchID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO audit_logs (id, user_id, branch_id, action, entity_name, entity_id, created_at)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, 'orders', $5, NOW())`,
		uuid.New().String(), actorID, order.BranchID, "ORDER_STATUS_"+status, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) RecordOrderPayment(id string, amount float64, method, reference, notes, branchID, actorID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var total float64
	var orderBranchID string
	if err := tx.QueryRow(`SELECT total_amount, branch_id FROM orders WHERE id = $1 FOR UPDATE`, id).Scan(&total, &orderBranchID); err != nil {
		return err
	}
	if branchID != "" && orderBranchID != branchID {
		return sql.ErrNoRows
	}
	var paid float64
	if err := tx.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE order_id = $1`, id).Scan(&paid); err != nil {
		return err
	}
	if amount <= 0 || amount > total-paid+0.009 {
		return fmt.Errorf("payment amount must be greater than zero and not exceed the balance due")
	}
	if _, err := tx.Exec(`
		INSERT INTO payments (id, order_id, amount, payment_method, reference_number, notes, recorded_by_user_id)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, '')::uuid)`,
		uuid.New().String(), id, amount, method, reference, notes, actorID); err != nil {
		return err
	}
	newStatus := "PARTIAL"
	if paid+amount >= total-0.009 {
		newStatus = "PAID"
	}
	if _, err := tx.Exec(`UPDATE orders SET payment_status = $1, updated_at = NOW() WHERE id = $2`, newStatus, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO audit_logs (id, user_id, branch_id, action, entity_name, entity_id, created_at)
		VALUES ($1, NULLIF($2, '')::uuid, $3, 'PAYMENT_RECORDED', 'orders', $4, NOW())`,
		uuid.New().String(), actorID, orderBranchID, id); err != nil {
		return err
	}
	return tx.Commit()
}

// Installment Applications
func (db *DB) CreateInstallmentApp(app *domain.InstallmentApplication) error {
	app.ID = uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO installment_applications (id, order_id, customer_id, branch_id, vehicle_price, down_payment, down_payment_percentage, loan_amount, interest_rate_annual, term_months, monthly_installment, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'SUBMITTED', NOW())`,
		app.ID, app.OrderID, app.CustomerID, app.BranchID, app.VehiclePrice, app.DownPayment, app.DownPaymentPercentage, app.LoanAmount, app.InterestRateAnnual, app.TermMonths, app.MonthlyInstallment)
	return err
}

func (db *DB) GetInstallmentApps(branchID string) ([]domain.InstallmentApplication, error) {
	query := `
		SELECT ia.id, ia.order_id, o.order_no, ia.customer_id, c.full_name, ia.branch_id, b.name,
		       ia.vehicle_price, ia.down_payment, ia.down_payment_percentage, ia.loan_amount, ia.interest_rate_annual, ia.term_months, ia.monthly_installment,
		       ia.status, COALESCE(ia.reviewer_notes, ''), ia.created_at
		FROM installment_applications ia
		JOIN orders o ON ia.order_id = o.id
		JOIN customers c ON ia.customer_id = c.id
		JOIN branches b ON ia.branch_id = b.id
		WHERE 1=1`

	var args []interface{}
	if branchID != "" {
		query += " AND ia.branch_id = $1"
		args = append(args, branchID)
	}
	query += " ORDER BY ia.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.InstallmentApplication
	for rows.Next() {
		var a domain.InstallmentApplication
		if err := rows.Scan(&a.ID, &a.OrderID, &a.OrderNo, &a.CustomerID, &a.CustomerName, &a.BranchID, &a.BranchName,
			&a.VehiclePrice, &a.DownPayment, &a.DownPaymentPercentage, &a.LoanAmount, &a.InterestRateAnnual, &a.TermMonths, &a.MonthlyInstallment,
			&a.Status, &a.ReviewerNotes, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (db *DB) UpdateInstallmentStatus(id, status, notes, branchID, reviewerID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE installment_applications
		SET status = $1, reviewer_notes = $2, reviewed_by_user_id = NULLIF($3, '')::uuid, updated_at = NOW()
		WHERE id = $4`
	args := []interface{}{status, notes, reviewerID, id}
	if branchID != "" {
		query += " AND branch_id = $5"
		args = append(args, branchID)
	}
	result, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	if status == "APPROVED" {
		_, err = tx.Exec(`
			INSERT INTO installment_payment_schedules
				(id, installment_application_id, customer_id, due_date, amount)
			SELECT uuid_generate_v4(), ia.id, ia.customer_id,
			       (CURRENT_DATE + (months.month_offset || ' months')::interval)::date,
			       ia.monthly_installment
			FROM installment_applications ia
			CROSS JOIN generate_series(1, ia.term_months) AS months(month_offset)
			WHERE ia.id = $1
			ON CONFLICT (installment_application_id, due_date) DO NOTHING`, id)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Trade-In Requests
func (db *DB) CreateTradeInRequest(req *domain.TradeInRequest) error {
	req.ID = uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO trade_in_requests (id, customer_id, target_vehicle_product_id, target_branch_id, trade_in_type, current_brand, current_model, current_year, current_mileage, condition_description, customer_expected_price, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'PENDING_INSPECTION', NOW())`,
		req.ID, req.CustomerID, req.TargetVehicleProductID, req.TargetBranchID, req.TradeInType, req.CurrentBrand, req.CurrentModel, req.CurrentYear, req.CurrentMileage, req.ConditionDescription, req.CustomerExpectedPrice)
	return err
}

func (db *DB) GetTradeInRequests(branchID string) ([]domain.TradeInRequest, error) {
	query := `
		SELECT tr.id, tr.customer_id, c.full_name, c.phone, tr.target_vehicle_product_id, COALESCE(p.name, ''), tr.target_branch_id, b.name,
		       tr.trade_in_type, tr.current_brand, tr.current_model, tr.current_year, tr.current_mileage, COALESCE(tr.condition_description, ''),
		       COALESCE(tr.customer_expected_price, 0), COALESCE(tr.appraised_value, 0), COALESCE(tr.appraisal_notes, ''), tr.status, tr.created_at
		FROM trade_in_requests tr
		JOIN customers c ON tr.customer_id = c.id
		JOIN branches b ON tr.target_branch_id = b.id
		LEFT JOIN products p ON tr.target_vehicle_product_id = p.id
		WHERE 1=1`

	var args []interface{}
	if branchID != "" {
		query += " AND tr.target_branch_id = $1"
		args = append(args, branchID)
	}
	query += " ORDER BY tr.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.TradeInRequest
	for rows.Next() {
		var t domain.TradeInRequest
		if err := rows.Scan(&t.ID, &t.CustomerID, &t.CustomerName, &t.CustomerPhone, &t.TargetVehicleProductID, &t.TargetProductName,
			&t.TargetBranchID, &t.TargetBranchName, &t.TradeInType, &t.CurrentBrand, &t.CurrentModel, &t.CurrentYear, &t.CurrentMileage,
			&t.ConditionDescription, &t.CustomerExpectedPrice, &t.AppraisedValue, &t.AppraisalNotes, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (db *DB) UpdateTradeInAppraisal(id string, appraisedValue float64, notes, status, branchID, reviewerID string) error {
	query := `
		UPDATE trade_in_requests 
		SET appraised_value = $1, appraisal_notes = $2, status = $3, appraised_by_user_id = NULLIF($4, '')::uuid, updated_at = NOW() 
		WHERE id = $5`
	args := []interface{}{appraisedValue, notes, status, reviewerID, id}
	if branchID != "" {
		query += " AND target_branch_id = $6"
		args = append(args, branchID)
	}
	result, err := db.Exec(query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Inter-Branch Inventory Transfer
func (db *DB) TransferInventory(req *domain.InterBranchTransferRequest) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Deduct from source branch
	res, err := tx.Exec(`
		UPDATE branch_inventory 
		SET quantity = quantity - $1, updated_at = NOW() 
		WHERE branch_id = $2 AND product_id = $3 AND quantity >= $1`,
		req.Quantity, req.FromBranchID, req.ProductID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient stock in source branch for transfer")
	}

	// 2. Add to destination branch (Upsert)
	_, err = tx.Exec(`
		INSERT INTO branch_inventory (id, branch_id, product_id, quantity, reserved_qty, min_threshold, selling_price, updated_at)
		VALUES ($1, $2, $3, $4, 0, 3, (SELECT base_price FROM products WHERE id = $3), NOW())
		ON CONFLICT (branch_id, product_id) 
		DO UPDATE SET quantity = branch_inventory.quantity + EXCLUDED.quantity, updated_at = NOW()`,
		uuid.New().String(), req.ToBranchID, req.ProductID, req.Quantity)
	if err != nil {
		return err
	}

	// 3. Log audit
	_, _ = tx.Exec(`
		INSERT INTO audit_logs (id, branch_id, action, entity_name, entity_id, created_at)
		VALUES ($1, $2, 'INTER_BRANCH_TRANSFER', 'branch_inventory', $3, NOW())`,
		uuid.New().String(), req.FromBranchID, req.ProductID)

	return tx.Commit()
}
