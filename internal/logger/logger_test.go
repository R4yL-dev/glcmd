package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetup_CreatesLogDirectory(t *testing.T) {
	// Use temp directory for testing
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "logs", "test.log")

	file, err := Setup(logFile, slog.LevelInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()

	// Verify log directory was created
	logDir := filepath.Dir(logFile)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("log directory %s was not created", logDir)
	}

	// Verify log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log file %s was not created", logFile)
	}
}

func TestSetup_WritesToFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	file, err := Setup(logFile, slog.LevelInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()

	// Write a log message
	testMessage := "test log message"
	slog.Info(testMessage)

	// Close the file to flush
	file.Close()

	// Read the log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Verify message was written
	if !strings.Contains(string(content), testMessage) {
		t.Errorf("log file does not contain expected message %q, got: %s", testMessage, content)
	}
}

func TestSetup_RespectsLogLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Setup with INFO level (DEBUG should be filtered)
	file, err := Setup(logFile, slog.LevelInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()

	// Write DEBUG and INFO messages
	debugMsg := "debug message"
	infoMsg := "info message"

	slog.Debug(debugMsg)
	slog.Info(infoMsg)

	file.Close()

	// Read log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(content)

	// DEBUG should NOT be in the file (level is INFO)
	if strings.Contains(contentStr, debugMsg) {
		t.Errorf("log file contains DEBUG message when level is INFO")
	}

	// INFO should be in the file
	if !strings.Contains(contentStr, infoMsg) {
		t.Errorf("log file does not contain INFO message")
	}
}

func TestNewTestLogger_WorksWithoutFiles(t *testing.T) {
	logger := NewTestLogger(slog.LevelInfo)
	if logger == nil {
		t.Fatal("expected logger to be non-nil")
	}

	// Should not panic when logging
	logger.Info("test message")
	logger.Error("test error")
}

func TestSetup_AppendsToExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// First setup and write
	file1, err := Setup(logFile, slog.LevelInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slog.Info("first message")
	file1.Close()

	// Second setup (should append, not truncate)
	file2, err := Setup(logFile, slog.LevelInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slog.Info("second message")
	file2.Close()

	// Read file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(content)

	// Both messages should be present
	if !strings.Contains(contentStr, "first message") {
		t.Error("first message not found in log file")
	}
	if !strings.Contains(contentStr, "second message") {
		t.Error("second message not found in log file")
	}
}
