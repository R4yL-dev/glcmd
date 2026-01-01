package storage

import (
	"time"

	"github.com/R4yL-dev/glcmd/internal/glucosemeasurement"
	"github.com/R4yL-dev/glcmd/internal/models"
)

// Storage defines the interface for persisting LibreView API data.
//
// Implementations must be thread-safe as they will be accessed concurrently
// by the daemon's fetch loop.
//
// Current implementation: MemoryStorage (in-memory, lost on restart)
// Future implementations: SQLite, PostgreSQL (via GORM)
type Storage interface {
	// Glucose Measurements
	SaveMeasurement(m *glucosemeasurement.GlucoseMeasurement) error
	GetLatestMeasurement() (*glucosemeasurement.GlucoseMeasurement, error)
	GetAllMeasurements() ([]*glucosemeasurement.GlucoseMeasurement, error)
	GetMeasurementsByTimeRange(start, end time.Time) ([]*glucosemeasurement.GlucoseMeasurement, error)

	// Sensor Configuration
	SaveSensor(s *models.SensorConfig) error
	GetActiveSensor() (*models.SensorConfig, error)
	GetAllSensors() ([]*models.SensorConfig, error)

	// User Preferences
	SaveUserPreferences(u *models.UserPreferences) error
	GetUserPreferences() (*models.UserPreferences, error)

	// Device Information
	SaveDeviceInfo(d *models.DeviceInfo) error
	GetDeviceInfo() (*models.DeviceInfo, error)

	// Glucose Targets
	SaveGlucoseTargets(t *models.GlucoseTargets) error
	GetGlucoseTargets() (*models.GlucoseTargets, error)
}
