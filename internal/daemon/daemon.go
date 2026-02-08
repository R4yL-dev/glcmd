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
//	d, err := daemon.New(storage, 5*time.Minute, "email@example.com", "password")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := d.Run(); err != nil {
//	    log.Fatal(err)
//	}
package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/libreclient"
	"github.com/R4yL-dev/glcmd/internal/logger"
	"github.com/R4yL-dev/glcmd/internal/service"
	"github.com/R4yL-dev/glcmd/internal/utils/timeparser"
)

// Daemon represents the background service that continuously fetches
// glucose data from the LibreView API.
//
// It manages:
//   - A ticker for periodic fetching (configurable via GLCMD_FETCH_INTERVAL)
//   - Context-based lifecycle management for graceful shutdown
//   - Business logic services for persisting fetched data
//   - Authentication with LibreView API
type Daemon struct {
	glucoseService       service.GlucoseService
	sensorService        service.SensorService
	configService        service.ConfigService
	ctx                  context.Context
	cancel               context.CancelFunc
	ticker               *time.Ticker
	config               *Config
	client               *libreclient.Client
	email                string
	password             string
	token                string
	accountID            string
	patientID            string
	consecutiveErrors    int       // Counter for consecutive fetch errors
	maxConsecutiveErrors int       // Max allowed consecutive errors before alerting
	lastFetchError       string    // Last fetch error message (empty if no error)
	lastFetchTime        time.Time // Last successful fetch time
	startTime            time.Time // Daemon start time
	lastTargets          *domain.GlucoseTargets // Cache to avoid redundant saves
}

// New creates a new Daemon instance.
//
// Parameters:
//   - glucoseService: Service for glucose measurement business logic
//   - sensorService: Service for sensor management business logic
//   - configService: Service for configuration management
//   - config: Daemon configuration (intervals, emojis, etc.)
//   - email: LibreView email for authentication
//   - password: LibreView password for authentication
//
// The daemon is created with a background context that can be cancelled
// via the Stop() method for graceful shutdown.
func New(
	glucoseService service.GlucoseService,
	sensorService service.SensorService,
	configService service.ConfigService,
	config *Config,
	email string,
	password string,
) (*Daemon, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Daemon{
		glucoseService:       glucoseService,
		sensorService:        sensorService,
		configService:        configService,
		ctx:                  ctx,
		cancel:               cancel,
		config:               config,
		client:               libreclient.NewClient(nil),
		email:                email,
		password:             password,
		consecutiveErrors:    0,
		maxConsecutiveErrors: 5, // Alert after 5 consecutive errors
		lastFetchError:       "",
		lastFetchTime:        time.Time{},
		startTime:            time.Now(),
	}, nil
}

// Run starts the daemon's main loop.
//
// This method blocks until the daemon is stopped via Stop() or an
// unrecoverable error occurs.
//
// The main loop:
//   - Authenticates with LibreView API
//   - Performs an initial fetch to populate historical data (12h)
//   - Starts a ticker for periodic fetches at the configured interval
//   - Waits for context cancellation to stop gracefully
//
// Returns an error if the daemon cannot start or encounters a fatal error.
func (d *Daemon) Run() error {
	// Step 1: Authenticate
	authStart := time.Now()
	if err := d.authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	slog.Info("authenticated", "duration", time.Since(authStart))

	// Step 2: Initial fetch (historical data from /graph)
	if err := d.initialFetch(); err != nil {
		return fmt.Errorf("initial fetch failed: %w", err)
	}

	// Step 3: Start ticker for periodic fetches
	d.ticker = time.NewTicker(d.config.FetchInterval)
	defer d.ticker.Stop()

	slog.Info("ready", "fetchInterval", d.config.FetchInterval)

	// Step 4: Main loop - fetch periodically until stopped
	for {
		select {
		case <-d.ticker.C:
			start := time.Now()
			if err := d.fetch(); err != nil {
				d.consecutiveErrors++
				d.lastFetchError = err.Error()

				slog.Error("fetch failed",
					"error", err,
					"duration", time.Since(start),
				)

				// Circuit breaker: alert after max consecutive errors
				if d.consecutiveErrors >= d.maxConsecutiveErrors {
					slog.Error("CRITICAL: max consecutive errors reached",
						"consecutiveErrors", d.consecutiveErrors,
						"maxAllowed", d.maxConsecutiveErrors,
					)
				}
			} else {
				duration := time.Since(start)
				if d.consecutiveErrors > 0 {
					slog.Info("fetch recovered", "previousErrors", d.consecutiveErrors)
				}
				d.consecutiveErrors = 0
				d.lastFetchError = ""
				d.lastFetchTime = time.Now()

				slog.Info("measurement fetched", "duration", duration)
			}

		case <-d.ctx.Done():
			return nil
		}
	}
}

// GetHealthStatus returns the current health status of the daemon.
// This is used by the healthcheck HTTP endpoint.
func (d *Daemon) GetHealthStatus() HealthStatus {
	status := "healthy"

	// Determine status based on consecutive errors
	if d.consecutiveErrors >= d.maxConsecutiveErrors {
		status = "unhealthy"
	} else if d.consecutiveErrors > 0 {
		status = "degraded"
	}

	return HealthStatus{
		Status:            status,
		Timestamp:         time.Now(),
		Uptime:            time.Since(d.startTime).String(),
		ConsecutiveErrors: d.consecutiveErrors,
		LastFetchError:    d.lastFetchError,
		LastFetchTime:     d.lastFetchTime,
	}
}

// HealthStatus represents the daemon's health status.
// This is exported for use by the healthcheck package.
type HealthStatus struct {
	Status            string    `json:"status"`
	Timestamp         time.Time `json:"timestamp"`
	Uptime            string    `json:"uptime"`
	ConsecutiveErrors int       `json:"consecutiveErrors"`
	LastFetchError    string    `json:"lastFetchError"`
	LastFetchTime     time.Time `json:"lastFetchTime"`
	DatabaseConnected bool      `json:"databaseConnected"` // Database health status
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

// authenticate authenticates with the LibreView API and stores credentials.
func (d *Daemon) authenticate() error {
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	token, userID, accountID, err := d.client.Authenticate(ctx, d.email, d.password)
	if err != nil {
		slog.Error("authentication failed", "error", err)
		return fmt.Errorf("authentication failed: %w", err)
	}

	d.token = token
	d.accountID = accountID
	// userID is not the same as patientID, we'll get patientID from /connections
	_ = userID

	slog.Debug("authentication successful", "accountID", logger.RedactSensitive(accountID))
	return nil
}

// initialFetch performs the initial data fetch from /connections and /graph.
func (d *Daemon) initialFetch() error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	// First, get connections to obtain patientID
	slog.Debug("fetching connections to obtain patientID")
	connectionsResp, err := d.client.GetConnections(ctx, d.token, d.accountID)
	if err != nil {
		return fmt.Errorf("failed to get connections: %w", err)
	}

	if len(connectionsResp.Data) == 0 {
		return fmt.Errorf("no patient data in connections response")
	}

	d.patientID = connectionsResp.Data[0].PatientID
	slog.Debug("patient ID obtained", "patientID", logger.RedactSensitive(d.patientID))

	// Store current measurement from /connections
	if _, err := d.storeCurrentMeasurement(&connectionsResp.Data[0].GlucoseMeasurement); err != nil {
		return fmt.Errorf("failed to store current measurement: %w", err)
	}

	// Now fetch historical data from /graph
	slog.Debug("fetching historical data from /graph")
	graphResp, err := d.client.GetGraph(ctx, d.token, d.accountID, d.patientID)
	if err != nil {
		return fmt.Errorf("failed to get graph data: %w", err)
	}

	// Store historical measurements and count new vs skipped
	newCount := 0
	skippedCount := 0
	for _, point := range graphResp.Data.GraphData {
		inserted, err := d.storeHistoricalMeasurement(&point)
		if err != nil {
			return fmt.Errorf("failed to store historical measurement: %w", err)
		}
		if inserted {
			newCount++
		} else {
			skippedCount++
		}
	}

	// Store sensor configuration
	sensor := &graphResp.Data.Connection.Sensor
	if err := d.storeSensor(sensor); err != nil {
		return fmt.Errorf("failed to store sensor: %w", err)
	}

	// Store glucose targets from /connections response
	d.storeTargets(connectionsResp)

	slog.Info("initial fetch completed",
		"new", newCount,
		"skipped", skippedCount,
		"duration", time.Since(start),
	)

	return nil
}

// fetch retrieves the latest glucose data from /connections.
// Used for periodic updates (every 5 minutes).
// If authentication fails (401), automatically re-authenticates with retry logic.
func (d *Daemon) fetch() error {
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	connectionsResp, err := d.client.GetConnections(ctx, d.token, d.accountID)
	if err != nil {
		// Check if it's an authentication error
		var authErr *libreclient.AuthError
		if errors.As(err, &authErr) {
			slog.Warn("authentication token expired, re-authenticating with retry")

			// Re-authenticate with retry (max 3 attempts)
			maxRetries := 3
			var lastErr error
			for attempt := 1; attempt <= maxRetries; attempt++ {
				slog.Info("re-authentication attempt", "attempt", attempt, "maxRetries", maxRetries)

				if err := d.authenticate(); err != nil {
					lastErr = err
					slog.Warn("re-authentication attempt failed",
						"attempt", attempt,
						"error", err,
					)

					// Exponential backoff: wait before retrying
					if attempt < maxRetries {
						backoff := time.Duration(attempt*attempt) * time.Second
						slog.Info("waiting before retry", "backoff", backoff)
						time.Sleep(backoff)
					}
					continue
				}

				// Re-authentication successful, retry the fetch
				ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
				defer cancel()

				connectionsResp, err = d.client.GetConnections(ctx, d.token, d.accountID)
				if err != nil {
					slog.Error("failed to get connections after re-authentication", "error", err)
					return fmt.Errorf("failed to get connections after re-auth: %w", err)
				}

				slog.Info("re-authentication successful, fetch completed")
				break
			}

			// If all retry attempts failed
			if lastErr != nil && connectionsResp == nil {
				slog.Error("re-authentication failed after all retries", "attempts", maxRetries, "error", lastErr)
				return fmt.Errorf("re-authentication failed after %d attempts: %w", maxRetries, lastErr)
			}
		} else {
			slog.Error("failed to get connections during periodic fetch", "error", err)
			return fmt.Errorf("failed to get connections: %w", err)
		}
	}

	if len(connectionsResp.Data) == 0 {
		return fmt.Errorf("no patient data in connections response")
	}

	gm := &connectionsResp.Data[0].GlucoseMeasurement

	// Store the measurement
	if _, err := d.storeCurrentMeasurement(gm); err != nil {
		return err
	}

	// Debug: log all measurement data
	slog.Debug("measurement",
		"value", gm.Value,
		"valueInMgPerDl", gm.ValueInMgPerDl,
		"trendArrow", gm.TrendArrow,
		"measurementColor", gm.MeasurementColor,
		"timestamp", gm.Timestamp,
	)

	// Also store/update the sensor
	sensor := &connectionsResp.Data[0].Sensor
	if err := d.storeSensor(sensor); err != nil {
		// Log but don't fail the fetch for sensor errors
		slog.Warn("failed to store sensor", "error", err)
	}

	// Store glucose targets
	d.storeTargets(connectionsResp)

	return nil
}

// storeCurrentMeasurement stores a current measurement (from /connections).
// Returns (true, nil) if inserted, (false, nil) if duplicate.
func (d *Daemon) storeCurrentMeasurement(gm *struct {
	ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
	Value            float64 `json:"Value"`
	TrendArrow       int     `json:"TrendArrow"`
	TrendMessage     string  `json:"TrendMessage"`
	MeasurementColor int     `json:"MeasurementColor"`
	GlucoseUnits     int     `json:"GlucoseUnits"`
	Timestamp        string  `json:"Timestamp"`
	IsHigh           bool    `json:"isHigh"`
	IsLow            bool    `json:"isLow"`
}) (bool, error) {
	timestamp, err := timeparser.ParseLibreViewTimestamp(gm.Timestamp)
	if err != nil {
		return false, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	trendArrow := gm.TrendArrow
	var trendMessage *string
	if gm.TrendMessage != "" {
		trendMessage = &gm.TrendMessage
	}

	measurement := &domain.GlucoseMeasurement{
		FactoryTimestamp: timestamp,
		Timestamp:        timestamp,
		Value:            gm.Value,
		ValueInMgPerDl:   gm.ValueInMgPerDl,
		TrendArrow:       &trendArrow,
		TrendMessage:     trendMessage,
		GlucoseColor:     gm.MeasurementColor,
		GlucoseUnits:     gm.GlucoseUnits,
		IsHigh:           gm.IsHigh,
		IsLow:            gm.IsLow,
		Type:             domain.GlucoseTypeCurrent,
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	inserted, err := d.glucoseService.SaveMeasurement(ctx, measurement)
	if err != nil {
		return false, err
	}

	// Update LastMeasurementAt on the current sensor
	if err := d.sensorService.UpdateLastMeasurementIfNewer(ctx, measurement.Timestamp); err != nil {
		slog.Warn("failed to update sensor LastMeasurementAt", "error", err)
	}

	return inserted, nil
}

// storeHistoricalMeasurement stores a historical measurement (from /graph).
// Returns (true, nil) if inserted, (false, nil) if duplicate.
func (d *Daemon) storeHistoricalMeasurement(point *struct {
	FactoryTimestamp string  `json:"FactoryTimestamp"`
	Timestamp        string  `json:"Timestamp"`
	ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
	Value            float64 `json:"Value"`
	MeasurementColor int     `json:"MeasurementColor"`
	GlucoseUnits     int     `json:"GlucoseUnits"`
	IsHigh           bool    `json:"isHigh"`
	IsLow            bool    `json:"isLow"`
	Type             int     `json:"type"`
}) (bool, error) {
	factoryTimestamp, err := timeparser.ParseLibreViewTimestamp(point.FactoryTimestamp)
	if err != nil {
		return false, fmt.Errorf("failed to parse factory timestamp: %w", err)
	}

	timestamp, err := timeparser.ParseLibreViewTimestamp(point.Timestamp)
	if err != nil {
		return false, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	measurement := &domain.GlucoseMeasurement{
		FactoryTimestamp: factoryTimestamp,
		Timestamp:        timestamp,
		Value:            point.Value,
		ValueInMgPerDl:   point.ValueInMgPerDl,
		TrendArrow:       nil, // Historical data has no trend arrow
		GlucoseColor:     point.MeasurementColor,
		GlucoseUnits:     point.GlucoseUnits,
		IsHigh:           point.IsHigh,
		IsLow:            point.IsLow,
		Type:             point.Type,
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	inserted, err := d.glucoseService.SaveMeasurement(ctx, measurement)
	if err != nil {
		return false, err
	}

	// Update LastMeasurementAt on the current sensor
	if err := d.sensorService.UpdateLastMeasurementIfNewer(ctx, measurement.Timestamp); err != nil {
		slog.Warn("failed to update sensor LastMeasurementAt", "error", err)
	}

	return inserted, nil
}

// storeSensor stores sensor configuration and handles sensor changes.
// The sensor change detection logic (setting EndedAt on old sensor)
// is handled by SensorService.HandleSensorChange() within a transaction.
func (d *Daemon) storeSensor(sensor *libreclient.SensorData) error {
	start := time.Now()

	// Convert Unix timestamp to time.Time (sensor.A is activation time)
	activationTime := time.Unix(int64(sensor.A), 0).UTC()

	// Calculate duration and expiration based on sensor type
	durationDays := domain.SensorDurationDays(sensor.PT)
	expiresAt := activationTime.AddDate(0, 0, durationDays)

	sensorConfig := &domain.SensorConfig{
		SerialNumber: sensor.SN,
		Activation:   activationTime,
		ExpiresAt:    expiresAt,
		SensorType:   sensor.PT,
		DurationDays: durationDays,
		DetectedAt:   time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	// HandleSensorChange manages sensor change detection atomically
	if err := d.sensorService.HandleSensorChange(ctx, sensorConfig); err != nil {
		return err
	}

	// Debug: log all sensor data (same pattern as measurements in fetch())
	slog.Debug("sensor",
		"serialNumber", sensor.SN,
		"activation", sensorConfig.Activation,
		"expiresAt", sensorConfig.ExpiresAt,
		"sensorType", sensor.PT,
		"durationDays", sensorConfig.DurationDays,
		"duration", time.Since(start),
	)
	return nil
}

// storeTargets extracts glucose targets from a ConnectionsResponse and saves them.
// Uses in-memory cache to avoid redundant saves when values haven't changed.
func (d *Daemon) storeTargets(resp *libreclient.ConnectionsResponse) {
	if len(resp.Data) == 0 {
		return
	}

	data := &resp.Data[0]
	if data.TargetHigh == 0 && data.TargetLow == 0 {
		return
	}

	// Check if values have changed
	if d.lastTargets != nil &&
		d.lastTargets.TargetHigh == data.TargetHigh &&
		d.lastTargets.TargetLow == data.TargetLow &&
		d.lastTargets.UnitOfMeasure == data.Uom {
		return // Unchanged, skip save
	}

	targets := &domain.GlucoseTargets{
		TargetHigh:    data.TargetHigh,
		TargetLow:     data.TargetLow,
		UnitOfMeasure: data.Uom,
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	if err := d.configService.SaveGlucoseTargets(ctx, targets); err != nil {
		slog.Warn("failed to store glucose targets", "error", err)
		return
	}

	// Update cache on successful save
	d.lastTargets = targets
}

