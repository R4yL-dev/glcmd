package daemon

import (
	"context"
	"testing"
	"time"
)

func TestGetHealthStatus_Healthy(t *testing.T) {
	// Create daemon with mock services (nil is OK for health status test)
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		lastFetchError:       "",
		lastFetchTime:        time.Now(),
		startTime:            time.Now().Add(-10 * time.Minute),
	}

	status := d.GetHealthStatus()

	if status.Status != "healthy" {
		t.Errorf("expected status = healthy, got %s", status.Status)
	}

	if status.ConsecutiveErrors != 0 {
		t.Errorf("expected ConsecutiveErrors = 0, got %d", status.ConsecutiveErrors)
	}

	if status.LastFetchError != "" {
		t.Errorf("expected empty LastFetchError, got %s", status.LastFetchError)
	}

	if status.Uptime == "" {
		t.Error("expected non-empty Uptime")
	}

	if !status.DataFresh {
		t.Error("expected DataFresh = true")
	}
}

func TestGetHealthStatus_Degraded(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    3, // Between 0 and max
		maxConsecutiveErrors: 5,
		lastFetchError:       "network timeout",
		lastFetchTime:        time.Now().Add(-3 * time.Minute),
		startTime:            time.Now().Add(-1 * time.Hour),
	}

	status := d.GetHealthStatus()

	if status.Status != "degraded" {
		t.Errorf("expected status = degraded, got %s", status.Status)
	}

	if status.ConsecutiveErrors != 3 {
		t.Errorf("expected ConsecutiveErrors = 3, got %d", status.ConsecutiveErrors)
	}

	if status.LastFetchError != "network timeout" {
		t.Errorf("expected LastFetchError = 'network timeout', got %s", status.LastFetchError)
	}

	if status.DataFresh {
		t.Error("expected DataFresh = false (3m > 2x1m)")
	}
}

func TestGetHealthStatus_Unhealthy(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    5, // Equal to max
		maxConsecutiveErrors: 5,
		lastFetchError:       "authentication failed",
		lastFetchTime:        time.Now().Add(-30 * time.Minute),
		startTime:            time.Now().Add(-2 * time.Hour),
	}

	status := d.GetHealthStatus()

	if status.Status != "unhealthy" {
		t.Errorf("expected status = unhealthy, got %s", status.Status)
	}

	if status.ConsecutiveErrors != 5 {
		t.Errorf("expected ConsecutiveErrors = 5, got %d", status.ConsecutiveErrors)
	}

	if status.LastFetchError == "" {
		t.Error("expected non-empty LastFetchError")
	}

	if status.DataFresh {
		t.Error("expected DataFresh = false (30m > 2x1m)")
	}
}

func TestGetHealthStatus_UnhealthyAboveMax(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    10, // Above max
		maxConsecutiveErrors: 5,
		lastFetchError:       "persistent error",
		startTime:            time.Now().Add(-5 * time.Hour),
	}

	status := d.GetHealthStatus()

	if status.Status != "unhealthy" {
		t.Errorf("expected status = unhealthy, got %s", status.Status)
	}

	if status.ConsecutiveErrors != 10 {
		t.Errorf("expected ConsecutiveErrors = 10, got %d", status.ConsecutiveErrors)
	}
}

func TestGetHealthStatus_TimestampPresent(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		startTime:            time.Now(),
	}

	before := time.Now()
	status := d.GetHealthStatus()
	after := time.Now()

	// Timestamp should be between before and after
	if status.Timestamp.Before(before) || status.Timestamp.After(after) {
		t.Errorf("expected Timestamp between %v and %v, got %v", before, after, status.Timestamp)
	}
}

func TestGetHealthStatus_UptimeCalculation(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Hour)
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		startTime:            startTime,
	}

	status := d.GetHealthStatus()

	// Uptime should be approximately 1 hour
	if status.Uptime == "" {
		t.Fatal("expected non-empty Uptime")
	}

	// Parse uptime and verify it's reasonable
	// Should contain "h" for hour
	if len(status.Uptime) < 2 {
		t.Errorf("uptime seems too short: %s", status.Uptime)
	}
}

func TestGetHealthStatus_EdgeCaseBoundary(t *testing.T) {
	tests := []struct {
		name              string
		consecutiveErrors int
		maxErrors         int
		expectedStatus    string
	}{
		{"Zero errors", 0, 5, "healthy"},
		{"One error", 1, 5, "degraded"},
		{"Just below max", 4, 5, "degraded"},
		{"At max", 5, 5, "unhealthy"},
		{"Above max", 6, 5, "unhealthy"},
		{"High max threshold", 3, 10, "degraded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Daemon{
				ctx:                  context.Background(),
				consecutiveErrors:    tt.consecutiveErrors,
				maxConsecutiveErrors: tt.maxErrors,
				startTime:            time.Now(),
			}

			status := d.GetHealthStatus()

			if status.Status != tt.expectedStatus {
				t.Errorf("expected status = %s, got %s", tt.expectedStatus, status.Status)
			}

			// Zero lastFetchTime should be considered fresh
			if !status.DataFresh {
				t.Error("expected DataFresh = true for zero lastFetchTime")
			}
		})
	}
}

func TestGetHealthStatus_LastFetchTimePreserved(t *testing.T) {
	lastFetch := time.Now().Add(-90 * time.Second) // Within 2x1m = 2m
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		lastFetchTime:        lastFetch,
		startTime:            time.Now().Add(-1 * time.Hour),
	}

	status := d.GetHealthStatus()

	if !status.LastFetchTime.Equal(lastFetch) {
		t.Errorf("expected LastFetchTime = %v, got %v", lastFetch, status.LastFetchTime)
	}
}

func TestGetHealthStatus_DataFresh(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		lastFetchTime:        time.Now().Add(-90 * time.Second), // 90s < 2x1m = 2m
		startTime:            time.Now().Add(-1 * time.Hour),
	}

	status := d.GetHealthStatus()

	if status.Status != "healthy" {
		t.Errorf("expected status = healthy, got %s", status.Status)
	}

	if !status.DataFresh {
		t.Error("expected DataFresh = true (90s < 2m)")
	}
}

func TestGetHealthStatus_DataStale_DegradedFromHealthy(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0, // No errors, would be healthy
		maxConsecutiveErrors: 5,
		lastFetchTime:        time.Now().Add(-15 * time.Minute), // 15m > 2x1m = 2m
		startTime:            time.Now().Add(-1 * time.Hour),
	}

	status := d.GetHealthStatus()

	if status.Status != "degraded" {
		t.Errorf("expected status = degraded (stale data), got %s", status.Status)
	}

	if status.DataFresh {
		t.Error("expected DataFresh = false (15m > 2m)")
	}
}

func TestGetHealthStatus_DataStale_RemainsUnhealthy(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    5, // At max -> unhealthy
		maxConsecutiveErrors: 5,
		lastFetchError:       "persistent error",
		lastFetchTime:        time.Now().Add(-20 * time.Minute), // Also stale
		startTime:            time.Now().Add(-2 * time.Hour),
	}

	status := d.GetHealthStatus()

	// Should remain unhealthy, not downgrade to degraded
	if status.Status != "unhealthy" {
		t.Errorf("expected status = unhealthy (should not downgrade), got %s", status.Status)
	}

	if status.DataFresh {
		t.Error("expected DataFresh = false (20m > 2m)")
	}
}

func TestGetHealthStatus_DataFresh_ZeroFetchTime(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		lastFetchTime:        time.Time{}, // Zero value - no fetch yet
		startTime:            time.Now(),
	}

	status := d.GetHealthStatus()

	if status.Status != "healthy" {
		t.Errorf("expected status = healthy, got %s", status.Status)
	}

	if !status.DataFresh {
		t.Error("expected DataFresh = true for zero lastFetchTime (no fetch yet)")
	}
}

func TestGetHealthStatus_SensorExpired_DegradedFromHealthy(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		lastFetchTime:        time.Now(),
		startTime:            time.Now().Add(-1 * time.Hour),
		sensorExpiresAt:      time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	status := d.GetHealthStatus()

	if status.Status != "degraded" {
		t.Errorf("expected status = degraded (sensor expired), got %s", status.Status)
	}

	if !status.SensorExpired {
		t.Error("expected SensorExpired = true")
	}

	if !status.DataFresh {
		t.Error("expected DataFresh = true")
	}
}

func TestGetHealthStatus_SensorExpired_RemainsUnhealthy(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    5,
		maxConsecutiveErrors: 5,
		lastFetchError:       "persistent error",
		lastFetchTime:        time.Now().Add(-20 * time.Minute),
		startTime:            time.Now().Add(-2 * time.Hour),
		sensorExpiresAt:      time.Now().Add(-1 * time.Hour), // Also expired
	}

	status := d.GetHealthStatus()

	if status.Status != "unhealthy" {
		t.Errorf("expected status = unhealthy (should not downgrade), got %s", status.Status)
	}

	if !status.SensorExpired {
		t.Error("expected SensorExpired = true")
	}
}

func TestGetHealthStatus_SensorNotExpired(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		lastFetchTime:        time.Now(),
		startTime:            time.Now().Add(-1 * time.Hour),
		sensorExpiresAt:      time.Now().Add(5 * 24 * time.Hour), // Expires in 5 days
	}

	status := d.GetHealthStatus()

	if status.Status != "healthy" {
		t.Errorf("expected status = healthy, got %s", status.Status)
	}

	if status.SensorExpired {
		t.Error("expected SensorExpired = false")
	}
}

func TestGetHealthStatus_SensorExpiresAt_ZeroValue(t *testing.T) {
	d := &Daemon{
		ctx:                  context.Background(),
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5,
		lastFetchTime:        time.Now(),
		startTime:            time.Now().Add(-1 * time.Hour),
		sensorExpiresAt:      time.Time{}, // Not set yet
	}

	status := d.GetHealthStatus()

	if status.Status != "healthy" {
		t.Errorf("expected status = healthy, got %s", status.Status)
	}

	if status.SensorExpired {
		t.Error("expected SensorExpired = false for zero sensorExpiresAt")
	}
}
