package memory

import (
	"errors"
	"sync"
	"time"

	"github.com/R4yL-dev/glcmd/internal/glucosemeasurement"
	"github.com/R4yL-dev/glcmd/internal/models"
	"github.com/R4yL-dev/glcmd/internal/storage"
)

// MemoryStorage is an in-memory implementation of the Storage interface.
//
// All data is stored in memory and will be lost when the daemon restarts.
// This implementation is thread-safe using sync.RWMutex for concurrent access.
//
// Used as the initial storage layer before implementing database persistence.
type MemoryStorage struct {
	mu              sync.RWMutex
	measurements    []*glucosemeasurement.GlucoseMeasurement
	sensors         []*models.SensorConfig
	userPreferences *models.UserPreferences
	deviceInfo      *models.DeviceInfo
	glucoseTargets  *models.GlucoseTargets
}

// Ensure MemoryStorage implements storage.Storage interface
var _ storage.Storage = (*MemoryStorage)(nil)

// New creates a new MemoryStorage instance.
func New() *MemoryStorage {
	return &MemoryStorage{
		measurements: make([]*glucosemeasurement.GlucoseMeasurement, 0),
		sensors:      make([]*models.SensorConfig, 0),
	}
}

// SaveMeasurement stores a glucose measurement in memory.
// Measurements are stored in chronological order.
// Duplicates (same Timestamp) are ignored to prevent storing the same measurement twice.
func (m *MemoryStorage) SaveMeasurement(measurement *glucosemeasurement.GlucoseMeasurement) error {
	if measurement == nil {
		return errors.New("measurement cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicates by Timestamp
	// This prevents storing the same measurement from /connections and /graph
	for _, existing := range m.measurements {
		if existing.Timestamp.Equal(measurement.Timestamp) {
			// Measurement already exists, skip
			return nil
		}
	}

	m.measurements = append(m.measurements, measurement)
	return nil
}

// GetLatestMeasurement returns the most recent glucose measurement by timestamp.
// Returns an error if no measurements exist.
func (m *MemoryStorage) GetLatestMeasurement() (*glucosemeasurement.GlucoseMeasurement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.measurements) == 0 {
		return nil, errors.New("no measurements available")
	}

	// Find the measurement with the most recent timestamp
	latest := m.measurements[0]
	for _, measurement := range m.measurements[1:] {
		if measurement.Timestamp.After(latest.Timestamp) {
			latest = measurement
		}
	}

	return latest, nil
}

// GetAllMeasurements returns all stored glucose measurements.
// Returns a copy of the slice to prevent external modifications.
func (m *MemoryStorage) GetAllMeasurements() ([]*glucosemeasurement.GlucoseMeasurement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make([]*glucosemeasurement.GlucoseMeasurement, len(m.measurements))
	copy(result, m.measurements)
	return result, nil
}

// GetMeasurementsByTimeRange returns measurements within the specified time range.
// The range is inclusive (start <= timestamp <= end).
func (m *MemoryStorage) GetMeasurementsByTimeRange(start, end time.Time) ([]*glucosemeasurement.GlucoseMeasurement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*glucosemeasurement.GlucoseMeasurement, 0)

	for _, measurement := range m.measurements {
		if !measurement.Timestamp.Before(start) && !measurement.Timestamp.After(end) {
			result = append(result, measurement)
		}
	}

	return result, nil
}

// SaveSensor stores sensor configuration in memory.
// If a sensor with the same serial number exists, it's updated.
func (m *MemoryStorage) SaveSensor(sensor *models.SensorConfig) error {
	if sensor == nil {
		return errors.New("sensor cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if sensor already exists
	for i, s := range m.sensors {
		if s.SerialNumber == sensor.SerialNumber {
			// Update existing sensor
			m.sensors[i] = sensor
			return nil
		}
	}

	// Add new sensor
	m.sensors = append(m.sensors, sensor)
	return nil
}

// GetActiveSensor returns the currently active sensor.
// Returns the sensor with IsActive=true, or the most recent one if none is marked active.
func (m *MemoryStorage) GetActiveSensor() (*models.SensorConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.sensors) == 0 {
		return nil, errors.New("no sensors available")
	}

	// Look for active sensor
	for _, sensor := range m.sensors {
		if sensor.IsActive {
			return sensor, nil
		}
	}

	// If no active sensor, return the most recent one
	return m.sensors[len(m.sensors)-1], nil
}

// GetAllSensors returns all stored sensor configurations.
// Returns a copy of the slice to prevent external modifications.
func (m *MemoryStorage) GetAllSensors() ([]*models.SensorConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*models.SensorConfig, len(m.sensors))
	copy(result, m.sensors)
	return result, nil
}

// SaveUserPreferences stores user preferences in memory.
// Replaces any existing user preferences.
func (m *MemoryStorage) SaveUserPreferences(user *models.UserPreferences) error {
	if user == nil {
		return errors.New("user preferences cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.userPreferences = user
	return nil
}

// GetUserPreferences returns the stored user preferences.
// Returns an error if no preferences have been saved.
func (m *MemoryStorage) GetUserPreferences() (*models.UserPreferences, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.userPreferences == nil {
		return nil, errors.New("no user preferences available")
	}

	return m.userPreferences, nil
}

// SaveDeviceInfo stores device information in memory.
// Replaces any existing device information.
func (m *MemoryStorage) SaveDeviceInfo(device *models.DeviceInfo) error {
	if device == nil {
		return errors.New("device info cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.deviceInfo = device
	return nil
}

// GetDeviceInfo returns the stored device information.
// Returns an error if no device info has been saved.
func (m *MemoryStorage) GetDeviceInfo() (*models.DeviceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.deviceInfo == nil {
		return nil, errors.New("no device info available")
	}

	return m.deviceInfo, nil
}

// SaveGlucoseTargets stores glucose targets in memory.
// Replaces any existing targets.
func (m *MemoryStorage) SaveGlucoseTargets(targets *models.GlucoseTargets) error {
	if targets == nil {
		return errors.New("glucose targets cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.glucoseTargets = targets
	return nil
}

// GetGlucoseTargets returns the stored glucose targets.
// Returns an error if no targets have been saved.
func (m *MemoryStorage) GetGlucoseTargets() (*models.GlucoseTargets, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.glucoseTargets == nil {
		return nil, errors.New("no glucose targets available")
	}

	return m.glucoseTargets, nil
}
