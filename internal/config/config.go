package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// Config holds all application configuration.
type Config struct {
	Daemon      DaemonConfig
	Database    DatabaseConfig
	API         APIConfig
	Credentials CredentialsConfig
}

// DaemonConfig holds daemon configuration.
type DaemonConfig struct {
	FetchInterval   time.Duration
	DisplayInterval time.Duration
	EnableEmojis    bool
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Type            string
	SQLitePath      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        string

	// PostgreSQL-specific
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// APIConfig holds API server configuration.
type APIConfig struct {
	Port int
}

// CredentialsConfig holds LibreView credentials.
type CredentialsConfig struct {
	Email    string
	Password string
}

// Load loads all application configuration from environment variables.
// Returns error if any required configuration is missing or invalid.
func Load() (*Config, error) {
	config := &Config{}

	// Load daemon config
	daemonCfg, err := loadDaemonConfig()
	if err != nil {
		return nil, fmt.Errorf("daemon config: %w", err)
	}
	config.Daemon = daemonCfg

	// Load database config
	dbCfg, err := loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("database config: %w", err)
	}
	config.Database = dbCfg

	// Load API config
	apiCfg, err := loadAPIConfig()
	if err != nil {
		return nil, fmt.Errorf("API config: %w", err)
	}
	config.API = apiCfg

	// Load credentials
	credsCfg, err := loadCredentialsConfig()
	if err != nil {
		return nil, fmt.Errorf("credentials config: %w", err)
	}
	config.Credentials = credsCfg

	return config, nil
}

// loadDaemonConfig loads daemon configuration using existing daemon package.
func loadDaemonConfig() (DaemonConfig, error) {
	cfg, err := daemon.LoadConfigFromEnv()
	if err != nil {
		return DaemonConfig{}, err
	}

	return DaemonConfig{
		FetchInterval:   cfg.FetchInterval,
		DisplayInterval: cfg.DisplayInterval,
		EnableEmojis:    cfg.EnableEmojis,
	}, nil
}

// loadDatabaseConfig loads database configuration with validation.
func loadDatabaseConfig() (DatabaseConfig, error) {
	// Use existing persistence package loader
	cfg := persistence.LoadDatabaseConfigFromEnv()

	// Add validation for PostgreSQL
	if cfg.Type == "postgres" && cfg.Password == "" {
		return DatabaseConfig{}, fmt.Errorf("GLCMD_DB_PASSWORD is required for PostgreSQL")
	}

	return DatabaseConfig{
		Type:            cfg.Type,
		SQLitePath:      cfg.SQLitePath,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		LogLevel:        cfg.LogLevel,
		Host:            cfg.Host,
		Port:            cfg.Port,
		Database:        cfg.Database,
		Username:        cfg.Username,
		Password:        cfg.Password,
		SSLMode:         cfg.SSLMode,
	}, nil
}

// loadAPIConfig loads API server configuration with validation.
func loadAPIConfig() (APIConfig, error) {
	port := 8080 // Default port

	if portStr := os.Getenv("GLCMD_API_PORT"); portStr != "" {
		parsedPort, err := strconv.Atoi(portStr)
		if err != nil {
			return APIConfig{}, fmt.Errorf("invalid GLCMD_API_PORT: %w (must be a number)", err)
		}
		if parsedPort < 1 || parsedPort > 65535 {
			return APIConfig{}, fmt.Errorf("invalid GLCMD_API_PORT: %d (must be between 1 and 65535)", parsedPort)
		}
		port = parsedPort
	}

	return APIConfig{Port: port}, nil
}

// loadCredentialsConfig loads LibreView credentials with validation.
func loadCredentialsConfig() (CredentialsConfig, error) {
	email := os.Getenv("GLCMD_EMAIL")
	if email == "" {
		return CredentialsConfig{}, fmt.Errorf("GLCMD_EMAIL environment variable is required")
	}

	password := os.Getenv("GLCMD_PASSWORD")
	if password == "" {
		return CredentialsConfig{}, fmt.Errorf("GLCMD_PASSWORD environment variable is required")
	}

	return CredentialsConfig{
		Email:    email,
		Password: password,
	}, nil
}

// ToPersistenceConfig converts DatabaseConfig to persistence.DatabaseConfig for backward compatibility.
func (c *DatabaseConfig) ToPersistenceConfig() *persistence.DatabaseConfig {
	return &persistence.DatabaseConfig{
		Type:            c.Type,
		SQLitePath:      c.SQLitePath,
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
		LogLevel:        c.LogLevel,
		Host:            c.Host,
		Port:            c.Port,
		Database:        c.Database,
		Username:        c.Username,
		Password:        c.Password,
		SSLMode:         c.SSLMode,
	}
}

// ToDaemonConfig converts DaemonConfig to daemon.Config for backward compatibility.
func (c *DaemonConfig) ToDaemonConfig() *daemon.Config {
	return &daemon.Config{
		FetchInterval:   c.FetchInterval,
		DisplayInterval: c.DisplayInterval,
		EnableEmojis:    c.EnableEmojis,
	}
}
