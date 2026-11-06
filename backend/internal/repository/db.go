package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"vehicle-commerce/backend/internal/config"
	"vehicle-commerce/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

func NewDB(cfg *config.Config) (*DB, error) {
	var connStr string
	if cfg.DatabaseURL != "" {
		connStr = cfg.DatabaseURL
	} else {
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	}

	pgConfig, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database connection: %w", err)
	}

	// Supabase's transaction pooler can switch PostgreSQL backends between
	// requests, so avoid server-side prepared statements on pooled connections.
	pgConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	db := stdlib.OpenDB(*pgConfig)

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Printf("[DB Notice] Database not reachable yet. Error: %v", err)
	} else {
		log.Println("[DB] Successfully connected to PostgreSQL / Supabase!")
	}

	return &DB{db}, nil
}

func (db *DB) RunMigrations(migrationDir string) error {
	setupFile := filepath.Join(migrationDir, "supabase_setup.sql")
	setupApplied := false
	if content, err := os.ReadFile(setupFile); err == nil {
		if _, err := db.Exec(string(content)); err != nil {
			log.Printf("[DB Migrations Notice] Execution notice on supabase_setup.sql: %v", err)
		} else {
			setupApplied = true
			log.Println("[DB Migrations] Applied supabase_setup.sql successfully!")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read migration supabase_setup.sql: %w", err)
	}

	files := migrationFiles(setupApplied)
	if setupApplied {
		log.Println("[DB Migrations] Skipping baseline schema and seed migrations already included in supabase_setup.sql")
	}

	for _, file := range files {
		filePath := filepath.Join(migrationDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", file, err)
		} else {
			log.Printf("[DB Migrations] Applied %s", file)
		}
	}

	return nil
}

func migrationFiles(setupApplied bool) []string {
	if setupApplied {
		return []string{
			"000004_admin_modules.sql",
			"000005_seed_more_vehicles.sql",
			"000006_car_only_catalog.sql",
			"000007_checkout_delivery.sql",
		}
	}
	return []string{
		"000001_init_schema.sql",
		"000002_seed_data.sql",
		"000003_seed_toyota_catalog.sql",
		"000004_admin_modules.sql",
		"000005_seed_more_vehicles.sql",
		"000006_car_only_catalog.sql",
		"000007_checkout_delivery.sql",
	}
}

// Branches
func (db *DB) GetBranches() ([]domain.Branch, error) {
	rows, err := db.Query(`
		SELECT id, code, name, city, address, phone, COALESCE(telegram_chat_id, ''), is_headquarters, is_active, created_at 
		FROM branches ORDER BY is_headquarters DESC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []domain.Branch
	for rows.Next() {
		var b domain.Branch
		if err := rows.Scan(&b.ID, &b.Code, &b.Name, &b.City, &b.Address, &b.Phone, &b.TelegramChatID, &b.IsHeadquarters, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, nil
}

// Categories
func (db *DB) GetCategories() ([]domain.Category, error) {
	rows, err := db.Query(`SELECT id, name, slug, type, COALESCE(description, ''), COALESCE(icon, ''), is_active, created_at FROM categories WHERE is_active = TRUE ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Type, &c.Description, &c.Icon, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// Products
func (db *DB) GetProducts(categorySlug, search string) ([]domain.Product, error) {
	query := `
		SELECT p.id, p.category_id, c.name as category_name, c.type as category_type, p.sku, p.name, p.brand, p.model, p.year, 
		       p.base_price, p.condition, COALESCE(p.description, ''), p.images, p.specs, p.is_active, p.created_at
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE p.is_active = TRUE`

	var args []interface{}
	idx := 1
	if categorySlug != "" && categorySlug != "all" {
		query += fmt.Sprintf(" AND c.slug = $%d", idx)
		args = append(args, categorySlug)
		idx++
	}
	if search != "" {
		query += fmt.Sprintf(" AND (p.name ILIKE $%d OR p.brand ILIKE $%d OR p.model ILIKE $%d)", idx, idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	query += " ORDER BY p.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.CategoryName, &p.CategoryType, &p.SKU, &p.Name, &p.Brand, &p.Model, &p.Year,
			&p.BasePrice, &p.Condition, &p.Description, &p.Images, &p.Specs, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (db *DB) GetProductByID(id string) (*domain.Product, error) {
	row := db.QueryRow(`
		SELECT p.id, p.category_id, c.name, c.type, p.sku, p.name, p.brand, p.model, p.year, 
		       p.base_price, p.condition, COALESCE(p.description, ''), p.images, p.specs, p.is_active, p.created_at
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1`, id)

	var p domain.Product
	err := row.Scan(&p.ID, &p.CategoryID, &p.CategoryName, &p.CategoryType, &p.SKU, &p.Name, &p.Brand, &p.Model, &p.Year,
		&p.BasePrice, &p.Condition, &p.Description, &p.Images, &p.Specs, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Vehicle Units (VIN)
func (db *DB) GetVehicleItems(productID, branchID string) ([]domain.VehicleItem, error) {
	query := `
		SELECT v.id, v.product_id, p.name, v.branch_id, b.name, v.vin, v.engine_no, v.color, v.mileage_km, 
		       COALESCE(v.plate_number, ''), v.selling_price, v.status, v.created_at
		FROM vehicle_items v
		JOIN products p ON v.product_id = p.id
		JOIN branches b ON v.branch_id = b.id
		WHERE 1=1`

	var args []interface{}
	idx := 1
	if productID != "" {
		query += fmt.Sprintf(" AND v.product_id = $%d", idx)
		args = append(args, productID)
		idx++
	}
	if branchID != "" {
		query += fmt.Sprintf(" AND v.branch_id = $%d", idx)
		args = append(args, branchID)
		idx++
	}
	query += " ORDER BY v.status ASC, v.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.VehicleItem
	for rows.Next() {
		var vi domain.VehicleItem
		if err := rows.Scan(&vi.ID, &vi.ProductID, &vi.ProductName, &vi.BranchID, &vi.BranchName, &vi.VIN, &vi.EngineNo,
			&vi.Color, &vi.MileageKM, &vi.PlateNumber, &vi.SellingPrice, &vi.Status, &vi.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, vi)
	}
	return items, nil
}

// Branch Inventory (Spare Parts & Bikes)
func (db *DB) GetBranchInventory(branchID string) ([]domain.BranchInventory, error) {
	query := `
		SELECT stock.id, stock.branch_id, b.name, stock.product_id, p.name,
		       stock.quantity, stock.reserved_qty, GREATEST(stock.quantity - stock.reserved_qty, 0),
		       stock.min_threshold, stock.selling_price, stock.updated_at
		FROM (
			SELECT bi.id::text AS id, bi.branch_id, bi.product_id, bi.quantity, bi.reserved_qty,
			       bi.min_threshold, bi.selling_price, bi.updated_at
			FROM branch_inventory bi
			UNION ALL
			SELECT vi.branch_id::text || ':' || vi.product_id::text AS id, vi.branch_id, vi.product_id,
			       COUNT(*)::int AS quantity,
			       COUNT(*) FILTER (WHERE vi.status = 'RESERVED')::int AS reserved_qty,
			       0 AS min_threshold, MAX(vi.selling_price) AS selling_price, MAX(vi.updated_at) AS updated_at
			FROM vehicle_items vi
			GROUP BY vi.branch_id, vi.product_id
		) stock
		JOIN branches b ON stock.branch_id = b.id
		JOIN products p ON stock.product_id = p.id
		WHERE 1=1`

	var args []interface{}
	if branchID != "" {
		query += " AND stock.branch_id = $1"
		args = append(args, branchID)
	}
	query += " ORDER BY stock.quantity ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.BranchInventory
	for rows.Next() {
		var item domain.BranchInventory
		if err := rows.Scan(&item.ID, &item.BranchID, &item.BranchName, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.ReservedQty, &item.AvailableQty, &item.MinThreshold, &item.SellingPrice, &item.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

// Users
func (db *DB) GetUserByUsername(username string) (*domain.User, error) {
	row := db.QueryRow(`
		SELECT id, branch_id, username, email, password_hash, full_name, COALESCE(phone, ''), role, status, created_at
		FROM users WHERE username = $1 OR email = $1 LIMIT 1`, username)

	var u domain.User
	err := row.Scan(&u.ID, &u.BranchID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.Role, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) GetUserByID(id string) (*domain.User, error) {
	row := db.QueryRow(`
		SELECT id, branch_id, username, email, password_hash, full_name, COALESCE(phone, ''), role, status, created_at
		FROM users WHERE id = $1`, id)

	var user domain.User
	err := row.Scan(
		&user.ID, &user.BranchID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.Phone, &user.Role, &user.Status, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
