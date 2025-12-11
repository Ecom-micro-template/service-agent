package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	// Server
	ServerPort int
	GinMode    string

	// Application
	LogLevel    string
	Environment string

	// Commission
	DefaultCommissionRate float64
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseHost:          getEnv("DB_HOST", "localhost"),
		DatabasePort:          getEnvAsInt("DB_PORT", 5432),
		DatabaseUser:          getEnv("DB_USER", "postgres"),
		DatabasePassword:      getEnv("DB_PASSWORD", "postgres"),
		DatabaseName:          getEnv("DB_NAME", "agent_db"),
		DatabaseSSLMode:       getEnv("DB_SSLMODE", "disable"),
		ServerPort:            getEnvAsInt("APP_PORT", 8006),
		GinMode:               getEnv("GIN_MODE", "debug"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		Environment:           getEnv("APP_ENV", "development"),
		DefaultCommissionRate: getEnvAsFloat("DEFAULT_COMMISSION_RATE", 10.0),
	}

	return cfg, nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseName,
		c.DatabaseSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
