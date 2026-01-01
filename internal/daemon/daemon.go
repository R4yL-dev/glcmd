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
	"fmt"
	"time"

	"github.com/R4yL-dev/glcmd/internal/glucosemeasurement"
	"github.com/R4yL-dev/glcmd/internal/libreclient"
	"github.com/R4yL-dev/glcmd/internal/models"
	"github.com/R4yL-dev/glcmd/internal/storage"
	"github.com/R4yL-dev/glcmd/internal/utils/timeparser"
)

// Daemon represents the background service that continuously fetches
// glucose data from the LibreView API.
//
// It manages:
//   - A ticker for periodic fetching (default 5 minutes)
//   - Context-based lifecycle management for graceful shutdown
//   - Storage backend for persisting fetched data
//   - Authentication with LibreView API
type Daemon struct {
	storage   storage.Storage
	ctx       context.Context
	cancel    context.CancelFunc
	ticker    *time.Ticker
	interval  time.Duration
	client    *libreclient.Client
	email     string
	password  string
	token     string
	accountID string
	patientID string
}

// New creates a new Daemon instance.
//
// Parameters:
//   - storage: The storage backend for persisting data
//   - interval: The time between fetch operations (e.g., 5*time.Minute)
//   - email: LibreView email for authentication
//   - password: LibreView password for authentication
//
// The daemon is created with a background context that can be cancelled
// via the Stop() method for graceful shutdown.
func New(storage storage.Storage, interval time.Duration, email string, password string) (*Daemon, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Daemon{
		storage:  storage,
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
		client:   libreclient.NewClient(nil),
		email:    email,
		password: password,
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
	if err := d.authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Step 2: Initial fetch (historical data from /graph)
	if err := d.initialFetch(); err != nil {
		return fmt.Errorf("initial fetch failed: %w", err)
	}

	// TODO (Step 7): Implement ticker loop and graceful shutdown
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

// authenticate authenticates with the LibreView API and stores credentials.
func (d *Daemon) authenticate() error {
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	token, userID, accountID, err := d.client.Authenticate(ctx, d.email, d.password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	d.token = token
	d.accountID = accountID
	// userID is not the same as patientID, we'll get patientID from /connections
	_ = userID

	return nil
}

// initialFetch performs the initial data fetch from /connections and /graph.
func (d *Daemon) initialFetch() error {
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	// First, get connections to obtain patientID
	connectionsResp, err := d.client.GetConnections(ctx, d.token, d.accountID)
	if err != nil {
		return fmt.Errorf("failed to get connections: %w", err)
	}

	if len(connectionsResp.Data) == 0 {
		return fmt.Errorf("no patient data in connections response")
	}

	d.patientID = connectionsResp.Data[0].PatientID

	// Store current measurement from /connections
	if err := d.storeCurrentMeasurement(&connectionsResp.Data[0].GlucoseMeasurement); err != nil {
		return fmt.Errorf("failed to store current measurement: %w", err)
	}

	// Now fetch historical data from /graph
	graphResp, err := d.client.GetGraph(ctx, d.token, d.accountID, d.patientID)
	if err != nil {
		return fmt.Errorf("failed to get graph data: %w", err)
	}

	// Store historical measurements
	for _, point := range graphResp.Data.GraphData {
		if err := d.storeHistoricalMeasurement(&point); err != nil {
			return fmt.Errorf("failed to store historical measurement: %w", err)
		}
	}

	// Store sensor configuration
	if err := d.storeSensor(&graphResp.Data.Connection.Sensor); err != nil {
		return fmt.Errorf("failed to store sensor: %w", err)
	}

	return nil
}

// fetch retrieves the latest glucose data from /connections.
// Used for periodic updates (every 5 minutes).
func (d *Daemon) fetch() error {
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	connectionsResp, err := d.client.GetConnections(ctx, d.token, d.accountID)
	if err != nil {
		return fmt.Errorf("failed to get connections: %w", err)
	}

	if len(connectionsResp.Data) == 0 {
		return fmt.Errorf("no patient data in connections response")
	}

	return d.storeCurrentMeasurement(&connectionsResp.Data[0].GlucoseMeasurement)
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

	measurement := &glucosemeasurement.GlucoseMeasurement{
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

	return d.storage.SaveMeasurement(measurement)
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

	measurement := &glucosemeasurement.GlucoseMeasurement{
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

	return d.storage.SaveMeasurement(measurement)
}

// storeSensor stores sensor configuration.
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

	sensorConfig := &models.SensorConfig{
		SerialNumber: sensor.SN,
		Activation:   activationTime,
		DeviceID:     sensor.DeviceID,
		SensorType:   sensor.PT,
		WarrantyDays: sensor.W,
		IsActive:     sensor.S,
		LowJourney:   sensor.LJ,
		DetectedAt:   time.Now().UTC(),
	}

	return d.storage.SaveSensor(sensorConfig)
}
