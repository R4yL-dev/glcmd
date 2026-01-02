package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/domain"
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
	email := os.Getenv("GL_EMAIL")
	if email == "" {
		slog.Error("GL_EMAIL environment variable is not set")
		os.Exit(1)
	}

	password := os.Getenv("GL_PASSWORD")
	if password == "" {
		slog.Error("GL_PASSWORD environment variable is not set")
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

	// Create daemon with 5-minute interval
	interval := 5 * time.Minute
	d, err := daemon.New(glucoseService, sensorService, configService, interval, email, password)
	if err != nil {
		slog.Error("failed to create daemon", "error", err)
		os.Exit(1)
	}

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
		d.Stop()
		// Wait for daemon to finish
		if err := <-errChan; err != nil {
			slog.Error("daemon stopped with error", "error", err)
			os.Exit(1)
		}
	case err := <-errChan:
		if err != nil {
			slog.Error("daemon stopped with error", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("glcmd stopped successfully")
}
