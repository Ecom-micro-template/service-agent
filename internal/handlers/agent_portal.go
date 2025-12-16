package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niaga-platform/service-agent/internal/database"
	"github.com/niaga-platform/service-agent/internal/models"
	"github.com/rs/zerolog/log"
)

// GetAgentFromContext retrieves the agent ID from the JWT context
func GetAgentFromContext(c *gin.Context) (uint, error) {
	agentID, exists := c.Get("agent_id")
	if !exists {
		return 0, fmt.Errorf("agent_id not found in context")
	}

	id, ok := agentID.(uint)
	if !ok {
		return 0, fmt.Errorf("invalid agent_id type")
	}

	return id, nil
}

// GetAgentProfile retrieves the authenticated agent's profile
func GetAgentProfile(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var agent models.Agent
	if err := database.GetDB().Preload("Team").Preload("Team.Leader").First(&agent, agentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// UpdateAgentProfileRequest represents the request body for updating agent profile
type UpdateAgentProfileRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// UpdateAgentProfile updates the authenticated agent's profile
func UpdateAgentProfile(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateAgentProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agent models.Agent
	if err := database.GetDB().First(&agent, agentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Phone != "" {
		agent.Phone = req.Phone
	}

	if err := database.GetDB().Save(&agent).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update agent profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	log.Info().Uint("agent_id", agentID).Msg("Agent profile updated")
	c.JSON(http.StatusOK, agent)
}

// GetAgentDashboard retrieves dashboard statistics for the agent
func GetAgentDashboard(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := database.GetDB()
	dashboard := models.Dashboard{}

	// Get current month start
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Total orders and sales
	db.Model(&models.Order{}).Where("agent_id = ?", agentID).Count(&dashboard.TotalOrders)
	db.Model(&models.Order{}).Where("agent_id = ?", agentID).Select("COALESCE(SUM(total), 0)").Scan(&dashboard.TotalSales)

	// Monthly stats
	db.Model(&models.Order{}).
		Where("agent_id = ? AND created_at >= ?", agentID, monthStart).
		Count(&dashboard.MonthlyOrders)
	db.Model(&models.Order{}).
		Where("agent_id = ? AND created_at >= ?", agentID, monthStart).
		Select("COALESCE(SUM(total), 0)").Scan(&dashboard.MonthlySales)

	// Total customers
	db.Model(&models.Customer{}).Where("agent_id = ?", agentID).Count(&dashboard.TotalCustomers)

	// Commission stats
	db.Model(&models.Commission{}).
		Where("agent_id = ?", agentID).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&dashboard.TotalCommission)

	db.Model(&models.Commission{}).
		Where("agent_id = ? AND status = ?", agentID, "pending").
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&dashboard.PendingCommission)

	db.Model(&models.Commission{}).
		Where("agent_id = ? AND status = ?", agentID, "approved").
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&dashboard.ApprovedCommission)

	db.Model(&models.Commission{}).
		Where("agent_id = ? AND status = ?", agentID, "paid").
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&dashboard.PaidCommission)

	db.Model(&models.Commission{}).
		Where("agent_id = ? AND created_at >= ?", agentID, monthStart).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&dashboard.MonthlyCommission)

	// Average order value
	if dashboard.TotalOrders > 0 {
		dashboard.AverageOrderValue = dashboard.TotalSales / float64(dashboard.TotalOrders)
	}

	// Commission breakdown
	dashboard.CommissionBreakdown = models.CommissionBreakdown{
		Pending:  dashboard.PendingCommission,
		Approved: dashboard.ApprovedCommission,
		Paid:     dashboard.PaidCommission,
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetAgentOrders retrieves all orders for the agent
func GetAgentOrders(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	offset := (page - 1) * limit

	var orders []models.Order
	query := database.GetDB().Model(&models.Order{}).Where("agent_id = ?", agentID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        orders,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// CreateAgentOrder creates a new order for the agent
func CreateAgentOrder(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Get customer details
	var customer models.Customer
	if err := db.First(&customer, req.CustomerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Get agent for commission rate
	var agent models.Agent
	if err := db.First(&agent, agentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Calculate total
	var total float64
	for _, item := range req.Items {
		total += item.Price * float64(item.Quantity)
	}

	// Generate order number
	var count int64
	db.Model(&models.Order{}).Count(&count)
	orderNumber := fmt.Sprintf("ORD-%s-%05d", time.Now().Format("20060102"), count+1)

	// Create order
	order := models.Order{
		OrderNumber:    orderNumber,
		AgentID:        agentID,
		CustomerID:     req.CustomerID,
		CustomerName:   customer.Name,
		CustomerEmail:  customer.Email,
		Total:          total,
		Status:         "pending",
		PaymentStatus:  "unpaid",
		CommissionRate: agent.CommissionRate,
		Commission:     total * agent.CommissionRate / 100,
	}

	if err := db.Create(&order).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	log.Info().Uint("agent_id", agentID).Uint("order_id", order.ID).Str("order_number", order.OrderNumber).Msg("Order created")
	c.JSON(http.StatusCreated, order)
}

// GetAgentOrder retrieves a single order
func GetAgentOrder(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orderID := c.Param("id")

	var order models.Order
	if err := database.GetDB().Where("agent_id = ?", agentID).First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetAgentCustomers retrieves all customers for the agent
func GetAgentCustomers(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	offset := (page - 1) * limit

	var customers []models.Customer
	query := database.GetDB().Model(&models.Customer{}).Where("agent_id = ?", agentID)

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&customers).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch customers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        customers,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// CreateAgentCustomer creates a new customer for the agent
func CreateAgentCustomer(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := models.Customer{
		AgentID:  agentID,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
		City:     req.City,
		State:    req.State,
		Postcode: req.Postcode,
	}

	if err := database.GetDB().Create(&customer).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create customer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	log.Info().Uint("agent_id", agentID).Uint("customer_id", customer.ID).Msg("Customer created")
	c.JSON(http.StatusCreated, customer)
}

// GetAgentCustomer retrieves a single customer
func GetAgentCustomer(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	customerID := c.Param("id")

	var customer models.Customer
	if err := database.GetDB().Where("agent_id = ?", agentID).First(&customer, customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// UpdateAgentCustomer updates a customer
func UpdateAgentCustomer(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	customerID := c.Param("id")

	var customer models.Customer
	if err := database.GetDB().Where("agent_id = ?", agentID).First(&customer, customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	var req models.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Address != "" {
		customer.Address = req.Address
	}
	if req.City != "" {
		customer.City = req.City
	}
	if req.State != "" {
		customer.State = req.State
	}
	if req.Postcode != "" {
		customer.Postcode = req.Postcode
	}

	if err := database.GetDB().Save(&customer).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update customer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	log.Info().Uint("agent_id", agentID).Uint("customer_id", customer.ID).Msg("Customer updated")
	c.JSON(http.StatusOK, customer)
}

// GetAgentCommissions retrieves all commissions for the agent
func GetAgentCommissions(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	offset := (page - 1) * limit

	var commissions []models.Commission
	query := database.GetDB().Model(&models.Commission{}).Where("agent_id = ?", agentID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&commissions).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch commissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        commissions,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetAgentPerformance retrieves monthly performance metrics
func GetAgentPerformance(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get last 12 months
	var performances []models.Performance
	db := database.GetDB()

	for i := 11; i >= 0; i-- {
		monthStart := time.Now().AddDate(0, -i, 0)
		monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)

		var perf models.Performance
		perf.Month = monthStart

		// Total sales and orders for this month
		db.Model(&models.Order{}).
			Where("agent_id = ? AND created_at >= ? AND created_at < ?", agentID, monthStart, monthEnd).
			Count(&perf.TotalOrders)

		db.Model(&models.Order{}).
			Where("agent_id = ? AND created_at >= ? AND created_at < ?", agentID, monthStart, monthEnd).
			Select("COALESCE(SUM(total), 0)").
			Scan(&perf.TotalSales)

		// Commission breakdown
		db.Model(&models.Commission{}).
			Where("agent_id = ? AND created_at >= ? AND created_at < ?", agentID, monthStart, monthEnd).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&perf.TotalCommission)

		db.Model(&models.Commission{}).
			Where("agent_id = ? AND status = ? AND created_at >= ? AND created_at < ?", agentID, "pending", monthStart, monthEnd).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&perf.CommissionPending)

		db.Model(&models.Commission{}).
			Where("agent_id = ? AND status = ? AND created_at >= ? AND created_at < ?", agentID, "approved", monthStart, monthEnd).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&perf.CommissionApproved)

		db.Model(&models.Commission{}).
			Where("agent_id = ? AND status = ? AND created_at >= ? AND created_at < ?", agentID, "paid", monthStart, monthEnd).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&perf.CommissionPaid)

		performances = append(performances, perf)
	}

	c.JSON(http.StatusOK, performances)
}

// GetAgentTeam retrieves team information
func GetAgentTeam(c *gin.Context) {
	agentID, err := GetAgentFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get agent with team
	var agent models.Agent
	if err := database.GetDB().Preload("Team").First(&agent, agentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	if agent.Team == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Agent is not part of any team"})
		return
	}

	// Get full team details with members
	var team models.Team
	if err := database.GetDB().Preload("Leader").Preload("Members").First(&team, agent.TeamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}
