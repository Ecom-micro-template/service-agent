package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Ecom-micro-template/service-agent/internal/database"
	"github.com/Ecom-micro-template/service-agent/internal/domain"
	"github.com/rs/zerolog/log"
)

type CreatePayoutRequest struct {
	AgentID uint   `json:"agent_id" binding:"required"`
	Period  string `json:"period" binding:"required"` // Format: YYYY-MM
}

// CreatePayout creates a new payout for approved commissions
func CreatePayout(c *gin.Context) {
	var req CreatePayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get all approved commissions for the agent that haven't been paid
	var commissions []domain.Commission
	if err := database.GetDB().
		Where("agent_id = ? AND status = ?", req.AgentID, "approved").
		Find(&commissions).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch commissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commissions"})
		return
	}

	if len(commissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No approved commissions found"})
		return
	}

	// Calculate total amount and collect commission IDs
	var totalAmount float64
	var commissionIDs []uint
	for _, commission := range commissions {
		totalAmount += commission.Amount
		commissionIDs = append(commissionIDs, commission.ID)
	}

	// Convert commission IDs to JSON
	commissionIDsJSON, _ := json.Marshal(commissionIDs)

	payout := domain.Payout{
		AgentID:       req.AgentID,
		Amount:        totalAmount,
		Period:        req.Period,
		CommissionIDs: string(commissionIDsJSON),
		Status:        "pending",
	}

	if err := database.GetDB().Create(&payout).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create payout")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payout"})
		return
	}

	// Update commissions status to 'paid'
	database.GetDB().Model(&domain.Commission{}).
		Where("id IN ?", commissionIDs).
		Update("status", "paid")

	log.Info().Uint("payout_id", payout.ID).Float64("amount", totalAmount).Msg("Payout created")
	c.JSON(http.StatusCreated, payout)
}

// GetAgentPayouts retrieves all payouts for an agent
func GetAgentPayouts(c *gin.Context) {
	agentID := c.Param("id")

	var payouts []domain.Payout
	if err := database.GetDB().
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Find(&payouts).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch payouts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payouts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": payouts,
	})
}

// GetPayout retrieves a single payout by ID
func GetPayout(c *gin.Context) {
	id := c.Param("id")

	var payout domain.Payout
	if err := database.GetDB().Preload("Agent").First(&payout, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payout not found"})
		return
	}

	c.JSON(http.StatusOK, payout)
}

// MarkPayoutPaid marks a payout as paid
func MarkPayoutPaid(c *gin.Context) {
	id := c.Param("id")

	var payout domain.Payout
	if err := database.GetDB().First(&payout, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payout not found"})
		return
	}

	now := time.Now()
	payout.Status = "paid"
	payout.PaidAt = &now

	if err := database.GetDB().Save(&payout).Error; err != nil {
		log.Error().Err(err).Msg("Failed to mark payout as paid")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark payout as paid"})
		return
	}

	log.Info().Uint("payout_id", payout.ID).Msg("Payout marked as paid")
	c.JSON(http.StatusOK, payout)
}

// GetAgentStats retrieves statistics for an agent
func GetAgentStats(c *gin.Context) {
	agentID := c.Param("id")

	var agent domain.Agent
	if err := database.GetDB().First(&agent, agentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Get total commissions
	var totalCommissions int64
	var totalCommissionAmount float64
	database.GetDB().Model(&domain.Commission{}).
		Where("agent_id = ?", agentID).
		Count(&totalCommissions)
	database.GetDB().Model(&domain.Commission{}).
		Where("agent_id = ?", agentID).
		Select("COALESCE(SUM(amount), 0)").
		Row().
		Scan(&totalCommissionAmount)

	// Get pending commissions
	var pendingCommissions int64
	var pendingAmount float64
	database.GetDB().Model(&domain.Commission{}).
		Where("agent_id = ? AND status = ?", agentID, "pending").
		Count(&pendingCommissions)
	database.GetDB().Model(&domain.Commission{}).
		Where("agent_id = ? AND status = ?", agentID, "pending").
		Select("COALESCE(SUM(amount), 0)").
		Row().
		Scan(&pendingAmount)

	// Get this month's commissions
	currentMonth := time.Now().Format("2006-01")
	var thisMonthAmount float64
	database.GetDB().Model(&domain.Commission{}).
		Where("agent_id = ? AND TO_CHAR(created_at, 'YYYY-MM') = ?", agentID, currentMonth).
		Select("COALESCE(SUM(amount), 0)").
		Row().
		Scan(&thisMonthAmount)

	// Get total payouts
	var totalPayouts int64
	database.GetDB().Model(&domain.Payout{}).
		Where("agent_id = ?", agentID).
		Count(&totalPayouts)

	stats := gin.H{
		"agent":                   agent,
		"total_commissions":       totalCommissions,
		"total_commission_amount": totalCommissionAmount,
		"pending_commissions":     pendingCommissions,
		"pending_amount":          pendingAmount,
		"this_month_amount":       thisMonthAmount,
		"total_payouts":           totalPayouts,
	}

	c.JSON(http.StatusOK, stats)
}
