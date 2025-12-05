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

	// Connect to database with FK disabled for migration
	DB, err = gorm.Open(postgres.Open(cfg.GetDatabaseURL()), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().
		Str("host", cfg.DatabaseHost).
		Int("port", cfg.DatabasePort).
		Str("database", cfg.DatabaseName).
		Msg("Connected to database")

	// Disable FK constraints for migration (circular reference: Agent ↔ Team)
	DB.Exec("SET session_replication_role = replica")

	// Auto migrate models - migrate all at once since FK constraints are disabled
	if err := DB.AutoMigrate(
		&models.Agent{},
		&models.Team{},
		&models.Commission{},
		&models.Payout{},
		&models.Customer{},
		&models.Order{},
	); err != nil {
		DB.Exec("SET session_replication_role = DEFAULT")
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// Re-enable FK constraints
	DB.Exec("SET session_replication_role = DEFAULT")

	log.Info().Msg("Database migrations completed")

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
