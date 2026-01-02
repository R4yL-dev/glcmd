package daemon

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.FetchInterval != 5*time.Minute {
		t.Errorf("expected FetchInterval = 5m, got %v", config.FetchInterval)
	}

	if config.DisplayInterval != 1*time.Minute {
		t.Errorf("expected DisplayInterval = 1m, got %v", config.DisplayInterval)
	}

	if config.EnableEmojis != true {
		t.Error("expected EnableEmojis = true")
	}
}

func TestLoadConfigFromEnv_Defaults(t *testing.T) {
	// Clear all relevant env vars
	os.Unsetenv("GLCMD_FETCH_INTERVAL")
	os.Unsetenv("GLCMD_DISPLAY_INTERVAL")
	os.Unsetenv("GLCMD_ENABLE_EMOJIS")

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use defaults
	if config.FetchInterval != 5*time.Minute {
		t.Errorf("expected default FetchInterval = 5m, got %v", config.FetchInterval)
	}

	if config.DisplayInterval != 1*time.Minute {
		t.Errorf("expected default DisplayInterval = 1m, got %v", config.DisplayInterval)
	}

	if config.EnableEmojis != true {
		t.Error("expected default EnableEmojis = true")
	}
}

func TestLoadConfigFromEnv_CustomValues(t *testing.T) {
	// Set custom env vars
	os.Setenv("GLCMD_FETCH_INTERVAL", "10m")
	os.Setenv("GLCMD_DISPLAY_INTERVAL", "30s")
	os.Setenv("GLCMD_ENABLE_EMOJIS", "false")
	defer func() {
		os.Unsetenv("GLCMD_FETCH_INTERVAL")
		os.Unsetenv("GLCMD_DISPLAY_INTERVAL")
		os.Unsetenv("GLCMD_ENABLE_EMOJIS")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.FetchInterval != 10*time.Minute {
		t.Errorf("expected FetchInterval = 10m, got %v", config.FetchInterval)
	}

	if config.DisplayInterval != 30*time.Second {
		t.Errorf("expected DisplayInterval = 30s, got %v", config.DisplayInterval)
	}

	if config.EnableEmojis != false {
		t.Error("expected EnableEmojis = false")
	}
}

func TestLoadConfigFromEnv_InvalidFetchInterval(t *testing.T) {
	os.Setenv("GLCMD_FETCH_INTERVAL", "invalid")
	defer os.Unsetenv("GLCMD_FETCH_INTERVAL")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid GLCMD_FETCH_INTERVAL, got nil")
	}

	// Error should mention the variable name
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestLoadConfigFromEnv_InvalidDisplayInterval(t *testing.T) {
	os.Setenv("GLCMD_DISPLAY_INTERVAL", "not-a-duration")
	defer os.Unsetenv("GLCMD_DISPLAY_INTERVAL")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid GLCMD_DISPLAY_INTERVAL, got nil")
	}
}

func TestLoadConfigFromEnv_InvalidEmojisFlag(t *testing.T) {
	os.Setenv("GLCMD_ENABLE_EMOJIS", "maybe")
	defer os.Unsetenv("GLCMD_ENABLE_EMOJIS")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid GLCMD_ENABLE_EMOJIS, got nil")
	}
}

func TestParseDuration_ValidFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"1h", 1 * time.Hour},
		{"90s", 90 * time.Second},
		{"1h30m", 90 * time.Minute},
		{"500ms", 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			duration, err := parseDuration(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if duration != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, duration)
			}
		})
	}
}

func TestParseDuration_InvalidFormats(t *testing.T) {
	tests := []string{
		"invalid",
		"5",         // Missing unit
		"five",      // Not a number
		"",          // Empty string
		"-5m",       // Negative (should fail validation)
		"0s",        // Zero (should fail validation)
		"abc123",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDuration(input)
			if err == nil {
				t.Errorf("expected error for %q, got nil", input)
			}
		})
	}
}

func TestParseDuration_NegativeDuration(t *testing.T) {
	duration, err := parseDuration("-5m")
	if err == nil {
		t.Fatal("expected error for negative duration, got nil")
	}
	if duration != 0 {
		t.Errorf("expected duration = 0 on error, got %v", duration)
	}
}

func TestParseDuration_ZeroDuration(t *testing.T) {
	duration, err := parseDuration("0s")
	if err == nil {
		t.Fatal("expected error for zero duration, got nil")
	}
	if duration != 0 {
		t.Errorf("expected duration = 0 on error, got %v", duration)
	}
}

func TestLoadConfigFromEnv_ComplexScenario(t *testing.T) {
	// Test with realistic production values
	os.Setenv("GLCMD_FETCH_INTERVAL", "3m")
	os.Setenv("GLCMD_DISPLAY_INTERVAL", "2m")
	os.Setenv("GLCMD_ENABLE_EMOJIS", "true")
	defer func() {
		os.Unsetenv("GLCMD_FETCH_INTERVAL")
		os.Unsetenv("GLCMD_DISPLAY_INTERVAL")
		os.Unsetenv("GLCMD_ENABLE_EMOJIS")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all values are correctly parsed
	if config.FetchInterval != 3*time.Minute {
		t.Errorf("FetchInterval: expected 3m, got %v", config.FetchInterval)
	}

	if config.DisplayInterval != 2*time.Minute {
		t.Errorf("DisplayInterval: expected 2m, got %v", config.DisplayInterval)
	}

	if config.EnableEmojis != true {
		t.Error("EnableEmojis: expected true")
	}
}
