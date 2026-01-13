package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/Ecom-micro-template/service-agent/internal/database"
	"github.com/Ecom-micro-template/service-agent/internal/domain"
	"github.com/rs/zerolog/log"
)

// AgentAuthMiddleware verifies JWT and sets agent_id in context
func AgentAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Get JWT secret
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "your-super-secret-jwt-key-change-in-production"
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Error().Err(err).Msg("Failed to parse JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Get user email from claims
		email, ok := claims["email"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
			c.Abort()
			return
		}

		// Get user role from claims
		role, _ := claims["role"].(string)
		if role != "agent" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied - agent role required"})
			c.Abort()
			return
		}

		// Find agent by email
		var agent domain.Agent
		if err := database.GetDB().Where("email = ?", email).First(&agent).Error; err != nil {
			log.Error().Str("email", email).Msg("Agent not found for email")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent not found"})
			c.Abort()
			return
		}

		// Set agent_id in context
		c.Set("agent_id", agent.ID)
		c.Set("agent_email", email)
		c.Set("agent_name", agent.Name)

		c.Next()
	}
}
