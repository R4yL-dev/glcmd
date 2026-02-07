package service

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
	"github.com/R4yL-dev/glcmd/internal/repository"
)

// MeasurementStats contains aggregated statistics for measurements
type MeasurementStats struct {
	Count          int        `json:"count"`
	Average        float64    `json:"average"`
	AverageMgDl    float64    `json:"averageMgDl"`
	Min            float64    `json:"min"`
	MinMgDl        int        `json:"minMgDl"`
	Max            float64    `json:"max"`
	MaxMgDl        int        `json:"maxMgDl"`
	StdDev         float64    `json:"stdDev"`
	LowCount       int        `json:"lowCount"`
	NormalCount    int        `json:"normalCount"`
	HighCount      int        `json:"highCount"`
	TimeInRange    float64    `json:"timeInRange"`
	TimeBelowRange float64    `json:"timeBelowRange"`
	TimeAboveRange float64    `json:"timeAboveRange"`
	GMI            *float64   `json:"gmi,omitempty"`
	FirstTimestamp *time.Time `json:"-"` // Oldest measurement (not in JSON, used for period)
	LastTimestamp  *time.Time `json:"-"` // Newest measurement (not in JSON, used for period)
}

// GlucoseServiceImpl implements GlucoseService.
type GlucoseServiceImpl struct {
	repo   repository.MeasurementRepository
	retry  *persistence.RetryConfig
	logger *slog.Logger
}

// NewGlucoseService creates a new GlucoseService.
func NewGlucoseService(
	repo repository.MeasurementRepository,
	logger *slog.Logger,
) *GlucoseServiceImpl {
	return &GlucoseServiceImpl{
		repo:   repo,
		retry:  persistence.DefaultRetryConfig(),
		logger: logger,
	}
}

// SaveMeasurement saves a glucose measurement with retry logic.
// Returns (true, nil) if inserted, (false, nil) if duplicate was ignored.
func (s *GlucoseServiceImpl) SaveMeasurement(ctx context.Context, m *domain.GlucoseMeasurement) (bool, error) {
	start := time.Now()
	var inserted bool

	// Execute with retry on retryable errors (database locks, etc.)
	err := persistence.ExecuteWithRetry(ctx, s.retry, func() error {
		var saveErr error
		inserted, saveErr = s.repo.Save(ctx, m)
		return saveErr
	})

	duration := time.Since(start)
	if err != nil {
		return false, err
	}

	s.logger.Debug("measurement saved",
		"timestamp", m.Timestamp,
		"value", m.Value,
		"inserted", inserted,
		"duration", duration,
	)

	return inserted, nil
}

// GetLatestMeasurement returns the most recent measurement.
func (s *GlucoseServiceImpl) GetLatestMeasurement(ctx context.Context) (*domain.GlucoseMeasurement, error) {
	return s.repo.FindLatest(ctx)
}

// GetAllMeasurements returns all measurements.
func (s *GlucoseServiceImpl) GetAllMeasurements(ctx context.Context) ([]*domain.GlucoseMeasurement, error) {
	return s.repo.FindAll(ctx)
}

// GetMeasurementsByTimeRange returns measurements within a time range.
func (s *GlucoseServiceImpl) GetMeasurementsByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.GlucoseMeasurement, error) {
	return s.repo.FindByTimeRange(ctx, start, end)
}

// GetMeasurementsWithFilters returns filtered and paginated measurements with total count.
func (s *GlucoseServiceImpl) GetMeasurementsWithFilters(ctx context.Context, filters repository.MeasurementFilters, limit, offset int) ([]*domain.GlucoseMeasurement, int64, error) {
	// Get measurements
	measurements, err := s.repo.FindWithFilters(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := s.repo.CountWithFilters(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return measurements, total, nil
}

// GetStatistics calculates aggregated statistics for a time range.
// If start and end are nil, returns statistics for all data (all time).
func (s *GlucoseServiceImpl) GetStatistics(ctx context.Context, start, end *time.Time, targets *domain.GlucoseTargets) (*MeasurementStats, error) {
	filters := repository.StatisticsFilters{
		StartTime: start,
		EndTime:   end,
	}

	if targets != nil {
		filters.TargetLowMgDl = &targets.TargetLow
		filters.TargetHighMgDl = &targets.TargetHigh
	}

	result, err := s.repo.GetStatistics(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Map StatisticsResult to MeasurementStats
	// Compute stddev from variance (sqrt computed in Go for SQLite compatibility)
	stats := &MeasurementStats{
		Count:          int(result.Count),
		Average:        result.Average,
		AverageMgDl:    result.AverageMgDl,
		Min:            result.Min,
		MinMgDl:        result.MinMgDl,
		Max:            result.Max,
		MaxMgDl:        result.MaxMgDl,
		StdDev:         math.Sqrt(result.Variance),
		LowCount:       int(result.LowCount),
		NormalCount:    int(result.NormalCount),
		HighCount:      int(result.HighCount),
		FirstTimestamp: result.FirstTimestamp,
		LastTimestamp:  result.LastTimestamp,
	}

	stats.GMI = domain.CalculateGMI(stats.AverageMgDl)

	// Calculate Time in Range percentages if targets were provided
	if result.Count > 0 && targets != nil {
		total := float64(result.Count)
		stats.TimeInRange = (float64(result.InRangeCount) / total) * 100
		stats.TimeBelowRange = (float64(result.BelowRangeCount) / total) * 100
		stats.TimeAboveRange = (float64(result.AboveRangeCount) / total) * 100
	}

	return stats, nil
}
