package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niaga-platform/service-agent/internal/config"
	"github.com/niaga-platform/service-agent/internal/database"
	"github.com/niaga-platform/service-agent/internal/handlers"
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

	// Setup Gin
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

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
		// Agent routes
		v1.POST("/agents", handlers.CreateAgent)
		v1.GET("/agents", handlers.GetAgents)
		v1.GET("/agents/:id", handlers.GetAgent)
		v1.PUT("/agents/:id", handlers.UpdateAgent)
		v1.DELETE("/agents/:id", handlers.DeleteAgent)
		v1.GET("/agents/:id/stats", handlers.GetAgentStats)

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
	}

	// Start server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Info().Str("addr", addr).Msg("Agent Service started")

	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
