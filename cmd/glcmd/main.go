package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/healthcheck"
	"github.com/R4yL-dev/glcmd/internal/logger"
	"github.com/R4yL-dev/glcmd/internal/persistence"
	"github.com/R4yL-dev/glcmd/internal/repository"
	"github.com/R4yL-dev/glcmd/internal/service"
)

func main() {
	// Setup logger
	logFile, err := logger.Setup(logger.DefaultLogFile, slog.LevelInfo)
	if err != nil {
		slog.Error("failed to setup logger", "error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	slog.Info("glcmd starting")

	// Get credentials from environment
	email := os.Getenv("GLCMD_EMAIL")
	if email == "" {
		slog.Error("GLCMD_EMAIL environment variable is not set")
		os.Exit(1)
	}

	password := os.Getenv("GLCMD_PASSWORD")
	if password == "" {
		slog.Error("GLCMD_PASSWORD environment variable is not set")
		os.Exit(1)
	}

	// Database setup
	dbConfig := persistence.LoadDatabaseConfigFromEnv()
	database, err := persistence.NewDatabase(dbConfig)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

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

	slog.Info("database connected successfully",
		"type", dbConfig.Type,
		"path", dbConfig.SQLitePath,
	)

	// Create repositories
	measurementRepo := repository.NewMeasurementRepository(database.DB())
	sensorRepo := repository.NewSensorRepository(database.DB())
	userRepo := repository.NewUserRepository(database.DB())
	deviceRepo := repository.NewDeviceRepository(database.DB())
	targetsRepo := repository.NewTargetsRepository(database.DB())

	// Create Unit of Work
	uow := repository.NewUnitOfWork(database.DB())

	// Create services
	glucoseService := service.NewGlucoseService(measurementRepo, slog.Default())
	sensorService := service.NewSensorService(sensorRepo, uow, slog.Default())
	configService := service.NewConfigService(userRepo, deviceRepo, targetsRepo, slog.Default())

	slog.Info("services initialized successfully")

	// Load daemon configuration
	daemonConfig, err := daemon.LoadConfigFromEnv()
	if err != nil {
		slog.Error("failed to load daemon configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("daemon configuration loaded",
		"fetchInterval", daemonConfig.FetchInterval,
		"displayInterval", daemonConfig.DisplayInterval,
		"emojisEnabled", daemonConfig.EnableEmojis,
	)

	// Create daemon
	d, err := daemon.New(glucoseService, sensorService, configService, daemonConfig, email, password)
	if err != nil {
		slog.Error("failed to create daemon", "error", err)
		os.Exit(1)
	}

	// Start healthcheck HTTP server (optional)
	healthcheckPort := 8080 // Default port
	if portStr := os.Getenv("GLCMD_HEALTHCHECK_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			healthcheckPort = port
		}
	}

	healthServer := healthcheck.NewServer(healthcheckPort, func() interface{} {
		return d.GetHealthStatus()
	})

	if err := healthServer.Start(); err != nil {
		slog.Error("failed to start healthcheck server", "error", err)
		os.Exit(1)
	}
	slog.Info("healthcheck server started", "port", healthcheckPort)

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
		slog.Info("received signal, shutting down", "signal", sig)

		// Stop daemon
		d.Stop()

		// Stop healthcheck server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := healthServer.Stop(ctx); err != nil {
			slog.Error("failed to stop healthcheck server", "error", err)
		}

		// Wait for daemon to finish
		if err := <-errChan; err != nil {
			slog.Error("daemon stopped with error", "error", err)
			os.Exit(1)
		}
	case err := <-errChan:
		if err != nil {
			slog.Error("daemon stopped with error", "error", err)

			// Stop healthcheck server
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			healthServer.Stop(ctx)

			os.Exit(1)
		}
	}

	slog.Info("glcmd stopped successfully")
}
