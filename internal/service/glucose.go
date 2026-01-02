package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
	"github.com/R4yL-dev/glcmd/internal/repository"
)

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
