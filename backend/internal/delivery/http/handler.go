package http

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"vehicle-commerce/backend/internal/config"
	"vehicle-commerce/backend/internal/domain"
	"vehicle-commerce/backend/internal/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	svc *usecase.Service
	cfg *config.Config
}

func NewHandler(svc *usecase.Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.Default()

	// CORS config
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	api := r.Group("/api/v1")
	{
		api.GET("/health", h.HealthCheck)
		api.POST("/auth/login", h.Login)

		// Public Catalog & Branches
		api.GET("/branches", h.GetBranches)
		api.GET("/categories", h.GetCategories)
		api.GET("/products", h.GetProducts)
		api.GET("/products/:id", h.GetProductByID)
		api.GET("/vehicles", h.GetVehicleUnits)

		// Public Tools & Services
		api.POST("/calculator/loan", h.CalculateLoan)
		api.POST("/trade-in", h.SubmitTradeIn)
		api.POST("/orders", h.CreateOrder)

		// Admin & Staff Operations
		admin := api.Group("/admin")
		admin.Use(h.AuthMiddleware(), h.ActiveAdminMiddleware())
		{
			admin.GET("/me", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN", "STAFF"), h.GetCurrentAdmin)
			admin.GET("/orders", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.GetOrders)
			admin.GET("/orders/lookup", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN", "STAFF"), h.GetOrderForPayment)
			admin.PATCH("/orders/:id/status", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.UpdateOrderStatus)
			admin.PATCH("/orders/:id/delivery-fee", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.ConfirmOrderDeliveryFee)
			admin.POST("/orders/:id/payments", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN", "STAFF"), h.RecordOrderPayment)
			admin.GET("/inventory", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN", "STAFF"), h.GetBranchInventory)
			admin.GET("/inventory/transfers", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.GetStockTransfers)
			admin.POST("/inventory/transfers", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.CreateStockTransfer)
			admin.PATCH("/inventory/transfers/:id/:action", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.UpdateStockTransfer)
			admin.POST("/inventory/stock", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.AddBranchStock)
			admin.POST("/inventory/vehicles", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.AddVehicleInventory)
			admin.GET("/inventory/history", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN", "STAFF"), h.GetStockMovements)
			admin.GET("/trade-in", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.GetTradeInRequests)
			admin.PATCH("/trade-in/:id/appraise", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.AppraiseTradeIn)
			admin.GET("/installments", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.GetInstallments)
			admin.PATCH("/installments/:id/status", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.UpdateInstallmentStatus)
			admin.GET("/installment-schedules", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.GetInstallmentSchedules)
			admin.PATCH("/installment-schedules/:id/payment", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN", "STAFF"), h.RecordInstallmentPayment)
			admin.POST("/installment-schedules/:id/reminder", h.RequireRoles("SUPER_ADMIN", "BRANCH_ADMIN"), h.SendInstallmentReminder)
			admin.GET("/users", h.RequireRoles("SUPER_ADMIN"), h.GetAdminUsers)
			admin.POST("/users", h.RequireRoles("SUPER_ADMIN"), h.CreateAdminUser)
			admin.PATCH("/users/:id", h.RequireRoles("SUPER_ADMIN"), h.UpdateAdminUser)
			admin.GET("/telegram/status", h.RequireRoles("SUPER_ADMIN"), h.GetTelegramStatus)
			admin.GET("/telegram/logs", h.RequireRoles("SUPER_ADMIN"), h.GetNotificationLogs)
			admin.DELETE("/telegram/logs", h.RequireRoles("SUPER_ADMIN"), h.ClearNotificationLogs)
			admin.POST("/telegram/logs/:id/resend", h.RequireRoles("SUPER_ADMIN"), h.ResendNotification)
			admin.POST("/telegram/test", h.RequireRoles("SUPER_ADMIN"), h.TestTelegramAlert)
		}
	}

	return r
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "Vehicle E-commerce & Multi-Branch Backend API",
		"version": "1.0.0",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetBranches(c *gin.Context) {
	branches, err := h.svc.GetBranches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": branches})
}

func (h *Handler) GetCategories(c *gin.Context) {
	cats, err := h.svc.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cats})
}

func (h *Handler) GetProducts(c *gin.Context) {
	cat := c.Query("category")
	search := c.Query("search")

	products, err := h.svc.GetProducts(cat, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *Handler) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	product, err := h.svc.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *Handler) GetVehicleUnits(c *gin.Context) {
	prodID := c.Query("product_id")
	branchID := c.Query("branch_id")

	items, err := h.svc.GetVehicleItems(prodID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) CalculateLoan(c *gin.Context) {
	var req domain.LoanCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := h.svc.CalculateInstallment(req)
	c.JSON(http.StatusOK, res)
}

func (h *Handler) SubmitTradeIn(c *gin.Context) {
	var req domain.TradeInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.SubmitTradeIn(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Trade-in request submitted successfully", "data": res})
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.svc.PlaceOrder(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Order placed successfully", "data": order})
}

// Admin Handlers
func (h *Handler) GetOrders(c *gin.Context) {
	branchID := c.Query("branch_id")
	status := c.Query("status")

	orders, err := h.svc.GetOrders(branchID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func (h *Handler) GetCurrentAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":        c.GetString("user_id"),
		"role":      c.GetString("role"),
		"branch_id": c.GetString("branch_id"),
	})
}

func (h *Handler) GetOrderForPayment(c *gin.Context) {
	reference := strings.TrimSpace(c.Query("reference"))
	if reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order number is required"})
		return
	}
	order, err := h.svc.GetOrderByReference(reference, c.GetString("branch_id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch body.Status {
	case "PENDING", "CONFIRMED", "PREPARING", "PROCESSING", "READY_FOR_DELIVERY", "READY_FOR_PICKUP", "COMPLETED", "CANCELLED":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order status"})
		return
	}
	err := h.svc.UpdateOrderStatus(c.Param("id"), body.Status, c.GetString("branch_id"), c.GetString("user_id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

func (h *Handler) ConfirmOrderDeliveryFee(c *gin.Context) {
	var body struct {
		DeliveryFee *float64 `json:"delivery_fee" binding:"required,gte=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.svc.ConfirmOrderDeliveryFee(c.Param("id"), *body.DeliveryFee, c.GetString("branch_id"), c.GetString("user_id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Delivery fee confirmed"})
}

func (h *Handler) RecordOrderPayment(c *gin.Context) {
	var body struct {
		Amount          float64 `json:"amount" binding:"required,gt=0"`
		PaymentMethod   string  `json:"payment_method" binding:"required"`
		ReferenceNumber string  `json:"reference_number"`
		Notes           string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch body.PaymentMethod {
	case "CASH", "KHQR", "BANK_TRANSFER", "CARD", "OTHER":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}
	err := h.svc.RecordOrderPayment(c.Param("id"), body.Amount, body.PaymentMethod, body.ReferenceNumber, body.Notes, c.GetString("branch_id"), c.GetString("user_id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Payment recorded"})
}

func (h *Handler) GetBranchInventory(c *gin.Context) {
	branchID := c.Query("branch_id")
	inv, err := h.svc.GetBranchInventory(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": inv})
}

func (h *Handler) GetStockTransfers(c *gin.Context) {
	transfers, err := h.svc.GetStockTransfers(c.GetString("branch_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": transfers})
}

func (h *Handler) CreateStockTransfer(c *gin.Context) {
	var req domain.InterBranchTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if c.GetString("role") != "SUPER_ADMIN" && req.FromBranchID != c.GetString("branch_id") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Transfers must originate from your assigned branch"})
		return
	}
	id, err := h.svc.CreateStockTransfer(&req, c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Stock transfer request created", "id": id})
}

func (h *Handler) UpdateStockTransfer(c *gin.Context) {
	action := c.Param("action")
	if action != "approve" && action != "receive" && action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer action"})
		return
	}
	err := h.svc.UpdateStockTransfer(c.Param("id"), action, c.GetString("branch_id"), c.GetString("user_id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found or not available for this action"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock transfer updated"})
}

func (h *Handler) AddBranchStock(c *gin.Context) {
	var body struct {
		BranchID     string `json:"branch_id" binding:"required"`
		ProductID    string `json:"product_id" binding:"required"`
		Quantity     int    `json:"quantity" binding:"required,min=1"`
		MinThreshold int    `json:"min_threshold" binding:"min=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if c.GetString("role") != "SUPER_ADMIN" && body.BranchID != c.GetString("branch_id") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Stock can only be adjusted for your assigned branch"})
		return
	}
	if err := h.svc.AddBranchStock(body.BranchID, body.ProductID, body.Quantity, body.MinThreshold, c.GetString("user_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Stock added"})
}

func (h *Handler) AddVehicleInventory(c *gin.Context) {
	var req domain.CreateVehicleItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if c.GetString("role") != "SUPER_ADMIN" && req.BranchID != c.GetString("branch_id") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Vehicle stock can only be added to your assigned branch"})
		return
	}
	if err := h.svc.CreateVehicleItem(&req, c.GetString("user_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Vehicle added to stock"})
}

func (h *Handler) GetStockMovements(c *gin.Context) {
	movements, err := h.svc.GetStockMovements(c.GetString("branch_id"), c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": movements})
}

func (h *Handler) TransferInventory(c *gin.Context) {
	var req domain.InterBranchTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if c.GetString("role") != "SUPER_ADMIN" && req.FromBranchID != c.GetString("branch_id") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Transfers must originate from your assigned branch"})
		return
	}

	err := h.svc.InterBranchTransfer(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inter-branch inventory transfer completed successfully"})
}

func (h *Handler) GetTradeInRequests(c *gin.Context) {
	branchID := c.Query("branch_id")
	list, err := h.svc.GetTradeInRequests(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *Handler) AppraiseTradeIn(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		AppraisedValue float64 `json:"appraised_value" binding:"required"`
		Notes          string  `json:"notes"`
		Status         string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := body.Status
	if status == "" {
		status = "APPRAISED"
	}

	if status != "APPRAISED" && status != "REJECTED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be APPRAISED or REJECTED"})
		return
	}
	err := h.svc.UpdateTradeInAppraisal(id, body.AppraisedValue, body.Notes, status, c.GetString("branch_id"), c.GetString("user_id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trade-in request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Trade-in appraisal updated successfully"})
}

func (h *Handler) GetInstallments(c *gin.Context) {
	branchID := c.Query("branch_id")
	apps, err := h.svc.GetInstallmentApps(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": apps})
}

func (h *Handler) UpdateInstallmentStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Status != "UNDER_REVIEW" && body.Status != "APPROVED" && body.Status != "REJECTED" && body.Status != "DISBURSED" && body.Status != "SETTLED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid installment status"})
		return
	}
	err := h.svc.UpdateInstallmentStatus(id, body.Status, body.Notes, c.GetString("branch_id"), c.GetString("user_id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Installment application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Installment status updated"})
}

func (h *Handler) GetInstallmentSchedules(c *gin.Context) {
	schedules, err := h.svc.GetInstallmentSchedules(c.GetString("branch_id"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

func (h *Handler) RecordInstallmentPayment(c *gin.Context) {
	if err := h.svc.RecordInstallmentPayment(c.Param("id"), c.GetString("branch_id"), c.GetString("user_id")); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pending installment payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Installment payment recorded"})
}

func (h *Handler) SendInstallmentReminder(c *gin.Context) {
	err := h.svc.SendInstallmentReminder(c.Param("id"), c.GetString("branch_id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment schedule not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payment reminder sent"})
}

func (h *Handler) GetAdminUsers(c *gin.Context) {
	users, err := h.svc.GetAdminUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *Handler) CreateAdminUser(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != domain.RoleSuperAdmin && req.Role != domain.RoleBranchAdmin && req.Role != domain.RoleStaff {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin role"})
		return
	}
	if req.Role != domain.RoleSuperAdmin && req.BranchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Branch assignment is required for branch admins and staff"})
		return
	}
	user, err := h.svc.CreateAdminUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *Handler) UpdateAdminUser(c *gin.Context) {
	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != domain.RoleSuperAdmin && req.Role != domain.RoleBranchAdmin && req.Role != domain.RoleStaff {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin role"})
		return
	}
	if req.Status != "ACTIVE" && req.Status != "DISABLED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be ACTIVE or DISABLED"})
		return
	}
	if req.Role != domain.RoleSuperAdmin && req.BranchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Branch assignment is required for branch admins and staff"})
		return
	}
	err := h.svc.UpdateAdminUser(c.Param("id"), c.GetString("user_id"), &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Admin user updated"})
}

func (h *Handler) TestTelegramAlert(c *gin.Context) {
	var body struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SendTelegramTest(body.Message); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Telegram test message sent successfully",
		"status":  "ok",
	})
}

func (h *Handler) GetTelegramStatus(c *gin.Context) {
	status, err := h.svc.GetTelegramStatus()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "data": status})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": status})
}

func (h *Handler) GetNotificationLogs(c *gin.Context) {
	logs, err := h.svc.GetNotificationLogs(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *Handler) ClearNotificationLogs(c *gin.Context) {
	if err := h.svc.ClearNotificationLogs(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notification logs cleared"})
}

func (h *Handler) ResendNotification(c *gin.Context) {
	err := h.svc.ResendNotification(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed notification not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notification resent"})
}

// Auth Middleware
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") || len(authHeader) <= len("Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(h.cfg.JWTSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
			return
		}
		role, ok := claims["role"].(string)
		if !ok || role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token role"})
			return
		}
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func (h *Handler) ActiveAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := h.svc.GetUserByID(c.GetString("user_id"))
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Admin account no longer exists"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to verify admin account"})
			return
		}
		if !applyAdminIdentity(c, user) {
			return
		}
		c.Next()
	}
}

func applyAdminIdentity(c *gin.Context, user *domain.User) bool {
	if user.Status != "ACTIVE" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin account is disabled"})
		return false
	}
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBranchAdmin && user.Role != domain.RoleStaff {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return false
	}
	c.Set("user_id", user.ID)
	c.Set("role", string(user.Role))
	if user.BranchID != nil && *user.BranchID != "" {
		branchID := *user.BranchID
		requestedBranchID := c.Query("branch_id")
		if user.Role != domain.RoleSuperAdmin && requestedBranchID != "" && requestedBranchID != branchID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access is limited to your assigned branch"})
			return false
		}
		if user.Role != domain.RoleSuperAdmin {
			c.Set("branch_id", branchID)
			query := c.Request.URL.Query()
			query.Set("branch_id", branchID)
			c.Request.URL.RawQuery = query.Encode()
		}
	} else if user.Role != domain.RoleSuperAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "A branch assignment is required"})
		return false
	}
	return true
}

func (h *Handler) RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allowed[c.GetString("role")]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		c.Next()
	}
}
