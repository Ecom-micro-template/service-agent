package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/niaga-platform/service-agent/internal/database"
	"github.com/niaga-platform/service-agent/internal/models"
	"github.com/rs/zerolog/log"
)

type CreateAgentRequest struct {
	Name           string  `json:"name" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	Phone          string  `json:"phone"`
	CommissionRate float64 `json:"commission_rate"`
}

type UpdateAgentRequest struct {
	Name           string  `json:"name"`
	Email          string  `json:"email" binding:"omitempty,email"`
	Phone          string  `json:"phone"`
	CommissionRate float64 `json:"commission_rate"`
	Status         string  `json:"status"`
}

// CreateAgent creates a new agent
func CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent := models.Agent{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		CommissionRate: req.CommissionRate,
	}

	if err := database.GetDB().Create(&agent).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	log.Info().Uint("agent_id", agent.ID).Str("code", agent.Code).Msg("Agent created")
	c.JSON(http.StatusCreated, agent)
}

// GetAgents lists all agents with pagination
func GetAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	offset := (page - 1) * limit

	var agents []models.Agent
	query := database.GetDB().Model(&models.Agent{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&agents).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch agents")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        agents,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetAgent retrieves a single agent by ID
func GetAgent(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	if err := database.GetDB().Preload("Commissions").Preload("Payouts").First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// UpdateAgent updates an existing agent
func UpdateAgent(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	if err := database.GetDB().First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Email != "" {
		agent.Email = req.Email
	}
	if req.Phone != "" {
		agent.Phone = req.Phone
	}
	if req.CommissionRate > 0 {
		agent.CommissionRate = req.CommissionRate
	}
	if req.Status != "" {
		agent.Status = req.Status
	}

	if err := database.GetDB().Save(&agent).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent"})
		return
	}

	log.Info().Uint("agent_id", agent.ID).Msg("Agent updated")
	c.JSON(http.StatusOK, agent)
}

// DeleteAgent soft deletes an agent
func DeleteAgent(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	if err := database.GetDB().First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	agent.Status = "inactive"
	if err := database.GetDB().Save(&agent).Error; err != nil {
		log.Error().Err(err).Msg("Failed to delete agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}

	log.Info().Uint("agent_id", agent.ID).Msg("Agent deleted")
	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}
