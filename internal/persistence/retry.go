package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// RetryConfig holds configuration for retry logic with exponential backoff.
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retry attempts
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	Multiplier     float64       // Backoff multiplier for exponential backoff
}

// DefaultRetryConfig returns default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
	}
}

// ExecuteWithRetry executes a function with retry logic and exponential backoff.
// Only retries if the error is retryable (determined by IsRetryable function).
func ExecuteWithRetry(ctx context.Context, config *RetryConfig, fn func() error) error {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute function
		lastErr = fn()

		// Success - return immediately
		if lastErr == nil {
			if attempt > 0 {
				slog.Debug("operation succeeded after retry",
					"attempt", attempt,
					"totalAttempts", attempt+1,
				)
			}
			return nil
		}

		// Check if error is retryable
		if !IsRetryable(lastErr) {
			slog.Debug("error is not retryable, failing immediately",
				"error", lastErr,
			)
			return lastErr
		}

		// Last attempt - don't sleep, just return the error
		if attempt == config.MaxRetries {
			break
		}

		// Log retry attempt
		slog.Warn("retrying operation after error",
			"attempt", attempt+1,
			"maxRetries", config.MaxRetries,
			"backoff", backoff,
			"error", lastErr,
		)

		// Sleep with backoff (respect context cancellation)
		select {
		case <-time.After(backoff):
			// Continue to next retry
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled by context: %w", ctx.Err())
		}

		// Increase backoff exponentially
		backoff = time.Duration(float64(backoff) * config.Multiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}
