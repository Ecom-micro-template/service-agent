package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Ecom-micro-template/service-agent/internal/domain"
	"github.com/Ecom-micro-template/service-agent/internal/infrastructure/persistence"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CategoryCommissionHandler handles category commission operations
type CategoryCommissionHandler struct {
	repo persistence.CategoryCommissionRepository
}

// NewCategoryCommissionHandler creates a new category commission handler
func NewCategoryCommissionHandler(db *gorm.DB) *CategoryCommissionHandler {
	return &CategoryCommissionHandler{
		repo: persistence.NewCategoryCommissionRepository(db),
	}
}

// GetAgentCategoryCommissions gets category-specific commissions for an agent
func (h *CategoryCommissionHandler) GetAgentCategoryCommissions(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := strconv.ParseUint(agentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	commissions, err := h.repo.GetByAgentID(c.Request.Context(), uint(agentID))
	if err != nil {
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
func (h *CategoryCommissionHandler) UpdateAgentCategoryCommissions(c *gin.Context) {
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

	// Build commission list
	commissions := make([]domain.AgentCategoryCommission, 0, len(req.Commissions))
	for _, comm := range req.Commissions {
		if comm.CommissionRate > 0 {
			commissions = append(commissions, domain.AgentCategoryCommission{
				AgentID:        uint(agentID),
				CategoryID:     comm.CategoryID,
				CategoryName:   comm.CategoryName,
				CommissionRate: comm.CommissionRate,
				IsActive:       comm.IsActive,
			})
		}
	}

	if err := h.repo.BulkReplace(c.Request.Context(), uint(agentID), commissions); err != nil {
		log.Error().Err(err).Msg("Failed to update category commissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category commissions"})
		return
	}

	log.Info().Uint64("agent_id", agentID).Int("count", len(commissions)).Msg("Category commissions updated")
	c.JSON(http.StatusOK, gin.H{"message": "Category commissions updated successfully"})
}

// Legacy function handlers for backward compatibility with existing routes

// GetAgentCategoryCommissionsLegacy is a legacy function-based handler
// Deprecated: Use CategoryCommissionHandler.GetAgentCategoryCommissions instead
func GetAgentCategoryCommissionsLegacy(db *gorm.DB) gin.HandlerFunc {
	handler := NewCategoryCommissionHandler(db)
	return handler.GetAgentCategoryCommissions
}

// UpdateAgentCategoryCommissionsLegacy is a legacy function-based handler
// Deprecated: Use CategoryCommissionHandler.UpdateAgentCategoryCommissions instead
func UpdateAgentCategoryCommissionsLegacy(db *gorm.DB) gin.HandlerFunc {
	handler := NewCategoryCommissionHandler(db)
	return handler.UpdateAgentCategoryCommissions
}
