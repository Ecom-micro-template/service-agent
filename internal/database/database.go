package database

import (
	"fmt"

	"github.com/niaga-platform/service-agent/internal/config"
	"github.com/niaga-platform/service-agent/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase(cfg *config.Config) error {
	var err error

	// Setup GORM logger
	gormLogger := logger.Default
	if cfg.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Connect to database
	DB, err = gorm.Open(postgres.Open(cfg.GetDatabaseURL()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().
		Str("host", cfg.DatabaseHost).
		Int("port", cfg.DatabasePort).
		Str("database", cfg.DatabaseName).
		Msg("Connected to database")

	// Auto migrate models
	if err := DB.AutoMigrate(
		&models.Agent{},
		&models.Commission{},
		&models.Payout{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Info().Msg("Database migrations completed")

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
