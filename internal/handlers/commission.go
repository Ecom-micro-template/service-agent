package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Ecom-micro-template/service-agent/internal/database"
	"github.com/Ecom-micro-template/service-agent/internal/domain"
	"github.com/rs/zerolog/log"
)

type CreateCommissionRequest struct {
	AgentID    uint    `json:"agent_id" binding:"required"`
	OrderID    string  `json:"order_id" binding:"required"`
	OrderTotal float64 `json:"order_total" binding:"required,gt=0"`
	Rate       float64 `json:"rate"`
}

// CreateCommission creates a new commission record
func CreateCommission(c *gin.Context) {
	var req CreateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agent to use their commission rate if not specified
	var agent models.Agent
	if err := database.GetDB().First(&agent, req.AgentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Use agent's commission rate if not provided
	rate := req.Rate
	if rate == 0 {
		rate = agent.CommissionRate
	}

	// Calculate commission amount
	amount := models.CalculateCommission(req.OrderTotal, rate)

	commission := models.Commission{
		AgentID:    req.AgentID,
		OrderID:    req.OrderID,
		OrderTotal: req.OrderTotal,
		Rate:       rate,
		Amount:     amount,
		Status:     "pending",
	}

	if err := database.GetDB().Create(&commission).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create commission")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create commission"})
		return
	}

	log.Info().Uint("commission_id", commission.ID).Float64("amount", amount).Msg("Commission created")
	c.JSON(http.StatusCreated, commission)
}

// GetAgentCommissionsByID retrieves all commissions for an agent by ID (admin function)
func GetAgentCommissionsByID(c *gin.Context) {
	agentID := c.Param("id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	offset := (page - 1) * limit

	var commissions []models.Commission
	query := database.GetDB().Where("agent_id = ?", agentID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Model(&models.Commission{}).Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&commissions).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch commissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commissions"})
		return
	}

	// Calculate totals
	var totalAmount float64
	database.GetDB().Model(&models.Commission{}).
		Where("agent_id = ?", agentID).
		Select("COALESCE(SUM(amount), 0)").
		Row().
		Scan(&totalAmount)

	var pendingAmount float64
	database.GetDB().Model(&models.Commission{}).
		Where("agent_id = ? AND status = ?", agentID, "pending").
		Select("COALESCE(SUM(amount), 0)").
		Row().
		Scan(&pendingAmount)

	c.JSON(http.StatusOK, gin.H{
		"data":           commissions,
		"total":          total,
		"page":           page,
		"limit":          limit,
		"total_pages":    (total + int64(limit) - 1) / int64(limit),
		"total_amount":   totalAmount,
		"pending_amount": pendingAmount,
	})
}

// ApproveCommission approves a pending commission
func ApproveCommission(c *gin.Context) {
	id := c.Param("id")

	var commission models.Commission
	if err := database.GetDB().First(&commission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Commission not found"})
		return
	}

	commission.Status = "approved"
	if err := database.GetDB().Save(&commission).Error; err != nil {
		log.Error().Err(err).Msg("Failed to approve commission")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve commission"})
		return
	}

	// Update agent's total earned
	var agent models.Agent
	if err := database.GetDB().First(&agent, commission.AgentID).Error; err == nil {
		agent.TotalEarned += commission.Amount
		database.GetDB().Save(&agent)
	}

	log.Info().Uint("commission_id", commission.ID).Msg("Commission approved")
	c.JSON(http.StatusOK, commission)
}

// GetPendingCommissions retrieves all pending commissions
func GetPendingCommissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	offset := (page - 1) * limit

	var commissions []models.Commission
	var total int64

	database.GetDB().Model(&models.Commission{}).Where("status = ?", "pending").Count(&total)

	if err := database.GetDB().
		Where("status = ?", "pending").
		Preload("Agent").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&commissions).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch pending commissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending commissions"})
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
