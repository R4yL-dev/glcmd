package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/R4yL-dev/glcmd/internal/api"
	"github.com/R4yL-dev/glcmd/internal/config"
	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/events"
	"github.com/R4yL-dev/glcmd/internal/persistence"
	"github.com/R4yL-dev/glcmd/internal/repository"
	"github.com/R4yL-dev/glcmd/internal/service"
)

// getLogLevel returns the slog level from GLCMD_LOG_LEVEL env var.
// Defaults to INFO.
func getLogLevel() slog.Level {
	level := strings.ToLower(os.Getenv("GLCMD_LOG_LEVEL"))
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// setupLogger configures slog based on environment variables.
// GLCMD_LOG_FORMAT: "text" (default) or "json"
// GLCMD_LOG_LEVEL: "debug", "info" (default), "warn", "error"
func setupLogger() {
	opts := &slog.HandlerOptions{
		Level: getLogLevel(),
	}

	var handler slog.Handler
	if os.Getenv("GLCMD_LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func main() {
	// Setup logger
	setupLogger()

	slog.Info("glcore starting")

	// Load centralized configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Database setup
	dbStart := time.Now()
	dbConfig := cfg.Database.ToPersistenceConfig()
	database, err := persistence.NewDatabase(dbConfig)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		database.Close()
		slog.Info("database closed")
	}()

	// Run migrations
	if err := database.AutoMigrate(
		&domain.GlucoseMeasurement{},
		&domain.SensorConfig{},
		&domain.UserPreferences{},
		&domain.DeviceInfo{},
		&domain.GlucoseTargets{},
	); err != nil {
		slog.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}

	// Database health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.Ping(ctx); err != nil {
		slog.Error("database health check failed", "error", err)
		os.Exit(1)
	}

	slog.Info("database ready",
		"type", dbConfig.Type,
		"duration", time.Since(dbStart),
	)

	// Create repositories
	glucoseRepo := repository.NewGlucoseRepository(database.DB())
	sensorRepo := repository.NewSensorRepository(database.DB())
	userRepo := repository.NewUserRepository(database.DB())
	deviceRepo := repository.NewDeviceRepository(database.DB())
	targetsRepo := repository.NewTargetsRepository(database.DB())

	// Create Unit of Work
	uow := repository.NewUnitOfWork(database.DB())

	// Create event broker for SSE streaming
	eventBroker := events.NewBroker(10, slog.Default())
	eventBroker.Start()
	defer eventBroker.Stop()

	// Create services with event broker
	glucoseService := service.NewGlucoseService(glucoseRepo, slog.Default(), eventBroker)
	sensorService := service.NewSensorService(sensorRepo, uow, slog.Default(), eventBroker)
	configService := service.NewConfigService(userRepo, deviceRepo, targetsRepo, slog.Default())

	// Create daemon
	d, err := daemon.New(glucoseService, sensorService, configService, cfg.Credentials.Email, cfg.Credentials.Password)
	if err != nil {
		slog.Error("failed to create daemon", "error", err)
		os.Exit(1)
	}

	// Create unified API server with daemon health status callback
	apiServer := api.NewServer(
		cfg.API.Port,
		glucoseService,
		sensorService,
		configService,
		eventBroker,
		func() daemon.HealthStatus {
			return d.GetHealthStatus()
		},
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			return database.Ping(ctx) == nil
		},
		func() *api.DatabasePoolStats {
			stats, err := database.Stats()
			if err != nil {
				return nil
			}
			return &api.DatabasePoolStats{
				OpenConnections: stats.OpenConnections,
				InUse:           stats.InUse,
				Idle:            stats.Idle,
				WaitCount:       stats.WaitCount,
				WaitDuration:    stats.WaitDuration.String(),
			}
		},
		slog.Default(),
	)

	if err := apiServer.Start(); err != nil {
		slog.Error("failed to start API server", "error", err)
		os.Exit(1)
	}
	slog.Info("API server listening", "port", cfg.API.Port)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run daemon in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- d.Run()
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		slog.Info("shutting down", "signal", sig)

		// Stop daemon
		d.Stop()

		// Stop API server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := apiServer.Stop(ctx); err != nil {
			slog.Error("failed to stop API server", "error", err)
		}

		// Wait for daemon to finish
		if err := <-errChan; err != nil {
			slog.Error("daemon stopped with error", "error", err)
			os.Exit(1)
		}
	case err := <-errChan:
		if err != nil {
			slog.Error("daemon stopped with error", "error", err)

			// Stop API server
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			apiServer.Stop(ctx)

			os.Exit(1)
		}
	}

	slog.Info("glcore stopped")
}
