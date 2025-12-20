package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/niaga-platform/service-agent/internal/database"
	"github.com/niaga-platform/service-agent/internal/models"
	"github.com/rs/zerolog/log"
)

// GetAgentCategoryCommissions gets category-specific commissions for an agent
func GetAgentCategoryCommissions(c *gin.Context) {
	agentID := c.Param("id")

	var commissions []models.AgentCategoryCommission
	if err := database.GetDB().Where("agent_id = ?", agentID).Find(&commissions).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch category commissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category commissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": commissions})
}

// UpdateAgentCategoryCommissionsRequest is the request for updating category commissions
type UpdateAgentCategoryCommissionsRequest struct {
	Commissions []struct {
		CategoryID     string  `json:"category_id"`
		CategoryName   string  `json:"category_name"`
		CommissionRate float64 `json:"commission_rate"`
		IsActive       bool    `json:"is_active"`
	} `json:"commissions"`
}

// UpdateAgentCategoryCommissions updates category-specific commissions for an agent
func UpdateAgentCategoryCommissions(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := strconv.ParseUint(agentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var req UpdateAgentCategoryCommissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Delete existing category commissions for this agent
	if err := db.Where("agent_id = ?", agentID).Delete(&models.AgentCategoryCommission{}).Error; err != nil {
		log.Error().Err(err).Msg("Failed to delete existing category commissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category commissions"})
		return
	}

	// Create new category commissions
	for _, comm := range req.Commissions {
		if comm.CommissionRate > 0 {
			newComm := models.AgentCategoryCommission{
				AgentID:        uint(agentID),
				CategoryID:     comm.CategoryID,
				CategoryName:   comm.CategoryName,
				CommissionRate: comm.CommissionRate,
				IsActive:       comm.IsActive,
			}
			if err := db.Create(&newComm).Error; err != nil {
				log.Error().Err(err).Msg("Failed to create category commission")
				// Continue with other commissions
			}
		}
	}

	log.Info().Uint64("agent_id", agentID).Int("count", len(req.Commissions)).Msg("Category commissions updated")
	c.JSON(http.StatusOK, gin.H{"message": "Category commissions updated successfully"})
}
