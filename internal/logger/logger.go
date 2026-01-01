// Package logger provides structured logging using slog.
//
// Features:
//   - Multi-output: stdout + file (./logs/glcmd.log)
//   - Structured JSON logs to file, human-readable to stdout
//   - Configurable log levels (DEBUG, INFO, ERROR)
//   - Automatic log directory creation
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	// DefaultLogFile is the default log file path
	DefaultLogFile = "./logs/glcmd.log"
)

// Setup configures the global slog logger with multi-output (stdout + file).
//
// Parameters:
//   - logFile: Path to log file (use DefaultLogFile for default)
//   - level: Log level (slog.LevelDebug, slog.LevelInfo, slog.LevelError)
//
// Returns the log file handle (caller should defer Close()) and any error.
func Setup(logFile string, level slog.Level) (*os.File, error) {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Open log file (create or append)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logFile, err)
	}

	// Create multi-writer (stdout + file)
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Create handler with human-readable text format
	// For production, consider JSON format for better parsing
	handler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level: level,
		// Add source location for debugging (file:line)
		AddSource: level == slog.LevelDebug,
	})

	// Set as default logger
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return file, nil
}

// NewTestLogger creates a logger that writes only to stdout (for testing).
// Does not create any files.
func NewTestLogger(level slog.Level) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}
