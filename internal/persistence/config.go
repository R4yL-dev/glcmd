package persistence

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Type            string        // "sqlite" or "postgres"
	SQLitePath      string        // Path to SQLite file (e.g., "./data/glcmd.db")
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
	LogLevel        string        // GORM log level: "silent", "error", "warn", "info"

	// PostgreSQL-specific (for future use)
	Host     string // PostgreSQL host
	Port     int    // PostgreSQL port
	Database string // PostgreSQL database name
	Username string // PostgreSQL username
	Password string // PostgreSQL password
	SSLMode  string // PostgreSQL SSL mode: "disable", "require", "verify-full"
}

// DefaultSQLiteConfig returns default configuration for SQLite.
func DefaultSQLiteConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:            "sqlite",
		SQLitePath:      "./data/glcmd.db",
		MaxOpenConns:    1,  // SQLite: 1 writer at a time
		MaxIdleConns:    1,  // Keep connection alive
		ConnMaxLifetime: time.Hour,
		LogLevel:        "warn", // Errors + warnings
	}
}

// LoadDatabaseConfigFromEnv loads database configuration from environment variables.
// Falls back to default SQLite config if no env vars are set.
func LoadDatabaseConfigFromEnv() *DatabaseConfig {
	config := DefaultSQLiteConfig()

	// SQLite configuration
	if path := os.Getenv("GLCMD_DB_PATH"); path != "" {
		config.SQLitePath = path
	}

	if logLevel := os.Getenv("GLCMD_DB_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// PostgreSQL configuration (future)
	if dbType := os.Getenv("GLCMD_DB_TYPE"); dbType == "postgres" {
		config.Type = "postgres"
		config.Host = getEnvOrDefault("GLCMD_DB_HOST", "localhost")
		config.Port = getEnvAsIntOrDefault("GLCMD_DB_PORT", 5432)
		config.Database = getEnvOrDefault("GLCMD_DB_NAME", "glcmd")
		config.Username = getEnvOrDefault("GLCMD_DB_USER", "glcmd")
		config.Password = os.Getenv("GLCMD_DB_PASSWORD") // No default for security
		config.SSLMode = getEnvOrDefault("GLCMD_DB_SSL_MODE", "require")

		// PostgreSQL can handle more connections
		config.MaxOpenConns = getEnvAsIntOrDefault("GLCMD_DB_MAX_OPEN_CONNS", 10)
		config.MaxIdleConns = getEnvAsIntOrDefault("GLCMD_DB_MAX_IDLE_CONNS", 2)
	}

	return config
}

// BuildDSN builds the database connection string (Data Source Name).
func (c *DatabaseConfig) BuildDSN() string {
	switch c.Type {
	case "sqlite":
		// Enable WAL mode for better concurrency
		// Set busy timeout to 5 seconds to avoid "database is locked" errors
		return fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", c.SQLitePath)

	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
		)

	default:
		return ""
	}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
