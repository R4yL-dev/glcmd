package repository

import (
	"context"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
)

// MeasurementFilters defines filter criteria for querying measurements
type MeasurementFilters struct {
	StartTime *time.Time
	EndTime   *time.Time
	Color     *int // 1=normal, 2=warning, 3=critical
	Type      *int // 0=historical, 1=current
}

// StatisticsFilters defines filter criteria for aggregated statistics
type StatisticsFilters struct {
	StartTime      *time.Time // nil = no lower bound
	EndTime        *time.Time // nil = no upper bound
	TargetLowMgDl  *int       // For Time in Range calculation
	TargetHighMgDl *int       // For Time in Range calculation
}

// StatisticsResult contains aggregated statistics computed by SQL
type StatisticsResult struct {
	Count           int64
	Average         float64
	AverageMgDl     float64
	Min             float64
	MinMgDl         int
	Max             float64
	MaxMgDl         int
	Variance        float64 // variance = E[X²] - E[X]², sqrt computed in Go for SQLite compatibility
	LowCount        int64
	NormalCount     int64
	HighCount       int64
	InRangeCount    int64
	BelowRangeCount int64
	AboveRangeCount int64
	FirstTimestamp  *time.Time // Oldest measurement timestamp
	LastTimestamp   *time.Time // Newest measurement timestamp
}

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

	// FindWithFilters returns measurements matching filters with pagination
	FindWithFilters(ctx context.Context, filters MeasurementFilters, limit, offset int) ([]*domain.GlucoseMeasurement, error)

	// CountWithFilters returns total count of measurements matching filters
	CountWithFilters(ctx context.Context, filters MeasurementFilters) (int64, error)

	// GetStatistics returns aggregated statistics computed by SQL
	GetStatistics(ctx context.Context, filters StatisticsFilters) (*StatisticsResult, error)
}

// SensorFilters defines filter criteria for querying sensors
type SensorFilters struct {
	StartTime *time.Time // filter on activation
	EndTime   *time.Time
}

// SensorStatisticsResult contains aggregated sensor statistics computed by SQL
type SensorStatisticsResult struct {
	TotalSensors int64
	EndedSensors int64
	AvgDuration  float64 // average days of use (ended sensors)
	MinDuration  float64
	MaxDuration  float64
	AvgExpected  float64 // average expected days
}

// SensorRepository defines the interface for sensor configuration persistence.
type SensorRepository interface {
	// Save creates or updates a sensor (upsert by serial number)
	Save(ctx context.Context, s *domain.SensorConfig) error

	// FindBySerialNumber returns a sensor by its serial number
	FindBySerialNumber(ctx context.Context, serial string) (*domain.SensorConfig, error)

	// FindCurrent returns the current sensor (EndedAt is null)
	FindCurrent(ctx context.Context) (*domain.SensorConfig, error)

	// FindAll returns all sensors ordered by detected_at descending
	FindAll(ctx context.Context) ([]*domain.SensorConfig, error)

	// FindWithFilters returns sensors matching filters with pagination
	FindWithFilters(ctx context.Context, filters SensorFilters, limit, offset int) ([]*domain.SensorConfig, error)

	// CountWithFilters returns total count of sensors matching filters
	CountWithFilters(ctx context.Context, filters SensorFilters) (int64, error)

	// GetStatistics returns aggregated sensor lifecycle statistics computed by SQL
	GetStatistics(ctx context.Context) (*SensorStatisticsResult, error)

	// SetEndedAt marks a sensor as ended (replaced by a new sensor)
	SetEndedAt(ctx context.Context, serial string, endedAt time.Time) error
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
