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
	Count         int     `json:"count"`
	Average       float64 `json:"average"`
	AverageMgDl   float64 `json:"averageMgDl"`
	Min           float64 `json:"min"`
	MinMgDl       int     `json:"minMgDl"`
	Max           float64 `json:"max"`
	MaxMgDl       int     `json:"maxMgDl"`
	StdDev        float64 `json:"stdDev"`
	LowCount      int     `json:"lowCount"`
	NormalCount   int     `json:"normalCount"`
	HighCount     int     `json:"highCount"`
	TimeInRange   float64 `json:"timeInRange"`
	TimeBelowRange float64 `json:"timeBelowRange"`
	TimeAboveRange float64 `json:"timeAboveRange"`
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
func (s *GlucoseServiceImpl) SaveMeasurement(ctx context.Context, m *domain.GlucoseMeasurement) error {
	start := time.Now()

	// Execute with retry on retryable errors (database locks, etc.)
	err := persistence.ExecuteWithRetry(ctx, s.retry, func() error {
		return s.repo.Save(ctx, m)
	})

	// Log performance metrics
	duration := time.Since(start)
	if err != nil {
		s.logger.Error("failed to save measurement",
			"error", err,
			"duration", duration,
			"timestamp", m.Timestamp,
		)
		return err
	}

	s.logger.Debug("measurement saved",
		"timestamp", m.Timestamp,
		"value", m.Value,
		"duration", duration,
	)

	return nil
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
func (s *GlucoseServiceImpl) GetStatistics(ctx context.Context, start, end time.Time, targets *domain.GlucoseTargets) (*MeasurementStats, error) {
	// Get measurements in the time range
	measurements, err := s.repo.FindByTimeRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	if len(measurements) == 0 {
		// Return empty stats if no measurements
		return &MeasurementStats{
			Count: 0,
		}, nil
	}

	// Calculate basic statistics
	stats := &MeasurementStats{
		Count: len(measurements),
	}

	var sum float64
	var sumMgDl float64
	stats.Min = measurements[0].Value
	stats.MinMgDl = measurements[0].ValueInMgPerDl
	stats.Max = measurements[0].Value
	stats.MaxMgDl = measurements[0].ValueInMgPerDl

	for _, m := range measurements {
		sum += m.Value
		sumMgDl += float64(m.ValueInMgPerDl)

		if m.Value < stats.Min {
			stats.Min = m.Value
			stats.MinMgDl = m.ValueInMgPerDl
		}
		if m.Value > stats.Max {
			stats.Max = m.Value
			stats.MaxMgDl = m.ValueInMgPerDl
		}

		// Count by color
		switch m.MeasurementColor {
		case 1:
			stats.NormalCount++
		case 2:
			// Warning can be either low or high
			if m.IsLow {
				stats.LowCount++
			} else {
				stats.HighCount++
			}
		case 3:
			// Critical can be either low or high
			if m.IsLow {
				stats.LowCount++
			} else {
				stats.HighCount++
			}
		}
	}

	stats.Average = sum / float64(len(measurements))
	stats.AverageMgDl = sumMgDl / float64(len(measurements))

	// Calculate standard deviation
	var sumSquares float64
	for _, m := range measurements {
		diff := m.Value - stats.Average
		sumSquares += diff * diff
	}
	stats.StdDev = math.Sqrt(sumSquares / float64(len(measurements)))

	// Calculate Time in Range if targets are provided
	if targets != nil {
		var inRange, below, above int
		targetLowMgDl := targets.TargetLow
		targetHighMgDl := targets.TargetHigh

		for _, m := range measurements {
			if m.ValueInMgPerDl < targetLowMgDl {
				below++
			} else if m.ValueInMgPerDl > targetHighMgDl {
				above++
			} else {
				inRange++
			}
		}

		total := float64(len(measurements))
		stats.TimeInRange = (float64(inRange) / total) * 100
		stats.TimeBelowRange = (float64(below) / total) * 100
		stats.TimeAboveRange = (float64(above) / total) * 100
	}

	return stats, nil
}
