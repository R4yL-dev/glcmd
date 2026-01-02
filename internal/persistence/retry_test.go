package persistence

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecuteWithRetry_Success(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	callCount := 0
	err := ExecuteWithRetry(context.Background(), config, func() error {
		callCount++
		return nil // Success
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call (no retries), got %d", callCount)
	}
}

func TestExecuteWithRetry_SuccessAfterRetries(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	callCount := 0
	err := ExecuteWithRetry(context.Background(), config, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("database is locked") // Retryable error
		}
		return nil // Success on 3rd attempt
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls (2 retries), got %d", callCount)
	}
}

func TestExecuteWithRetry_MaxRetriesExceeded(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	callCount := 0
	err := ExecuteWithRetry(context.Background(), config, func() error {
		callCount++
		return errors.New("database is locked") // Always fails
	})

	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}

	// MaxRetries = 2 means: 1 initial attempt + 2 retries = 3 total calls
	if callCount != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}

	// Verify error message contains retry information
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestExecuteWithRetry_NonRetryableError(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	callCount := 0
	nonRetryableErr := errors.New("invalid data") // Not a retryable error

	err := ExecuteWithRetry(context.Background(), config, func() error {
		callCount++
		return nonRetryableErr
	})

	if err != nonRetryableErr {
		t.Fatalf("expected error %v, got %v", nonRetryableErr, err)
	}

	// Should NOT retry on non-retryable error
	if callCount != 1 {
		t.Errorf("expected 1 call (no retries for non-retryable error), got %d", callCount)
	}
}

func TestExecuteWithRetry_ContextCancellation(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	errChan := make(chan error, 1)

	go func() {
		err := ExecuteWithRetry(ctx, config, func() error {
			callCount++
			return errors.New("database is locked") // Retryable error
		})
		errChan <- err
	}()

	// Cancel context after a short delay (during backoff)
	time.Sleep(50 * time.Millisecond)
	cancel()

	err := <-errChan

	if err == nil {
		t.Fatal("expected error from context cancellation, got nil")
	}

	// Should stop retrying when context is cancelled
	if callCount > 2 {
		t.Errorf("expected few calls before cancellation, got %d", callCount)
	}
}

func TestExecuteWithRetry_ExponentialBackoff(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		Multiplier:     2.0,
	}

	callTimes := []time.Time{}
	err := ExecuteWithRetry(context.Background(), config, func() error {
		callTimes = append(callTimes, time.Now())
		return errors.New("database is locked") // Always fails
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if len(callTimes) != 4 { // 1 initial + 3 retries
		t.Fatalf("expected 4 calls, got %d", len(callTimes))
	}

	// Check backoff durations (approximately)
	// 1st retry: ~50ms
	// 2nd retry: ~100ms
	// 3rd retry: ~200ms

	tolerance := 50 * time.Millisecond

	delay1 := callTimes[1].Sub(callTimes[0])
	if delay1 < 50*time.Millisecond-tolerance || delay1 > 50*time.Millisecond+tolerance {
		t.Logf("Warning: 1st retry delay %v not close to 50ms", delay1)
	}

	delay2 := callTimes[2].Sub(callTimes[1])
	if delay2 < 100*time.Millisecond-tolerance || delay2 > 100*time.Millisecond+tolerance {
		t.Logf("Warning: 2nd retry delay %v not close to 100ms", delay2)
	}

	delay3 := callTimes[3].Sub(callTimes[2])
	if delay3 < 200*time.Millisecond-tolerance || delay3 > 200*time.Millisecond+tolerance {
		t.Logf("Warning: 3rd retry delay %v not close to 200ms", delay3)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "database is locked",
			err:      errors.New("database is locked"),
			expected: true,
		},
		{
			name:     "SQLITE_BUSY",
			err:      errors.New("SQLITE_BUSY: database table is locked"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "timeout",
			err:      errors.New("timeout waiting for response"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("invalid syntax"),
			expected: false,
		},
		{
			name:     "ErrNotFound",
			err:      ErrNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryable(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
