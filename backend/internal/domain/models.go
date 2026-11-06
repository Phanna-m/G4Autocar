package domain

import (
	"encoding/json"
	"time"
)

type UserRole string

const (
	RoleSuperAdmin  UserRole = "SUPER_ADMIN"
	RoleBranchAdmin UserRole = "BRANCH_ADMIN"
	RoleStaff       UserRole = "STAFF"
	RoleCustomer    UserRole = "CUSTOMER"
)

type Branch struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	City           string    `json:"city"`
	Address        string    `json:"address"`
	Phone          string    `json:"phone"`
	TelegramChatID string    `json:"telegram_chat_id,omitempty"`
	IsHeadquarters bool      `json:"is_headquarters"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type User struct {
	ID           string    `json:"id"`
	BranchID     *string   `json:"branch_id,omitempty"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	Role         UserRole  `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Type        string    `json:"type"` // ELECTRIC_VEHICLE, GASOLINE_VEHICLE, BICYCLE, SPARE_PART
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID             string          `json:"id"`
	CategoryID     string          `json:"category_id"`
	CategoryName   string          `json:"category_name,omitempty"`
	CategoryType   string          `json:"category_type,omitempty"`
	SKU            string          `json:"sku"`
	Name           string          `json:"name"`
	Brand          string          `json:"brand"`
	Model          string          `json:"model"`
	Year           int             `json:"year"`
	BasePrice      float64         `json:"base_price"`
	Condition      string          `json:"condition"` // NEW, USED, CERTIFIED
	Description    string          `json:"description"`
	Images         json.RawMessage `json:"images"`
	Specs          json.RawMessage `json:"specs"`
	IsActive       bool            `json:"is_active"`
	AvailableUnits int             `json:"available_units,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type VehicleItem struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	ProductName  string    `json:"product_name,omitempty"`
	BranchID     string    `json:"branch_id"`
	BranchName   string    `json:"branch_name,omitempty"`
	VIN          string    `json:"vin"`
	EngineNo     string    `json:"engine_no"`
	Color        string    `json:"color"`
	MileageKM    int       `json:"mileage_km"`
	PlateNumber  string    `json:"plate_number,omitempty"`
	CostPrice    float64   `json:"cost_price,omitempty"`
	SellingPrice float64   `json:"selling_price"`
	Status       string    `json:"status"` // AVAILABLE, RESERVED, SOLD, IN_TRANSIT
	CreatedAt    time.Time `json:"created_at"`
}

type BranchInventory struct {
	ID           string    `json:"id"`
	BranchID     string    `json:"branch_id"`
	BranchName   string    `json:"branch_name,omitempty"`
	ProductID    string    `json:"product_id"`
	ProductName  string    `json:"product_name,omitempty"`
	Quantity     int       `json:"quantity"`
	ReservedQty  int       `json:"reserved_qty"`
	AvailableQty int       `json:"available_qty"`
	MinThreshold int       `json:"min_threshold"`
	UnitCost     float64   `json:"unit_cost,omitempty"`
	SellingPrice float64   `json:"selling_price"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StockTransfer struct {
	ID             string    `json:"id"`
	FromBranchID   string    `json:"from_branch_id"`
	FromBranchName string    `json:"from_branch_name"`
	ToBranchID     string    `json:"to_branch_id"`
	ToBranchName   string    `json:"to_branch_name"`
	ProductID      string    `json:"product_id"`
	ProductName    string    `json:"product_name"`
	Quantity       int       `json:"quantity"`
	Status         string    `json:"status"`
	Notes          string    `json:"notes,omitempty"`
	RequestedAt    time.Time `json:"requested_at"`
}

type StockMovement struct {
	ID              string    `json:"id"`
	BranchID        string    `json:"branch_id"`
	BranchName      string    `json:"branch_name"`
	ProductID       string    `json:"product_id"`
	ProductName     string    `json:"product_name"`
	TransactionType string    `json:"transaction_type"`
	Quantity        int       `json:"quantity"`
	PerformedBy     string    `json:"performed_by"`
	CreatedAt       time.Time `json:"created_at"`
}

type InstallmentPaymentSchedule struct {
	ID           string     `json:"id"`
	CustomerID   string     `json:"customer_id"`
	CustomerName string     `json:"customer_name"`
	OrderNo      string     `json:"order_no"`
	BranchID     string     `json:"branch_id"`
	BranchName   string     `json:"branch_name"`
	DueDate      time.Time  `json:"due_date"`
	Amount       float64    `json:"monthly_amount"`
	Status       string     `json:"payment_status"`
	OverdueDays  int        `json:"overdue_days"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
}

type NotificationLog struct {
	ID             string    `json:"id"`
	EventType      string    `json:"event_type"`
	TargetChannel  string    `json:"target_channel"`
	MessagePayload string    `json:"message_payload"`
	DeliveryStatus string    `json:"delivery_status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	CreatedAt      time.Time `json:"timestamp"`
}

type CreateUserRequest struct {
	BranchID string   `json:"branch_id"`
	Username string   `json:"username" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	FullName string   `json:"full_name" binding:"required"`
	Phone    string   `json:"phone"`
	Role     UserRole `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	BranchID string   `json:"branch_id"`
	Status   string   `json:"status" binding:"required"`
	Role     UserRole `json:"role" binding:"required"`
}

type Customer struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id,omitempty"`
	FullName   string    `json:"full_name"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	NationalID string    `json:"national_id,omitempty"`
	Address    string    `json:"address,omitempty"`
	TelegramID string    `json:"telegram_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Order struct {
	ID                string      `json:"id"`
	OrderNo           string      `json:"order_no"`
	BranchID          string      `json:"branch_id"`
	BranchName        string      `json:"branch_name,omitempty"`
	CustomerID        string      `json:"customer_id"`
	CustomerName      string      `json:"customer_name,omitempty"`
	CustomerPhone     string      `json:"customer_phone,omitempty"`
	HandledByUserID   *string     `json:"handled_by_user_id,omitempty"`
	PaymentMethod     string      `json:"payment_method"` // FULL_PAYMENT_CASH, FULL_PAYMENT_KHQR, INSTALLMENT, TRADE_IN
	DeliveryMethod    string      `json:"delivery_method"`
	DeliveryAddress   string      `json:"delivery_address,omitempty"`
	DeliveryFee       float64     `json:"delivery_fee"`
	DeliveryFeeStatus string      `json:"delivery_fee_status"`
	Subtotal          float64     `json:"subtotal"`
	DiscountAmount    float64     `json:"discount_amount"`
	TaxAmount         float64     `json:"tax_amount"`
	TotalAmount       float64     `json:"total_amount"`
	PaidAmount        float64     `json:"paid_amount"`
	BalanceDue        float64     `json:"balance_due"`
	Status            string      `json:"status"` // PENDING, CONFIRMED, PROCESSING, READY_FOR_PICKUP, COMPLETED, CANCELLED
	PaymentStatus     string      `json:"payment_status"`
	Notes             string      `json:"notes,omitempty"`
	Items             []OrderItem `json:"items,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"order_id"`
	ProductID     string  `json:"product_id"`
	VehicleItemID *string `json:"vehicle_item_id,omitempty"`
	ItemName      string  `json:"item_name"`
	UnitPrice     float64 `json:"unit_price"`
	Quantity      int     `json:"quantity"`
	Subtotal      float64 `json:"subtotal"`
}

type InstallmentApplication struct {
	ID                    string          `json:"id"`
	OrderID               string          `json:"order_id"`
	OrderNo               string          `json:"order_no,omitempty"`
	CustomerID            string          `json:"customer_id"`
	CustomerName          string          `json:"customer_name,omitempty"`
	BranchID              string          `json:"branch_id"`
	BranchName            string          `json:"branch_name,omitempty"`
	VehiclePrice          float64         `json:"vehicle_price"`
	DownPayment           float64         `json:"down_payment"`
	DownPaymentPercentage float64         `json:"down_payment_percentage"`
	LoanAmount            float64         `json:"loan_amount"`
	InterestRateAnnual    float64         `json:"interest_rate_annual"`
	TermMonths            int             `json:"term_months"`
	MonthlyInstallment    float64         `json:"monthly_installment"`
	EmploymentInfo        json.RawMessage `json:"employment_info,omitempty"`
	GuarantorInfo         json.RawMessage `json:"guarantor_info,omitempty"`
	Status                string          `json:"status"` // SUBMITTED, UNDER_REVIEW, APPROVED, REJECTED, DISBURSED
	ReviewedByUserID      *string         `json:"reviewed_by_user_id,omitempty"`
	ReviewerNotes         string          `json:"reviewer_notes,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
}

type TradeInRequest struct {
	ID                     string    `json:"id"`
	CustomerID             string    `json:"customer_id"`
	CustomerName           string    `json:"customer_name,omitempty"`
	CustomerPhone          string    `json:"customer_phone,omitempty"`
	TargetVehicleProductID *string   `json:"target_vehicle_product_id,omitempty"`
	TargetProductName      string    `json:"target_product_name,omitempty"`
	TargetBranchID         string    `json:"target_branch_id"`
	TargetBranchName       string    `json:"target_branch_name,omitempty"`
	TradeInType            string    `json:"trade_in_type"` // TRADE_IN, BUY_RESELL
	CurrentBrand           string    `json:"current_brand"`
	CurrentModel           string    `json:"current_model"`
	CurrentYear            int       `json:"current_year"`
	CurrentMileage         int       `json:"current_mileage"`
	ConditionDescription   string    `json:"condition_description"`
	CustomerExpectedPrice  float64   `json:"customer_expected_price"`
	AppraisedValue         float64   `json:"appraised_value"`
	AppraisalNotes         string    `json:"appraisal_notes,omitempty"`
	Status                 string    `json:"status"` // PENDING_INSPECTION, APPRAISED, ACCEPTED_BY_CUSTOMER, REJECTED, APPLIED_TO_ORDER
	CreatedAt              time.Time `json:"created_at"`
}

// DTOs
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type LoanCalculationRequest struct {
	VehiclePrice float64 `json:"vehicle_price" binding:"required"`
	DownPayment  float64 `json:"down_payment"`
	InterestRate float64 `json:"interest_rate"` // Annual rate in % (e.g. 9.0)
	TermMonths   int     `json:"term_months"`   // 12, 24, 36, 48, 60, 72
}

type LoanCalculationResponse struct {
	VehiclePrice          float64 `json:"vehicle_price"`
	DownPayment           float64 `json:"down_payment"`
	DownPaymentPercentage float64 `json:"down_payment_percentage"`
	LoanAmount            float64 `json:"loan_amount"`
	InterestRateAnnual    float64 `json:"interest_rate_annual"`
	TermMonths            int     `json:"term_months"`
	MonthlyPayment        float64 `json:"monthly_payment"`
	TotalInterest         float64 `json:"total_interest"`
	TotalRepayment        float64 `json:"total_repayment"`
}

type CreateOrderRequest struct {
	BranchID        string               `json:"branch_id" binding:"required"`
	CustomerName    string               `json:"customer_name" binding:"required"`
	CustomerPhone   string               `json:"customer_phone" binding:"required"`
	CustomerEmail   string               `json:"customer_email"`
	PaymentMethod   string               `json:"payment_method" binding:"required"`
	DeliveryMethod  string               `json:"delivery_method"`
	DeliveryAddress string               `json:"delivery_address"`
	Notes           string               `json:"notes"`
	Items           []CreateOrderItemReq `json:"items" binding:"required,min=1"`
}

type CreateOrderItemReq struct {
	ProductID     string  `json:"product_id" binding:"required"`
	VehicleItemID *string `json:"vehicle_item_id"`
	Quantity      int     `json:"quantity"`
}

type InterBranchTransferRequest struct {
	FromBranchID string `json:"from_branch_id" binding:"required"`
	ToBranchID   string `json:"to_branch_id" binding:"required"`
	ProductID    string `json:"product_id" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required,min=1"`
	Notes        string `json:"notes"`
}

type CreateVehicleItemRequest struct {
	BranchID     string  `json:"branch_id" binding:"required"`
	ProductID    string  `json:"product_id" binding:"required"`
	VIN          string  `json:"vin" binding:"required"`
	EngineNo     string  `json:"engine_no"`
	Color        string  `json:"color" binding:"required"`
	MileageKM    int     `json:"mileage_km" binding:"min=0"`
	PlateNumber  string  `json:"plate_number"`
	CostPrice    float64 `json:"cost_price" binding:"min=0"`
	SellingPrice float64 `json:"selling_price" binding:"required,gt=0"`
}
