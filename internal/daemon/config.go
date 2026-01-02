package daemon

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds daemon configuration.
type Config struct {
	FetchInterval   time.Duration // Interval between API fetches (default: 5m)
	DisplayInterval time.Duration // Interval between measurement displays (default: 1m)
	EnableEmojis    bool          // Enable emoji display (default: true)
}

// DefaultConfig returns default daemon configuration.
func DefaultConfig() *Config {
	return &Config{
		FetchInterval:   5 * time.Minute,
		DisplayInterval: 1 * time.Minute,
		EnableEmojis:    true,
	}
}

// LoadConfigFromEnv loads daemon configuration from environment variables.
// Falls back to defaults if env vars are not set.
func LoadConfigFromEnv() (*Config, error) {
	config := DefaultConfig()

	// Fetch interval
	if fetchInterval := os.Getenv("GLCMD_FETCH_INTERVAL"); fetchInterval != "" {
		duration, err := parseDuration(fetchInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid GLCMD_FETCH_INTERVAL: %w", err)
		}
		config.FetchInterval = duration
	}

	// Display interval
	if displayInterval := os.Getenv("GLCMD_DISPLAY_INTERVAL"); displayInterval != "" {
		duration, err := parseDuration(displayInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid GLCMD_DISPLAY_INTERVAL: %w", err)
		}
		config.DisplayInterval = duration
	}

	// Enable emojis
	if enableEmojis := os.Getenv("GLCMD_ENABLE_EMOJIS"); enableEmojis != "" {
		enabled, err := strconv.ParseBool(enableEmojis)
		if err != nil {
			return nil, fmt.Errorf("invalid GLCMD_ENABLE_EMOJIS (use true/false): %w", err)
		}
		config.EnableEmojis = enabled
	}

	return config, nil
}

// parseDuration parses a duration string.
// Supports Go duration format (e.g., "5m", "1h30m", "90s").
func parseDuration(s string) (time.Duration, error) {
	duration, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration %q: %w (use format like 5m, 1h, 90s)", s, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive, got %v", duration)
	}
	return duration, nil
}
