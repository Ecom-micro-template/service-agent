package database

import (
	"fmt"
	"time"

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

	// Configure connection pooling for production performance
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings optimized for VPS (4GB RAM)
	sqlDB.SetMaxIdleConns(10)              // Keep 10 idle connections ready
	sqlDB.SetMaxOpenConns(50)              // Max 50 concurrent connections
	sqlDB.SetConnMaxLifetime(time.Hour)    // Recycle connections after 1 hour
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections after 10 minutes

	log.Info().
		Str("host", cfg.DatabaseHost).
		Int("port", cfg.DatabasePort).
		Str("database", cfg.DatabaseName).
		Int("max_open_conns", 50).
		Int("max_idle_conns", 10).
		Msg("Connected to database with connection pooling")

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
