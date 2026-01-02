package repository

import (
	"context"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
)

// MeasurementRepository defines the interface for glucose measurement persistence.
type MeasurementRepository interface {
	// Save creates or ignores a measurement (duplicate timestamps are silently ignored)
	Save(ctx context.Context, m *domain.GlucoseMeasurement) error

	// FindLatest returns the most recent measurement by timestamp
	FindLatest(ctx context.Context) (*domain.GlucoseMeasurement, error)

	// FindAll returns all measurements ordered by timestamp descending
	FindAll(ctx context.Context) ([]*domain.GlucoseMeasurement, error)

	// FindByTimeRange returns measurements within a time range (inclusive)
	FindByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.GlucoseMeasurement, error)
}

// SensorRepository defines the interface for sensor configuration persistence.
type SensorRepository interface {
	// Save creates or updates a sensor (upsert by serial number)
	Save(ctx context.Context, s *domain.SensorConfig) error

	// FindBySerialNumber returns a sensor by its serial number
	FindBySerialNumber(ctx context.Context, serial string) (*domain.SensorConfig, error)

	// FindActive returns the currently active sensor
	FindActive(ctx context.Context) (*domain.SensorConfig, error)

	// FindAll returns all sensors ordered by detected_at descending
	FindAll(ctx context.Context) ([]*domain.SensorConfig, error)

	// UpdateActiveStatus updates the active status of a sensor
	UpdateActiveStatus(ctx context.Context, serial string, active bool) error
}

// UserRepository defines the interface for user preferences persistence.
// This is a singleton repository - only one user record is expected.
type UserRepository interface {
	// Save creates or updates user preferences (singleton)
	Save(ctx context.Context, u *domain.UserPreferences) error

	// Find returns the user preferences (only one record expected)
	Find(ctx context.Context) (*domain.UserPreferences, error)
}

// DeviceRepository defines the interface for device information persistence.
// This is a singleton repository - only one device record is expected.
type DeviceRepository interface {
	// Save creates or updates device info (singleton)
	Save(ctx context.Context, d *domain.DeviceInfo) error

	// Find returns the device info (only one record expected)
	Find(ctx context.Context) (*domain.DeviceInfo, error)
}

// TargetsRepository defines the interface for glucose targets persistence.
// This is a singleton repository - only one targets record is expected.
type TargetsRepository interface {
	// Save creates or updates glucose targets (singleton)
	Save(ctx context.Context, t *domain.GlucoseTargets) error

	// Find returns the glucose targets (only one record expected)
	Find(ctx context.Context) (*domain.GlucoseTargets, error)
}
