package daemon

import (
	"fmt"
	"os"
	"time"
)

// Config holds daemon configuration.
type Config struct {
	FetchInterval time.Duration // Interval between API fetches (default: 5m)
}

// DefaultConfig returns default daemon configuration.
func DefaultConfig() *Config {
	return &Config{
		FetchInterval: 5 * time.Minute,
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
