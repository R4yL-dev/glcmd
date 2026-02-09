package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database manages the database connection and lifecycle.
type Database struct {
	db     *gorm.DB
	config *DatabaseConfig
}

// NewDatabase creates a new database connection based on the provided configuration.
func NewDatabase(config *DatabaseConfig) (*Database, error) {
	// For SQLite, ensure the directory exists
	if config.Type == "sqlite" {
		dir := filepath.Dir(config.SQLitePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	// Select database driver based on type
	var dialector gorm.Dialector
	switch config.Type {
	case "sqlite":
		dialector = sqlite.Open(config.BuildDSN())
	case "postgres":
		dialector = postgres.Open(config.BuildDSN())
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(parseLogLevel(config.LogLevel)),
		NowFunc: func() time.Time {
			return time.Now().UTC() // Always use UTC for consistency
		},
		PrepareStmt: true, // Enable prepared statement cache for better performance
	}

	// Open database connection
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", config.Type, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	slog.Info("database connection established",
		"type", config.Type,
		"maxOpenConns", config.MaxOpenConns,
		"maxIdleConns", config.MaxIdleConns,
	)

	return &Database{
		db:     db,
		config: config,
	}, nil
}

// AutoMigrate runs automatic migration for all GORM models.
// It also handles dropping legacy indexes that have been replaced.
func (d *Database) AutoMigrate(models ...interface{}) error {
	slog.Info("running database migrations", "type", d.config.Type)

	// Drop legacy indexes before AutoMigrate (which cannot remove old indexes)
	d.dropLegacyIndexes()

	if err := d.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("database migrations completed successfully")
	return nil
}

// dropLegacyIndexes removes old indexes that have been replaced.
// Errors are logged but not fatal (index may not exist on new databases).
func (d *Database) dropLegacyIndexes() {
	migrator := d.db.Migrator()

	// idx_unique_timestamp replaced by idx_unique_factory_ts (dedup key moved to factory_timestamp)
	legacyIndexes := []struct {
		table string
		index string
	}{
		{"glucose_measurements", "idx_unique_timestamp"},
		{"glucose_measurements", "idx_factory_ts"},
	}

	for _, li := range legacyIndexes {
		if migrator.HasIndex(li.table, li.index) {
			if err := migrator.DropIndex(li.table, li.index); err != nil {
				slog.Warn("failed to drop legacy index", "table", li.table, "index", li.index, "error", err)
			} else {
				slog.Info("dropped legacy index", "table", li.table, "index", li.index)
			}
		}
	}
}

// DB returns the underlying GORM database instance.
func (d *Database) DB() *gorm.DB {
	return d.db
}

// Close closes the database connection.
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB for closing: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	slog.Info("database connection closed")
	return nil
}

// Ping checks if the database connection is alive.
func (d *Database) Ping(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB for ping: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Stats returns the database connection pool statistics.
func (d *Database) Stats() (sql.DBStats, error) {
	sqlDB, err := d.db.DB()
	if err != nil {
		return sql.DBStats{}, fmt.Errorf("failed to get underlying sql.DB for stats: %w", err)
	}

	return sqlDB.Stats(), nil
}

// parseLogLevel converts a string log level to GORM's logger.LogLevel.
func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Warn // Default to warn
	}
}
