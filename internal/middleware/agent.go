package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Ecom-micro-template/service-agent/internal/database"
	"github.com/Ecom-micro-template/service-agent/internal/domain"
)

// RequireAgent middleware checks if the authenticated user is an agent
// It extracts the user_id from JWT and verifies agent status
func RequireAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - No user ID"})
			c.Abort()
			return
		}

		// Get user_type from JWT
		userType, exists := c.Get("user_type")
		if !exists || userType != "agent" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied - Agent role required"})
			c.Abort()
			return
		}

		// Get agent record using user_id
		var agent models.Agent
		// Assuming there's a user_id field in agents table or email matching
		// For now, we'll use the ID directly since the existing model uses auto-increment ID
		agentID := userID.(uint)

		if err := database.GetDB().First(&agent, agentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent profile not found"})
			c.Abort()
			return
		}

		// Check if agent is active
		if agent.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Agent account is not active"})
			c.Abort()
			return
		}

		// Set agent_id and agent details in context for handlers
		c.Set("agent_id", agent.ID)
		c.Set("agent", agent)

		c.Next()
	}
}

// OptionalAgent is similar to RequireAgent but doesn't abort if agent not found
// Useful for endpoints that can work with or without agent context
func OptionalAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userType, exists := c.Get("user_type")
		if !exists || userType != "agent" {
			c.Next()
			return
		}

		var agent models.Agent
		agentID := userID.(uint)

		if err := database.GetDB().First(&agent, agentID).Error; err == nil {
			if agent.Status == "active" {
				c.Set("agent_id", agent.ID)
				c.Set("agent", agent)
			}
		}

		c.Next()
	}
}
