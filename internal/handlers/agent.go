package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/niaga-platform/service-agent/internal/database"
	"github.com/niaga-platform/service-agent/internal/models"
	"github.com/rs/zerolog/log"
)

type CreateAgentRequest struct {
	Name           string  `json:"name" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	Password       string  `json:"password" binding:"required,min=8"`
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

// AuthRegisterRequest is the request to register user with auth service
type AuthRegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// CreateAgent creates a new agent and registers them with auth service
func CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First, register the agent as a user in auth service
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://kilang-auth:8001"
	}

	// Split name into first/last name
	firstName := req.Name
	lastName := ""

	// Register with auth service
	authReq := AuthRegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: firstName,
		LastName:  lastName,
		Role:      "agent", // Role for agents
	}

	authBody, _ := json.Marshal(authReq)
	authResp, err := http.Post(authURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(authBody))
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register agent credentials"})
		return
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusCreated && authResp.StatusCode != http.StatusOK {
		// Try to get error message from auth service
		var authError map[string]interface{}
		json.NewDecoder(authResp.Body).Decode(&authError)
		log.Error().Interface("auth_error", authError).Int("status", authResp.StatusCode).Msg("Auth service registration failed")

		errorMsg := "Failed to register agent credentials"
		if msg, ok := authError["error"].(string); ok {
			errorMsg = msg
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	// Now create the agent record
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

	log.Info().Uint("agent_id", agent.ID).Str("code", agent.Code).Msg("Agent created with auth credentials")
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

// ResetAgentPasswordRequest is the request to reset agent password
type ResetAgentPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

// AuthResetPasswordRequest is the request to reset password in auth service
type AuthResetPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ResetAgentPassword resets an agent's password
func ResetAgentPassword(c *gin.Context) {
	id := c.Param("id")

	// First get the agent to find their email
	var agent models.Agent
	if err := database.GetDB().First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	var req ResetAgentPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call auth service to reset password
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://kilang-auth:8001"
	}

	// Use admin password reset endpoint
	authReq := AuthResetPasswordRequest{
		Email:    agent.Email,
		Password: req.Password,
	}

	authBody, _ := json.Marshal(authReq)
	httpReq, _ := http.NewRequest("PUT", authURL+"/api/v1/admin/users/reset-password-by-email", bytes.NewBuffer(authBody))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	authResp, err := client.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to auth service for password reset")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusOK {
		var authError map[string]interface{}
		json.NewDecoder(authResp.Body).Decode(&authError)
		log.Error().Interface("auth_error", authError).Int("status", authResp.StatusCode).Msg("Auth service password reset failed")

		errorMsg := "Failed to reset password"
		if msg, ok := authError["error"].(string); ok {
			errorMsg = msg
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	log.Info().Uint("agent_id", agent.ID).Str("email", agent.Email).Msg("Agent password reset successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
