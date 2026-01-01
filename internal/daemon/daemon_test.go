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

	if daemon.email != email {
		t.Errorf("expected email = %s, got %s", email, daemon.email)
	}

	if daemon.password != password {
		t.Errorf("expected password = %s, got %s", password, daemon.password)
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

// TestRun_GracefulShutdown tests that Run() stops gracefully when Stop() is called
func TestRun_GracefulShutdown(t *testing.T) {
	storage := memory.New()
	daemon, err := New(storage, 100*time.Millisecond, "test@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Run in goroutine (it will fail at authenticate, but that's ok for this test)
	done := make(chan error, 1)
	go func() {
		done <- daemon.Run()
	}()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Call Stop()
	daemon.Stop()

	// Run() should return quickly
	select {
	case err := <-done:
		// It will return an auth error, but the important part is that it returned
		_ = err // We expect an error (auth will fail with fake credentials)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run() did not stop after Stop() was called")
	}
}

// TestStop_Idempotent tests that calling Stop() multiple times is safe
func TestStop_Idempotent(t *testing.T) {
	storage := memory.New()
	daemon, err := New(storage, 5*time.Minute, "test@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Call Stop multiple times - should not panic
	daemon.Stop()
	daemon.Stop()
	daemon.Stop()

	// Context should be cancelled
	select {
	case <-daemon.ctx.Done():
		// Expected
	default:
		t.Fatal("context should be cancelled")
	}
}
