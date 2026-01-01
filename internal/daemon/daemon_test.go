package daemon

import (
	"testing"
	"time"

	"github.com/R4yL-dev/glcmd/internal/storage/memory"
)

// TestNew_Initialization tests that New() creates a properly initialized Daemon
func TestNew_Initialization(t *testing.T) {
	storage := memory.New()
	interval := 5 * time.Minute
	email := "test@example.com"
	password := "password123"

	daemon, err := New(storage, interval, email, password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if daemon == nil {
		t.Fatal("expected daemon to be non-nil")
	}

	if daemon.storage == nil {
		t.Error("expected storage to be set")
	}

	if daemon.ctx == nil {
		t.Error("expected context to be set")
	}

	if daemon.cancel == nil {
		t.Error("expected cancel function to be set")
	}

	if daemon.interval != interval {
		t.Errorf("expected interval = %v, got %v", interval, daemon.interval)
	}

	if daemon.client == nil {
		t.Error("expected HTTP client to be set")
	}

	if daemon.headers == nil {
		t.Error("expected headers to be set")
	}

	if daemon.creds == nil {
		t.Error("expected credentials to be set")
	}

	// Ticker should be nil until Run() is called
	if daemon.ticker != nil {
		t.Error("expected ticker to be nil before Run()")
	}
}

// TestNew_InvalidCredentials tests that New() returns error for invalid credentials
func TestNew_InvalidCredentials(t *testing.T) {
	storage := memory.New()
	interval := 5 * time.Minute

	// Empty email should fail
	_, err := New(storage, interval, "", "password")
	if err == nil {
		t.Error("expected error for empty email")
	}

	// Empty password should fail
	_, err = New(storage, interval, "test@example.com", "")
	if err == nil {
		t.Error("expected error for empty password")
	}
}

// TestStop_CancelsContext tests that Stop() cancels the daemon's context
func TestStop_CancelsContext(t *testing.T) {
	storage := memory.New()
	daemon, err := New(storage, 5*time.Minute, "test@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Context should not be cancelled initially
	select {
	case <-daemon.ctx.Done():
		t.Fatal("context should not be cancelled before Stop()")
	default:
		// Expected: context is still active
	}

	// Call Stop()
	daemon.Stop()

	// Context should now be cancelled
	select {
	case <-daemon.ctx.Done():
		// Expected: context is cancelled
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should be cancelled after Stop()")
	}
}

// TestStop_StopsTicker tests that Stop() stops the ticker if it exists
func TestStop_StopsTicker(t *testing.T) {
	storage := memory.New()
	daemon, err := New(storage, 5*time.Minute, "test@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Manually create a ticker to simulate Run() behavior
	daemon.ticker = time.NewTicker(1 * time.Second)

	// Stop should not panic even with a ticker
	daemon.Stop()

	// Ticker should be stopped (we can't directly verify, but no panic = success)
}

// TestStop_WithoutTicker tests that Stop() works even if ticker is nil
func TestStop_WithoutTicker(t *testing.T) {
	storage := memory.New()
	daemon, err := New(storage, 5*time.Minute, "test@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stop should not panic even without a ticker
	daemon.Stop()

	// Verify context is cancelled
	select {
	case <-daemon.ctx.Done():
		// Expected
	default:
		t.Fatal("context should be cancelled")
	}
}
