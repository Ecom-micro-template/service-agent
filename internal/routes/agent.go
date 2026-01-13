package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Ecom-micro-template/service-agent/internal/handlers"
	"github.com/Ecom-micro-template/service-agent/internal/middleware"
)

// RegisterAgentRoutes registers all agent portal routes
func RegisterAgentRoutes(r *gin.Engine) {
	// Agent Portal API - requires authentication and agent role
	agentAPI := r.Group("/api/v1/agent")
	agentAPI.Use(middleware.RequireAgent()) // Assumes auth middleware is already applied
	{
		// Profile
		agentAPI.GET("/profile", handlers.GetAgentProfile)
		agentAPI.PUT("/profile", handlers.UpdateAgentProfile)

		// Dashboard
		agentAPI.GET("/dashboard", handlers.GetAgentDashboard)

		// Orders
		agentAPI.GET("/orders", handlers.GetAgentOrders)
		agentAPI.POST("/orders", handlers.CreateAgentOrder)
		agentAPI.GET("/orders/:id", handlers.GetAgentOrder)

		// Customers
		agentAPI.GET("/customers", handlers.GetAgentCustomers)
		agentAPI.POST("/customers", handlers.CreateAgentCustomer)
		agentAPI.GET("/customers/:id", handlers.GetAgentCustomer)
		agentAPI.PUT("/customers/:id", handlers.UpdateAgentCustomer)

		// Commissions
		agentAPI.GET("/commissions", handlers.GetAgentCommissions)

		// Performance
		agentAPI.GET("/performance", handlers.GetAgentPerformance)

		// Team
		agentAPI.GET("/team", handlers.GetAgentTeam)
	}
}

// RegisterAdminAgentRoutes registers admin routes for managing agents
func RegisterAdminAgentRoutes(r *gin.Engine) {
	// Admin API for managing agents - requires admin role
	adminAPI := r.Group("/api/v1/admin/agents")
	// adminAPI.Use(middleware.RequireAdmin()) // Add admin middleware
	{
		adminAPI.POST("", handlers.CreateAgent)
		adminAPI.GET("", handlers.GetAgents)
		adminAPI.GET("/:id", handlers.GetAgent)
		adminAPI.PUT("/:id", handlers.UpdateAgent)
		adminAPI.DELETE("/:id", handlers.DeleteAgent)
	}

	// Commission management
	commissionAPI := r.Group("/api/v1/admin/commissions")
	// commissionAPI.Use(middleware.RequireAdmin())
	{
		commissionAPI.GET("", handlers.GetPendingCommissions)
		commissionAPI.GET("/:id/agent/:agent_id", handlers.GetAgentCommissionsByID)
		commissionAPI.PUT("/:id/approve", handlers.ApproveCommission)
	}

	// Payout management
	payoutAPI := r.Group("/api/v1/admin/payouts")
	// payoutAPI.Use(middleware.RequireAdmin())
	{
		payoutAPI.POST("", handlers.CreatePayout)
		payoutAPI.GET("/:id", handlers.GetPayout)
		payoutAPI.PUT("/:id/mark-paid", handlers.MarkPayoutPaid)
	}
}
