// Package daemon implements the continuous background process that fetches
// glucose data from the LibreView API at regular intervals.
//
// The daemon runs a main loop that:
//   - Fetches data every 5 minutes using a ticker
//   - Stores all received data in the configured storage backend
//   - Handles graceful shutdown via context cancellation
//   - Logs all operations and errors
//
// Usage:
//
//	storage := memory.New()
//	d := daemon.New(storage, 5*time.Minute)
//	if err := d.Run(); err != nil {
//	    log.Fatal(err)
//	}
package daemon

import (
	"context"
	"time"

	"github.com/R4yL-dev/glcmd/internal/storage"
)

// Daemon represents the background service that continuously fetches
// glucose data from the LibreView API.
//
// It manages:
//   - A ticker for periodic fetching (default 5 minutes)
//   - Context-based lifecycle management for graceful shutdown
//   - Storage backend for persisting fetched data
type Daemon struct {
	storage  storage.Storage
	ctx      context.Context
	cancel   context.CancelFunc
	ticker   *time.Ticker
	interval time.Duration
}

// New creates a new Daemon instance.
//
// Parameters:
//   - storage: The storage backend for persisting data
//   - interval: The time between fetch operations (e.g., 5*time.Minute)
//
// The daemon is created with a background context that can be cancelled
// via the Stop() method for graceful shutdown.
func New(storage storage.Storage, interval time.Duration) *Daemon {
	ctx, cancel := context.WithCancel(context.Background())

	return &Daemon{
		storage:  storage,
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
	}
}

// Run starts the daemon's main loop.
//
// This method blocks until the daemon is stopped via Stop() or an
// unrecoverable error occurs.
//
// The main loop:
//   - Performs an initial fetch to populate historical data
//   - Starts a ticker for periodic fetches at the configured interval
//   - Waits for context cancellation to stop gracefully
//
// Returns an error if the daemon cannot start or encounters a fatal error.
func (d *Daemon) Run() error {
	// TODO (Step 5): Implement initial fetch (historical data)
	// TODO (Step 6): Implement ticker loop and graceful shutdown
	return nil
}

// Stop initiates a graceful shutdown of the daemon.
//
// This method:
//   - Cancels the daemon's context
//   - Stops the ticker if running
//   - Allows in-progress operations to complete
//
// After calling Stop(), the Run() method will return.
func (d *Daemon) Stop() {
	if d.ticker != nil {
		d.ticker.Stop()
	}
	d.cancel()
}

// fetch retrieves the latest glucose data from the LibreView API
// and stores it in the configured storage backend.
//
// This method:
//   - Authenticates with the LibreView API
//   - Fetches current glucose measurements from /connections
//   - Updates sensor configuration if changed
//   - Stores all received data
//   - Logs errors without stopping the daemon
//
// Returns an error if the fetch operation fails. Network errors are
// logged but do not stop the daemon.
func (d *Daemon) fetch() error {
	// TODO (Step 5): Implement actual fetching logic
	// - Authenticate (or reuse token)
	// - Fetch from /connections
	// - Parse and store data
	// - Detect sensor changes
	// - Log operations
	return nil
}
