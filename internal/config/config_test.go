package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Success(t *testing.T) {
	// Set required environment variables
	os.Setenv("GLCMD_EMAIL", "test@example.com")
	os.Setenv("GLCMD_PASSWORD", "testpassword")
	defer func() {
		os.Unsetenv("GLCMD_EMAIL")
		os.Unsetenv("GLCMD_PASSWORD")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify daemon defaults
	if cfg.Daemon.FetchInterval != 5*time.Minute {
		t.Errorf("expected FetchInterval 5m, got %v", cfg.Daemon.FetchInterval)
	}
	if cfg.Daemon.DisplayInterval != 1*time.Minute {
		t.Errorf("expected DisplayInterval 1m, got %v", cfg.Daemon.DisplayInterval)
	}
	if !cfg.Daemon.EnableEmojis {
		t.Errorf("expected EnableEmojis true, got false")
	}

	// Verify database defaults (SQLite)
	if cfg.Database.Type != "sqlite" {
		t.Errorf("expected database type sqlite, got %s", cfg.Database.Type)
	}
	if cfg.Database.SQLitePath != "./data/glcmd.db" {
		t.Errorf("expected SQLite path ./data/glcmd.db, got %s", cfg.Database.SQLitePath)
	}

	// Verify API defaults
	if cfg.API.Port != 8080 {
		t.Errorf("expected API port 8080, got %d", cfg.API.Port)
	}

	// Verify credentials
	if cfg.Credentials.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", cfg.Credentials.Email)
	}
	if cfg.Credentials.Password != "testpassword" {
		t.Errorf("expected password testpassword, got %s", cfg.Credentials.Password)
	}
}

func TestLoad_MissingEmail(t *testing.T) {
	// Unset email
	os.Unsetenv("GLCMD_EMAIL")
	os.Setenv("GLCMD_PASSWORD", "testpassword")
	defer os.Unsetenv("GLCMD_PASSWORD")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing GLCMD_EMAIL, got nil")
	}
}

func TestLoad_MissingPassword(t *testing.T) {
	// Unset password
	os.Setenv("GLCMD_EMAIL", "test@example.com")
	os.Unsetenv("GLCMD_PASSWORD")
	defer os.Unsetenv("GLCMD_EMAIL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing GLCMD_PASSWORD, got nil")
	}
}

func TestLoad_InvalidAPIPort(t *testing.T) {
	os.Setenv("GLCMD_EMAIL", "test@example.com")
	os.Setenv("GLCMD_PASSWORD", "testpassword")
	os.Setenv("GLCMD_API_PORT", "invalid")
	defer func() {
		os.Unsetenv("GLCMD_EMAIL")
		os.Unsetenv("GLCMD_PASSWORD")
		os.Unsetenv("GLCMD_API_PORT")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid GLCMD_API_PORT, got nil")
	}
}

func TestLoad_APIPortOutOfRange(t *testing.T) {
	os.Setenv("GLCMD_EMAIL", "test@example.com")
	os.Setenv("GLCMD_PASSWORD", "testpassword")
	os.Setenv("GLCMD_API_PORT", "99999")
	defer func() {
		os.Unsetenv("GLCMD_EMAIL")
		os.Unsetenv("GLCMD_PASSWORD")
		os.Unsetenv("GLCMD_API_PORT")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for out-of-range GLCMD_API_PORT, got nil")
	}
}

func TestLoad_PostgreSQLMissingPassword(t *testing.T) {
	os.Setenv("GLCMD_EMAIL", "test@example.com")
	os.Setenv("GLCMD_PASSWORD", "testpassword")
	os.Setenv("GLCMD_DB_TYPE", "postgres")
	// Don't set GLCMD_DB_PASSWORD
	defer func() {
		os.Unsetenv("GLCMD_EMAIL")
		os.Unsetenv("GLCMD_PASSWORD")
		os.Unsetenv("GLCMD_DB_TYPE")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for PostgreSQL without password, got nil")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("GLCMD_EMAIL", "custom@example.com")
	os.Setenv("GLCMD_PASSWORD", "custompassword")
	os.Setenv("GLCMD_API_PORT", "9090")
	os.Setenv("GLCMD_FETCH_INTERVAL", "10m")
	os.Setenv("GLCMD_DISPLAY_INTERVAL", "2m")
	os.Setenv("GLCMD_ENABLE_EMOJIS", "false")
	os.Setenv("GLCMD_DB_PATH", "/custom/path/db.sqlite")
	defer func() {
		os.Unsetenv("GLCMD_EMAIL")
		os.Unsetenv("GLCMD_PASSWORD")
		os.Unsetenv("GLCMD_API_PORT")
		os.Unsetenv("GLCMD_FETCH_INTERVAL")
		os.Unsetenv("GLCMD_DISPLAY_INTERVAL")
		os.Unsetenv("GLCMD_ENABLE_EMOJIS")
		os.Unsetenv("GLCMD_DB_PATH")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify custom values
	if cfg.Credentials.Email != "custom@example.com" {
		t.Errorf("expected email custom@example.com, got %s", cfg.Credentials.Email)
	}
	if cfg.API.Port != 9090 {
		t.Errorf("expected API port 9090, got %d", cfg.API.Port)
	}
	if cfg.Daemon.FetchInterval != 10*time.Minute {
		t.Errorf("expected FetchInterval 10m, got %v", cfg.Daemon.FetchInterval)
	}
	if cfg.Daemon.DisplayInterval != 2*time.Minute {
		t.Errorf("expected DisplayInterval 2m, got %v", cfg.Daemon.DisplayInterval)
	}
	if cfg.Daemon.EnableEmojis {
		t.Errorf("expected EnableEmojis false, got true")
	}
	if cfg.Database.SQLitePath != "/custom/path/db.sqlite" {
		t.Errorf("expected SQLite path /custom/path/db.sqlite, got %s", cfg.Database.SQLitePath)
	}
}

func TestToPersistenceConfig(t *testing.T) {
	dbCfg := DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: "./test.db",
		LogLevel:   "warn",
	}

	persistenceCfg := dbCfg.ToPersistenceConfig()

	if persistenceCfg.Type != dbCfg.Type {
		t.Errorf("expected Type %s, got %s", dbCfg.Type, persistenceCfg.Type)
	}
	if persistenceCfg.SQLitePath != dbCfg.SQLitePath {
		t.Errorf("expected SQLitePath %s, got %s", dbCfg.SQLitePath, persistenceCfg.SQLitePath)
	}
}

func TestToDaemonConfig(t *testing.T) {
	daemonCfg := DaemonConfig{
		FetchInterval:   5 * time.Minute,
		DisplayInterval: 1 * time.Minute,
		EnableEmojis:    true,
	}

	cfg := daemonCfg.ToDaemonConfig()

	if cfg.FetchInterval != daemonCfg.FetchInterval {
		t.Errorf("expected FetchInterval %v, got %v", daemonCfg.FetchInterval, cfg.FetchInterval)
	}
	if cfg.DisplayInterval != daemonCfg.DisplayInterval {
		t.Errorf("expected DisplayInterval %v, got %v", daemonCfg.DisplayInterval, cfg.DisplayInterval)
	}
	if cfg.EnableEmojis != daemonCfg.EnableEmojis {
		t.Errorf("expected EnableEmojis %v, got %v", daemonCfg.EnableEmojis, cfg.EnableEmojis)
	}
}
