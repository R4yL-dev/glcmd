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

	daemon := New(storage, interval)

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

	// Ticker should be nil until Run() is called
	if daemon.ticker != nil {
		t.Error("expected ticker to be nil before Run()")
	}
}

// TestStop_CancelsContext tests that Stop() cancels the daemon's context
func TestStop_CancelsContext(t *testing.T) {
	storage := memory.New()
	daemon := New(storage, 5*time.Minute)

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
	daemon := New(storage, 5*time.Minute)

	// Manually create a ticker to simulate Run() behavior
	daemon.ticker = time.NewTicker(1 * time.Second)

	// Stop should not panic even with a ticker
	daemon.Stop()

	// Ticker should be stopped (we can't directly verify, but no panic = success)
}

// TestStop_WithoutTicker tests that Stop() works even if ticker is nil
func TestStop_WithoutTicker(t *testing.T) {
	storage := memory.New()
	daemon := New(storage, 5*time.Minute)

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

// TestRun_ReturnsWithoutError tests that Run() can be called (skeleton implementation)
func TestRun_ReturnsWithoutError(t *testing.T) {
	storage := memory.New()
	daemon := New(storage, 5*time.Minute)

	// Current skeleton implementation should return immediately
	err := daemon.Run()
	if err != nil {
		t.Errorf("expected no error from skeleton Run(), got %v", err)
	}
}

// TestFetch_ReturnsWithoutError tests that fetch() can be called (skeleton implementation)
func TestFetch_ReturnsWithoutError(t *testing.T) {
	storage := memory.New()
	daemon := New(storage, 5*time.Minute)

	// Current skeleton implementation should return immediately
	err := daemon.fetch()
	if err != nil {
		t.Errorf("expected no error from skeleton fetch(), got %v", err)
	}
}
