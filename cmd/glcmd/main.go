package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/R4yL-dev/glcmd/internal/api"
	"github.com/R4yL-dev/glcmd/internal/config"
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

	// Load centralized configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Database setup
	dbConfig := cfg.Database.ToPersistenceConfig()
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

	// Convert daemon config
	daemonConfig := cfg.Daemon.ToDaemonConfig()

	slog.Info("daemon configuration loaded",
		"fetchInterval", daemonConfig.FetchInterval,
		"displayInterval", daemonConfig.DisplayInterval,
		"emojisEnabled", daemonConfig.EnableEmojis,
	)

	// Create daemon
	d, err := daemon.New(glucoseService, sensorService, configService, daemonConfig, cfg.Credentials.Email, cfg.Credentials.Password)
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
		func() daemon.HealthStatus {
			return d.GetHealthStatus()
		},
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			return database.Ping(ctx) == nil
		},
		slog.Default(),
	)

	if err := apiServer.Start(); err != nil {
		slog.Error("failed to start unified API server", "error", err)
		os.Exit(1)
	}
	slog.Info("unified API server started", "port", cfg.API.Port)

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

	slog.Info("glcmd stopped successfully")
}
