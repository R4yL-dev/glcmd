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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/R4yL-dev/glcmd/internal/config"
	"github.com/R4yL-dev/glcmd/internal/credentials"
	"github.com/R4yL-dev/glcmd/internal/glucosemeasurement"
	"github.com/R4yL-dev/glcmd/internal/headers"
	"github.com/R4yL-dev/glcmd/internal/httpreq"
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
	creds     *credentials.Credentials
	client    *http.Client
	headers   *headers.Headers
	authToken string
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
	ctx, cancel := context.WithCancel(context.Background())

	creds, err := credentials.NewCredentials(email, password)
	if err != nil {
		cancel() // Clean up context
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	return &Daemon{
		storage:  storage,
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
		creds:    creds,
		client:   &http.Client{Timeout: 30 * time.Second},
		headers:  headers.NewHeaders(),
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

	// Step 2: Get patientID
	if err := d.fetchPatientID(); err != nil {
		return fmt.Errorf("failed to get patient ID: %w", err)
	}

	// Step 3: Initial fetch (historical data from /graph)
	if err := d.initialFetch(); err != nil {
		return fmt.Errorf("initial fetch failed: %w", err)
	}

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

// authenticate authenticates with the LibreView API and builds auth headers.
func (d *Daemon) authenticate() error {
	// Prepare credentials JSON
	payload, err := d.creds.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	// Make auth request
	req, err := httpreq.NewHttpReq("POST", config.LoginURL, payload, d.headers.DefaultHeader(), d.client)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	res, err := req.Do()
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}

	// Parse auth response
	auth, err := NewAuthFromResponse(res)
	if err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	// Build auth header for future requests
	d.headers.BuildAuthHeader(auth.Token, auth.UserID)
	d.authToken = auth.Token
	d.patientID = auth.UserID // Temporarily store userID, will get actual patientID next

	return nil
}

// AuthResponse represents the authentication response structure
type AuthResponse struct {
	Token   string
	UserID  string
}

// NewAuthFromResponse parses authentication response
func NewAuthFromResponse(data []byte) (*AuthResponse, error) {
	var tmp struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
			AuthTicket struct {
				Token string `json:"token"`
			} `json:"authTicket"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:  tmp.Data.AuthTicket.Token,
		UserID: tmp.Data.User.ID,
	}, nil
}

// fetchPatientID retrieves the patient ID from the /connections endpoint.
func (d *Daemon) fetchPatientID() error {
	req, err := httpreq.NewHttpReq("GET", config.ConnectionsURL, nil, d.headers.AuthHeader(), d.client)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := req.Do()
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	var result struct {
		Data []struct {
			PatientID string `json:"patientId"`
		} `json:"data"`
	}

	if err := json.Unmarshal(res, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return fmt.Errorf("no patient data in response")
	}

	d.patientID = result.Data[0].PatientID
	return nil
}

// initialFetch performs the initial data fetch from /graph (12h historical data).
func (d *Daemon) initialFetch() error {
	url := fmt.Sprintf("https://api-eu.libreview.io/llu/connections/%s/graph", d.patientID)
	req, err := httpreq.NewHttpReq("GET", url, nil, d.headers.AuthHeader(), d.client)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := req.Do()
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	var result struct {
		Data struct {
			Connection struct {
				GlucoseMeasurement struct {
					ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
					Value            float64 `json:"Value"`
					TrendArrow       int     `json:"TrendArrow"`
					MeasurementColor int     `json:"MeasurementColor"`
					GlucoseUnits     int     `json:"GlucoseUnits"`
					Timestamp        string  `json:"Timestamp"`
					IsHigh           bool    `json:"isHigh"`
					IsLow            bool    `json:"isLow"`
				} `json:"glucoseMeasurement"`
				Sensor struct {
					DeviceID string `json:"deviceId"`
					SN       string `json:"sn"`
					A        int    `json:"a"`
					W        int    `json:"w"`
					PT       int    `json:"pt"`
					S        bool   `json:"s"`
					LJ       bool   `json:"lj"`
				} `json:"sensor"`
			} `json:"connection"`
			GraphData []struct {
				FactoryTimestamp string  `json:"FactoryTimestamp"`
				Timestamp        string  `json:"Timestamp"`
				ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
				Value            float64 `json:"Value"`
				MeasurementColor int     `json:"MeasurementColor"`
				GlucoseUnits     int     `json:"GlucoseUnits"`
				IsHigh           bool    `json:"isHigh"`
				IsLow            bool    `json:"isLow"`
				Type             int     `json:"type"`
			} `json:"graphData"`
		} `json:"data"`
	}

	if err := json.Unmarshal(res, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Store current measurement
	current := result.Data.Connection.GlucoseMeasurement
	timestamp, err := timeparser.ParseLibreViewTimestamp(current.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse current timestamp: %w", err)
	}

	trendArrow := current.TrendArrow
	currentMeasurement := &glucosemeasurement.GlucoseMeasurement{
		FactoryTimestamp: timestamp, // Same as Timestamp for current
		Timestamp:        timestamp,
		Value:            current.Value,
		ValueInMgPerDl:   current.ValueInMgPerDl,
		TrendArrow:       &trendArrow,
		MeasurementColor: current.MeasurementColor,
		GlucoseUnits:     current.GlucoseUnits,
		IsHigh:           current.IsHigh,
		IsLow:            current.IsLow,
		Type:             1, // Current measurement
	}

	if err := d.storage.SaveMeasurement(currentMeasurement); err != nil {
		return fmt.Errorf("failed to save current measurement: %w", err)
	}

	// Store historical measurements
	for _, point := range result.Data.GraphData {
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

		if err := d.storage.SaveMeasurement(measurement); err != nil {
			return fmt.Errorf("failed to save historical measurement: %w", err)
		}
	}

	// Store sensor configuration
	sensor := result.Data.Connection.Sensor
	activationTime, err := timeparser.ParseLibreViewTimestamp(result.Data.Connection.GlucoseMeasurement.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse sensor activation: %w", err)
	}

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

	if err := d.storage.SaveSensor(sensorConfig); err != nil {
		return fmt.Errorf("failed to save sensor config: %w", err)
	}

	return nil
}

// fetch retrieves the latest glucose data from the LibreView API
// and stores it in the configured storage backend.
//
// This method fetches from /connections which returns only the current measurement.
// Used for periodic updates (every 5 minutes).
func (d *Daemon) fetch() error {
	req, err := httpreq.NewHttpReq("GET", config.ConnectionsURL, nil, d.headers.AuthHeader(), d.client)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := req.Do()
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	var result struct {
		Data []struct {
			GlucoseMeasurement struct {
				ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
				Value            float64 `json:"Value"`
				TrendArrow       int     `json:"TrendArrow"`
				TrendMessage     string  `json:"TrendMessage"`
				MeasurementColor int     `json:"MeasurementColor"`
				GlucoseUnits     int     `json:"GlucoseUnits"`
				Timestamp        string  `json:"Timestamp"`
				IsHigh           bool    `json:"isHigh"`
				IsLow            bool    `json:"isLow"`
			} `json:"glucoseMeasurement"`
		} `json:"data"`
	}

	if err := json.Unmarshal(res, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return fmt.Errorf("no data in response")
	}

	current := result.Data[0].GlucoseMeasurement
	timestamp, err := timeparser.ParseLibreViewTimestamp(current.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	trendArrow := current.TrendArrow
	var trendMessage *string
	if current.TrendMessage != "" {
		trendMessage = &current.TrendMessage
	}

	measurement := &glucosemeasurement.GlucoseMeasurement{
		FactoryTimestamp: timestamp,
		Timestamp:        timestamp,
		Value:            current.Value,
		ValueInMgPerDl:   current.ValueInMgPerDl,
		TrendArrow:       &trendArrow,
		TrendMessage:     trendMessage,
		MeasurementColor: current.MeasurementColor,
		GlucoseUnits:     current.GlucoseUnits,
		IsHigh:           current.IsHigh,
		IsLow:            current.IsLow,
		Type:             1,
	}

	if err := d.storage.SaveMeasurement(measurement); err != nil {
		return fmt.Errorf("failed to save measurement: %w", err)
	}

	return nil
}
