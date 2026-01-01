package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/logger"
	"github.com/R4yL-dev/glcmd/internal/storage/memory"
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

	// Create in-memory storage
	storage := memory.New()
	slog.Info("storage initialized", "type", "memory")

	// Create daemon with 5-minute interval
	interval := 5 * time.Minute
	d, err := daemon.New(storage, interval, email, password)
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
