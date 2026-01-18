package service

import (
	"context"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/repository"
)

// GlucoseService defines the interface for glucose measurement business logic.
type GlucoseService interface {
	// SaveMeasurement saves a glucose measurement with retry logic
	SaveMeasurement(ctx context.Context, m *domain.GlucoseMeasurement) error

	// GetLatestMeasurement returns the most recent measurement
	GetLatestMeasurement(ctx context.Context) (*domain.GlucoseMeasurement, error)

	// GetAllMeasurements returns all measurements
	GetAllMeasurements(ctx context.Context) ([]*domain.GlucoseMeasurement, error)

	// GetMeasurementsByTimeRange returns measurements within a time range
	GetMeasurementsByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.GlucoseMeasurement, error)

	// GetMeasurementsWithFilters returns filtered and paginated measurements with total count
	GetMeasurementsWithFilters(ctx context.Context, filters repository.MeasurementFilters, limit, offset int) ([]*domain.GlucoseMeasurement, int64, error)

	// GetStatistics calculates aggregated statistics for a time range.
	// If start and end are nil, returns statistics for all data (all time).
	GetStatistics(ctx context.Context, start, end *time.Time, targets *domain.GlucoseTargets) (*MeasurementStats, error)
}

// SensorService defines the interface for sensor management business logic.
type SensorService interface {
	// SaveSensor saves a sensor configuration
	SaveSensor(ctx context.Context, s *domain.SensorConfig) error

	// GetCurrentSensor returns the current sensor (not ended)
	GetCurrentSensor(ctx context.Context) (*domain.SensorConfig, error)

	// GetAllSensors returns all sensors
	GetAllSensors(ctx context.Context) ([]*domain.SensorConfig, error)

	// HandleSensorChange handles sensor change detection.
	// This method uses a transaction to ensure atomicity:
	// 1. Check for existing current sensor
	// 2. If serial number changed, set EndedAt on old sensor
	// 3. Save new sensor
	HandleSensorChange(ctx context.Context, newSensor *domain.SensorConfig) error
}

// ConfigService defines the interface for configuration management (user, device, targets).
type ConfigService interface {
	// SaveUserPreferences saves user preferences
	SaveUserPreferences(ctx context.Context, u *domain.UserPreferences) error

	// GetUserPreferences returns user preferences
	GetUserPreferences(ctx context.Context) (*domain.UserPreferences, error)

	// SaveDeviceInfo saves device information
	SaveDeviceInfo(ctx context.Context, d *domain.DeviceInfo) error

	// GetDeviceInfo returns device information
	GetDeviceInfo(ctx context.Context) (*domain.DeviceInfo, error)

	// SaveGlucoseTargets saves glucose targets
	SaveGlucoseTargets(ctx context.Context, t *domain.GlucoseTargets) error

	// GetGlucoseTargets returns glucose targets
	GetGlucoseTargets(ctx context.Context) (*domain.GlucoseTargets, error)
}
