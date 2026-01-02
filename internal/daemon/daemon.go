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
	"github.com/R4yL-dev/glcmd/internal/service"
	"github.com/R4yL-dev/glcmd/internal/utils/timeparser"
)

// Daemon represents the background service that continuously fetches
// glucose data from the LibreView API.
//
// It manages:
//   - A ticker for periodic fetching (default 5 minutes)
//   - A ticker for periodic display (1 minute)
//   - Context-based lifecycle management for graceful shutdown
//   - Business logic services for persisting fetched data
//   - Authentication with LibreView API
type Daemon struct {
	glucoseService service.GlucoseService
	sensorService  service.SensorService
	configService  service.ConfigService
	ctx            context.Context
	cancel         context.CancelFunc
	ticker         *time.Ticker
	displayTicker  *time.Ticker
	interval       time.Duration
	client         *libreclient.Client
	email          string
	password       string
	token          string
	accountID      string
	patientID      string
}

// New creates a new Daemon instance.
//
// Parameters:
//   - glucoseService: Service for glucose measurement business logic
//   - sensorService: Service for sensor management business logic
//   - configService: Service for configuration management
//   - interval: The time between fetch operations (e.g., 5*time.Minute)
//   - email: LibreView email for authentication
//   - password: LibreView password for authentication
//
// The daemon is created with a background context that can be cancelled
// via the Stop() method for graceful shutdown.
func New(
	glucoseService service.GlucoseService,
	sensorService service.SensorService,
	configService service.ConfigService,
	interval time.Duration,
	email string,
	password string,
) (*Daemon, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Daemon{
		glucoseService: glucoseService,
		sensorService:  sensorService,
		configService:  configService,
		ctx:            ctx,
		cancel:         cancel,
		interval:       interval,
		client:         libreclient.NewClient(nil),
		email:          email,
		password:       password,
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
	slog.Info("starting daemon", "interval", d.interval)

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
	d.ticker = time.NewTicker(d.interval)
	defer d.ticker.Stop()

	// Step 4: Start ticker for periodic display (every 1 minute)
	d.displayTicker = time.NewTicker(1 * time.Minute)
	defer d.displayTicker.Stop()

	slog.Info("daemon started successfully", "interval", d.interval)

	// Step 5: Main loop - fetch periodically until stopped
	for {
		select {
		case <-d.ticker.C:
			// Time to fetch new data
			slog.Info("fetching new measurement")
			if err := d.fetch(); err != nil {
				// Log error but don't stop the daemon
				// Network errors are expected and should not kill the daemon
				slog.Error("fetch failed", "error", err)
			} else {
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

	slog.Debug("authentication successful", "accountID", accountID)
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
	slog.Debug("patient ID obtained", "patientID", d.patientID)

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
	if err := d.storeSensor(&graphResp.Data.Connection.Sensor); err != nil {
		slog.Error("failed to store sensor", "error", err)
		return fmt.Errorf("failed to store sensor: %w", err)
	}
	slog.Debug("sensor configuration stored", "serialNumber", graphResp.Data.Connection.Sensor.SN)

	return nil
}

// fetch retrieves the latest glucose data from /connections.
// Used for periodic updates (every 5 minutes).
// If authentication fails (401), automatically re-authenticates and retries once.
func (d *Daemon) fetch() error {
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	connectionsResp, err := d.client.GetConnections(ctx, d.token, d.accountID)
	if err != nil {
		// Check if it's an authentication error
		var authErr *libreclient.AuthError
		if errors.As(err, &authErr) {
			slog.Warn("authentication token expired, re-authenticating")

			// Re-authenticate
			if err := d.authenticate(); err != nil {
				slog.Error("re-authentication failed", "error", err)
				return fmt.Errorf("re-authentication failed: %w", err)
			}

			// Retry the fetch with new token
			ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
			defer cancel()

			connectionsResp, err = d.client.GetConnections(ctx, d.token, d.accountID)
			if err != nil {
				slog.Error("failed to get connections after re-authentication", "error", err)
				return fmt.Errorf("failed to get connections after re-auth: %w", err)
			}

			slog.Info("re-authentication successful, fetch completed")
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

	return d.storeCurrentMeasurement(gm)
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
		Type:             1, // Current measurement
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	return d.glucoseService.SaveMeasurement(ctx, measurement)
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

	return d.glucoseService.SaveMeasurement(ctx, measurement)
}

// storeSensor stores sensor configuration and handles sensor changes.
// The sensor change detection logic (deactivating old sensor, activating new sensor)
// is now handled by SensorService.HandleSensorChange() within a transaction.
func (d *Daemon) storeSensor(sensor *struct {
	DeviceID string `json:"deviceId"`
	SN       string `json:"sn"`
	A        int    `json:"a"`
	W        int    `json:"w"`
	PT       int    `json:"pt"`
	S        bool   `json:"s"`
	LJ       bool   `json:"lj"`
}) error {
	// Convert Unix timestamp to time.Time (sensor.A is activation time)
	activationTime := time.Unix(int64(sensor.A), 0).UTC()

	sensorConfig := &domain.SensorConfig{
		SerialNumber: sensor.SN,
		Activation:   activationTime,
		DeviceID:     sensor.DeviceID,
		SensorType:   sensor.PT,
		WarrantyDays: sensor.W,
		IsActive:     sensor.S,
		LowJourney:   sensor.LJ,
		DetectedAt:   time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	// HandleSensorChange manages sensor change detection and deactivation atomically
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
		switch *measurement.TrendArrow {
		case 1:
			trendArrowStr = "⬇️⬇️"
		case 2:
			trendArrowStr = "⬇️"
		case 3:
			trendArrowStr = "➡️"
		case 4:
			trendArrowStr = "⬆️"
		case 5:
			trendArrowStr = "⬆️⬆️"
		}
	}

	// Build status indicator
	statusStr := ""
	switch measurement.MeasurementColor {
	case 1:
		statusStr = "🟢 Normal"
	case 2:
		statusStr = "🟠 Warning"
	case 3:
		statusStr = "🔴 Critical"
	}

	// Log the measurement with all relevant information
	slog.Info("last measurement",
		"value", fmt.Sprintf("%.1f mmol/L (%d mg/dL)", measurement.Value, measurement.ValueInMgPerDl),
		"trend", trendArrowStr,
		"status", statusStr,
		"timestamp", measurement.Timestamp.Format("2006-01-02 15:04:05"),
	)
}
