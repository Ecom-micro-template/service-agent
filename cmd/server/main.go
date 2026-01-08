package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	libmiddleware "github.com/niaga-platform/lib-common/middleware"
	"github.com/niaga-platform/service-agent/internal/config"
	"github.com/niaga-platform/service-agent/internal/database"
	"github.com/niaga-platform/service-agent/internal/handlers"
	"github.com/niaga-platform/service-agent/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting Agent Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().
		Str("environment", cfg.Environment).
		Int("port", cfg.ServerPort).
		Msg("Configuration loaded")

	// Initialize database
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	// Get database instance for handlers
	_ = database.GetDB() // db not used after removing admin handlers

	// Initialize admin handler - removed duplicate
	// adminAgentHandler := handlers.NewAdminAgentHandler(db)

	// Setup Gin
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS - use environment-based configuration
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:3002,http://localhost:3003")
	router.Use(libmiddleware.CORSWithOrigins(allowedOrigins))

	// Security headers
	router.Use(libmiddleware.SecurityHeaders())

	// Input validation
	router.Use(libmiddleware.InputValidation())

	// Rate limiting (50 requests per minute)
	rateLimiter := libmiddleware.NewRateLimiter(50, 100)
	rateLimiter.CleanupLimiters()

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Admin Agent routes (CRUD) - under /agents for backwards compatibility
		v1.POST("/agents", handlers.CreateAgent)
		v1.GET("/agents", handlers.GetAgents)
		v1.GET("/agents/:id", handlers.GetAgent)
		v1.PUT("/agents/:id", handlers.UpdateAgent)
		v1.DELETE("/agents/:id", handlers.DeleteAgent)
		v1.GET("/agents/:id/stats", handlers.GetAgentStats)

		// Agent Category Commission routes
		v1.GET("/agents/:id/category-commissions", handlers.GetAgentCategoryCommissions)
		v1.PUT("/agents/:id/category-commissions", handlers.UpdateAgentCategoryCommissions)

		// Password reset route
		v1.PUT("/agents/:id/reset-password", handlers.ResetAgentPassword)

		// Commission routes
		v1.POST("/commissions", handlers.CreateCommission)
		v1.GET("/agents/:id/commissions", handlers.GetAgentCommissions)
		v1.GET("/commissions/pending", handlers.GetPendingCommissions)
		v1.PUT("/commissions/:id/approve", handlers.ApproveCommission)

		// Payout routes
		v1.POST("/payouts", handlers.CreatePayout)
		v1.GET("/agents/:id/payouts", handlers.GetAgentPayouts)
		v1.GET("/payouts/:id", handlers.GetPayout)
		v1.PUT("/payouts/:id/mark-paid", handlers.MarkPayoutPaid)

		// Agent Portal routes (for frontend - require agent auth)
		agent := v1.Group("/agent")
		agent.Use(middleware.AgentAuthMiddleware())
		{
			agent.GET("/profile", handlers.GetAgentProfile)
			agent.GET("/dashboard", handlers.GetAgentDashboard)
			agent.GET("/orders", handlers.GetAgentOrders)
			agent.POST("/orders", handlers.CreateAgentOrder)
			agent.GET("/orders/:id", handlers.GetAgentOrder)
			agent.GET("/customers", handlers.GetAgentCustomers)
			agent.POST("/customers", handlers.CreateAgentCustomer)
			agent.GET("/customers/:id", handlers.GetAgentCustomer)
			agent.PUT("/customers/:id", handlers.UpdateAgentCustomer)
			agent.GET("/commissions", handlers.GetAgentCommissions)
			agent.GET("/performance", handlers.GetAgentPerformance)
			agent.GET("/team", handlers.GetAgentTeam)
		}

		// Admin routes (require admin middleware)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware()) // Parse JWT and set user_role
		admin.Use(libmiddleware.RequireAdmin())     // Verify admin role
		{
			// Agent management
			admin.GET("/agents", handlers.GetAgents)
			admin.POST("/agents", handlers.CreateAgent)
			admin.GET("/agents/:id", handlers.GetAgent)
			admin.PUT("/agents/:id", handlers.UpdateAgent)
			admin.DELETE("/agents/:id", handlers.DeleteAgent)
			admin.GET("/agents/:id/stats", handlers.GetAgentStats)
			admin.GET("/agents/:id/commissions", handlers.GetAgentCommissions)
			admin.GET("/agents/:id/payouts", handlers.GetAgentPayouts)
			admin.GET("/agents/:id/category-commissions", handlers.GetAgentCategoryCommissions)
			admin.PUT("/agents/:id/category-commissions", handlers.UpdateAgentCategoryCommissions)
			admin.PUT("/agents/:id/reset-password", handlers.ResetAgentPassword)

			// Commission management
			admin.GET("/commissions", handlers.GetPendingCommissions)
			admin.POST("/commissions", handlers.CreateCommission)
			admin.PUT("/commissions/:id/approve", handlers.ApproveCommission)

			// Payout management
			admin.POST("/payouts", handlers.CreatePayout)
			admin.GET("/payouts/:id", handlers.GetPayout)
			admin.PUT("/payouts/:id/mark-paid", handlers.MarkPayoutPaid)
		}
	}

	// Start server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Info().Str("addr", addr).Msg("Agent Service started")

	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
