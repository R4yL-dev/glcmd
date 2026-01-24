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
//   - A ticker for periodic display (configurable via GLCMD_DISPLAY_INTERVAL)
//   - Context-based lifecycle management for graceful shutdown
//   - Business logic services for persisting fetched data
//   - Authentication with LibreView API
type Daemon struct {
	glucoseService      service.GlucoseService
	sensorService       service.SensorService
	configService       service.ConfigService
	ctx                 context.Context
	cancel              context.CancelFunc
	ticker              *time.Ticker
	displayTicker       *time.Ticker
	config              *Config
	client              *libreclient.Client
	email               string
	password            string
	token               string
	accountID           string
	patientID           string
	consecutiveErrors    int       // Counter for consecutive fetch errors
	maxConsecutiveErrors int       // Max allowed consecutive errors before alerting
	lastFetchError       string    // Last fetch error message (empty if no error)
	lastFetchTime        time.Time // Last successful fetch time
	startTime            time.Time // Daemon start time
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
	slog.Info("starting daemon",
		"fetchInterval", d.config.FetchInterval,
		"displayInterval", d.config.DisplayInterval,
		"emojisEnabled", d.config.EnableEmojis,
	)

	// Step 1: Authenticate
	slog.Info("authenticating with LibreView API")
	if err := d.authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	slog.Info("authentication successful")

	// Step 2: Initial fetch (historical data from /graph)
	slog.Info("performing initial data fetch")
	if err := d.initialFetch(); err != nil {
		return fmt.Errorf("initial fetch failed: %w", err)
	}
	slog.Info("initial fetch completed successfully")

	// Step 3: Start ticker for periodic fetches
	d.ticker = time.NewTicker(d.config.FetchInterval)
	defer d.ticker.Stop()

	// Step 4: Start ticker for periodic display
	d.displayTicker = time.NewTicker(d.config.DisplayInterval)
	defer d.displayTicker.Stop()

	slog.Info("daemon started successfully",
		"fetchInterval", d.config.FetchInterval,
		"displayInterval", d.config.DisplayInterval,
	)

	// Step 5: Main loop - fetch periodically until stopped
	for {
		select {
		case <-d.ticker.C:
			// Time to fetch new data
			slog.Info("fetching new measurement")
			if err := d.fetch(); err != nil {
				// Increment error counter
				d.consecutiveErrors++
				d.lastFetchError = err.Error()

				// Log error but don't stop the daemon
				// Network errors are expected and should not kill the daemon
				slog.Error("fetch failed",
					"error", err,
					"consecutiveErrors", d.consecutiveErrors,
				)

				// Circuit breaker: alert after max consecutive errors
				if d.consecutiveErrors >= d.maxConsecutiveErrors {
					slog.Error("CRITICAL: max consecutive errors reached, daemon may be unhealthy",
						"consecutiveErrors", d.consecutiveErrors,
						"maxAllowed", d.maxConsecutiveErrors,
					)
				}
			} else {
				// Reset error counter on success
				if d.consecutiveErrors > 0 {
					slog.Info("fetch recovered after errors",
						"previousConsecutiveErrors", d.consecutiveErrors,
					)
					d.consecutiveErrors = 0
				}
				d.lastFetchError = ""
				d.lastFetchTime = time.Now()
				slog.Info("measurement fetched successfully")
			}

		case <-d.displayTicker.C:
			// Time to display the last measurement
			d.displayLastMeasurement()

		case <-d.ctx.Done():
			// Context cancelled - graceful shutdown
			slog.Info("daemon shutting down gracefully")
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
//   - Stops the tickers if running
//   - Allows in-progress operations to complete
//
// After calling Stop(), the Run() method will return.
func (d *Daemon) Stop() {
	if d.ticker != nil {
		d.ticker.Stop()
	}
	if d.displayTicker != nil {
		d.displayTicker.Stop()
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
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	// First, get connections to obtain patientID
	slog.Debug("fetching connections to obtain patientID")
	connectionsResp, err := d.client.GetConnections(ctx, d.token, d.accountID)
	if err != nil {
		slog.Error("failed to get connections", "error", err)
		return fmt.Errorf("failed to get connections: %w", err)
	}

	if len(connectionsResp.Data) == 0 {
		slog.Error("no patient data in connections response")
		return fmt.Errorf("no patient data in connections response")
	}

	d.patientID = connectionsResp.Data[0].PatientID
	slog.Debug("patient ID obtained", "patientID", logger.RedactSensitive(d.patientID))

	// Store current measurement from /connections
	if err := d.storeCurrentMeasurement(&connectionsResp.Data[0].GlucoseMeasurement); err != nil {
		slog.Error("failed to store current measurement", "error", err)
		return fmt.Errorf("failed to store current measurement: %w", err)
	}
	slog.Debug("current measurement stored")

	// Now fetch historical data from /graph
	slog.Debug("fetching historical data from /graph")
	graphResp, err := d.client.GetGraph(ctx, d.token, d.accountID, d.patientID)
	if err != nil {
		slog.Error("failed to get graph data", "error", err)
		return fmt.Errorf("failed to get graph data: %w", err)
	}

	// Store historical measurements
	storedCount := 0
	for _, point := range graphResp.Data.GraphData {
		if err := d.storeHistoricalMeasurement(&point); err != nil {
			slog.Error("failed to store historical measurement", "error", err)
			return fmt.Errorf("failed to store historical measurement: %w", err)
		}
		storedCount++
	}
	slog.Info("historical measurements stored", "count", storedCount)

	// Store sensor configuration
	sensor := &graphResp.Data.Connection.Sensor
	if err := d.storeSensor(sensor); err != nil {
		slog.Error("failed to store sensor", "error", err)
		return fmt.Errorf("failed to store sensor: %w", err)
	}
	slog.Debug("sensor configuration stored", "serialNumber", sensor.SN)

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
		slog.Error("no patient data in periodic fetch")
		return fmt.Errorf("no patient data in connections response")
	}

	gm := &connectionsResp.Data[0].GlucoseMeasurement
	slog.Debug("measurement received", "value", gm.Value, "trendArrow", gm.TrendArrow)

	// Store the measurement
	if err := d.storeCurrentMeasurement(gm); err != nil {
		return err
	}

	// Also store/update the sensor
	sensor := &connectionsResp.Data[0].Sensor
	if err := d.storeSensor(sensor); err != nil {
		// Log but don't fail the fetch for sensor errors
		slog.Warn("failed to store sensor", "error", err)
	}

	return nil
}

// storeCurrentMeasurement stores a current measurement (from /connections).
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
}) error {
	timestamp, err := timeparser.ParseLibreViewTimestamp(gm.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
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
		MeasurementColor: gm.MeasurementColor,
		GlucoseUnits:     gm.GlucoseUnits,
		IsHigh:           gm.IsHigh,
		IsLow:            gm.IsLow,
		Type:             domain.MeasurementTypeCurrent,
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	if err := d.glucoseService.SaveMeasurement(ctx, measurement); err != nil {
		return err
	}

	// Update LastMeasurementAt on the current sensor
	if err := d.sensorService.UpdateLastMeasurementIfNewer(ctx, measurement.Timestamp); err != nil {
		slog.Warn("failed to update sensor LastMeasurementAt", "error", err)
	}

	return nil
}

// storeHistoricalMeasurement stores a historical measurement (from /graph).
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
}) error {
	factoryTimestamp, err := timeparser.ParseLibreViewTimestamp(point.FactoryTimestamp)
	if err != nil {
		return fmt.Errorf("failed to parse factory timestamp: %w", err)
	}

	timestamp, err := timeparser.ParseLibreViewTimestamp(point.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	measurement := &domain.GlucoseMeasurement{
		FactoryTimestamp: factoryTimestamp,
		Timestamp:        timestamp,
		Value:            point.Value,
		ValueInMgPerDl:   point.ValueInMgPerDl,
		TrendArrow:       nil, // Historical data has no trend arrow
		MeasurementColor: point.MeasurementColor,
		GlucoseUnits:     point.GlucoseUnits,
		IsHigh:           point.IsHigh,
		IsLow:            point.IsLow,
		Type:             point.Type,
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	if err := d.glucoseService.SaveMeasurement(ctx, measurement); err != nil {
		return err
	}

	// Update LastMeasurementAt on the current sensor
	if err := d.sensorService.UpdateLastMeasurementIfNewer(ctx, measurement.Timestamp); err != nil {
		slog.Warn("failed to update sensor LastMeasurementAt", "error", err)
	}

	return nil
}

// storeSensor stores sensor configuration and handles sensor changes.
// The sensor change detection logic (setting EndedAt on old sensor)
// is handled by SensorService.HandleSensorChange() within a transaction.
func (d *Daemon) storeSensor(sensor *libreclient.SensorData) error {
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
	return d.sensorService.HandleSensorChange(ctx, sensorConfig)
}

// displayLastMeasurement retrieves and displays the last recorded measurement.
// This is called every minute by the displayTicker.
func (d *Daemon) displayLastMeasurement() {
	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	measurement, err := d.glucoseService.GetLatestMeasurement(ctx)
	if err != nil {
		slog.Warn("no measurement available to display", "error", err)
		return
	}

	// Build trend arrow display
	trendArrowStr := ""
	if measurement.TrendArrow != nil {
		if d.config.EnableEmojis {
			switch *measurement.TrendArrow {
			case domain.TrendArrowFallingRapidly:
				trendArrowStr = "⬇️⬇️"
			case domain.TrendArrowFalling:
				trendArrowStr = "⬇️"
			case domain.TrendArrowStable:
				trendArrowStr = "➡️"
			case domain.TrendArrowRising:
				trendArrowStr = "⬆️"
			case domain.TrendArrowRisingRapidly:
				trendArrowStr = "⬆️⬆️"
			}
		} else {
			switch *measurement.TrendArrow {
			case domain.TrendArrowFallingRapidly:
				trendArrowStr = "Falling rapidly"
			case domain.TrendArrowFalling:
				trendArrowStr = "Falling"
			case domain.TrendArrowStable:
				trendArrowStr = "Stable"
			case domain.TrendArrowRising:
				trendArrowStr = "Rising"
			case domain.TrendArrowRisingRapidly:
				trendArrowStr = "Rising rapidly"
			}
		}
	}

	// Build status indicator
	statusStr := ""
	if d.config.EnableEmojis {
		switch measurement.MeasurementColor {
		case domain.MeasurementColorNormal:
			statusStr = "🟢 Normal"
		case domain.MeasurementColorWarning:
			statusStr = "🟠 Warning"
		case domain.MeasurementColorCritical:
			statusStr = "🔴 Critical"
		}
	} else {
		switch measurement.MeasurementColor {
		case domain.MeasurementColorNormal:
			statusStr = "Normal"
		case domain.MeasurementColorWarning:
			statusStr = "Warning"
		case domain.MeasurementColorCritical:
			statusStr = "Critical"
		}
	}

	// Log the measurement with all relevant information
	slog.Info("last measurement",
		"value", fmt.Sprintf("%.1f mmol/L (%d mg/dL)", measurement.Value, measurement.ValueInMgPerDl),
		"trend", trendArrowStr,
		"status", statusStr,
		"timestamp", measurement.Timestamp.Format("2006-01-02 15:04:05"),
	)
}
