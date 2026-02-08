package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
	"github.com/R4yL-dev/glcmd/internal/repository"
)

// SensorServiceImpl implements SensorService.
type SensorServiceImpl struct {
	repo   repository.SensorRepository
	uow    repository.UnitOfWork
	logger *slog.Logger
}

// NewSensorService creates a new SensorService.
func NewSensorService(
	repo repository.SensorRepository,
	uow repository.UnitOfWork,
	logger *slog.Logger,
) *SensorServiceImpl {
	return &SensorServiceImpl{
		repo:   repo,
		uow:    uow,
		logger: logger,
	}
}

// SaveSensor saves a sensor configuration.
func (s *SensorServiceImpl) SaveSensor(ctx context.Context, sensor *domain.SensorConfig) error {
	return s.repo.Save(ctx, sensor)
}

// GetCurrentSensor returns the current sensor (not ended).
func (s *SensorServiceImpl) GetCurrentSensor(ctx context.Context) (*domain.SensorConfig, error) {
	return s.repo.FindCurrent(ctx)
}

// GetAllSensors returns all sensors.
func (s *SensorServiceImpl) GetAllSensors(ctx context.Context) ([]*domain.SensorConfig, error) {
	return s.repo.FindAll(ctx)
}

// HandleSensorChange handles sensor change detection.
//
// This method implements the business logic for sensor changes:
// 1. Check for existing current sensor
// 2. If serial number changed, set EndedAt on old sensor
// 3. Save new sensor
//
// All operations are executed within a transaction to ensure atomicity.
func (s *SensorServiceImpl) HandleSensorChange(ctx context.Context, newSensor *domain.SensorConfig) error {
	return s.uow.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Check for existing current sensor
		currentSensor, err := s.repo.FindCurrent(txCtx)
		if err != nil && !errors.Is(err, persistence.ErrNotFound) {
			return fmt.Errorf("failed to find current sensor: %w", err)
		}

		// 2. If sensor changed, mark old one as ended
		if currentSensor != nil && currentSensor.SerialNumber != newSensor.SerialNumber {
			// Use LastMeasurementAt if available for more accurate EndedAt
			var endedAt time.Time
			if currentSensor.LastMeasurementAt != nil {
				endedAt = *currentSensor.LastMeasurementAt
			} else {
				endedAt = time.Now().UTC()
			}

			s.logger.Info("sensor change detected",
				"oldSerial", currentSensor.SerialNumber,
				"newSerial", newSensor.SerialNumber,
				"oldActivation", currentSensor.Activation,
				"newActivation", newSensor.Activation,
				"oldEndedAt", endedAt,
			)

			err = s.repo.SetEndedAt(txCtx, currentSensor.SerialNumber, endedAt)
			if err != nil {
				return fmt.Errorf("failed to set ended_at on old sensor: %w", err)
			}

			// Calculate actual days the old sensor was used
			actualDays := endedAt.Sub(currentSensor.Activation).Hours() / 24

			s.logger.Info("old sensor ended",
				"serialNumber", currentSensor.SerialNumber,
				"actualDays", fmt.Sprintf("%.1f", actualDays),
				"expectedDays", currentSensor.DurationDays,
			)
		}

		// 3. Save new sensor
		if err := s.repo.Save(txCtx, newSensor); err != nil {
			return fmt.Errorf("failed to save sensor: %w", err)
		}

		// Log only if it's a new sensor (not just an update)
		if currentSensor == nil || currentSensor.SerialNumber != newSensor.SerialNumber {
			s.logger.Info("new sensor detected",
				"serialNumber", newSensor.SerialNumber,
				"activation", newSensor.Activation,
				"expiresAt", newSensor.ExpiresAt,
				"durationDays", newSensor.DurationDays,
			)
		}

		return nil
	})
}

// SensorStats contains aggregated sensor lifecycle statistics
type SensorStats struct {
	TotalSensors  int     `json:"totalSensors"`
	EndedSensors  int     `json:"endedSensors"`
	AvgDuration   float64 `json:"avgDuration"`   // days
	MinDuration   float64 `json:"minDuration"`
	MaxDuration   float64 `json:"maxDuration"`
	AvgExpected   float64 `json:"avgExpected"`
	AvgDifference float64 `json:"avgDifference"` // avg_duration - avg_expected
}

// GetSensorsWithFilters returns filtered and paginated sensors with total count.
func (s *SensorServiceImpl) GetSensorsWithFilters(ctx context.Context, filters repository.SensorFilters, limit, offset int) ([]*domain.SensorConfig, int64, error) {
	sensors, err := s.repo.FindWithFilters(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountWithFilters(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return sensors, total, nil
}

// GetStatistics returns aggregated sensor lifecycle statistics.
func (s *SensorServiceImpl) GetStatistics(ctx context.Context, start, end *time.Time) (*SensorStats, error) {
	filters := repository.SensorStatisticsFilters{
		StartTime: start,
		EndTime:   end,
	}
	result, err := s.repo.GetStatistics(ctx, filters)
	if err != nil {
		return nil, err
	}

	stats := &SensorStats{
		TotalSensors:  int(result.TotalSensors),
		EndedSensors:  int(result.EndedSensors),
		AvgDuration:   result.AvgDuration,
		MinDuration:   result.MinDuration,
		MaxDuration:   result.MaxDuration,
		AvgExpected:   result.AvgExpected,
		AvgDifference: result.AvgDuration - result.AvgExpected,
	}

	return stats, nil
}

// UpdateLastMeasurementIfNewer updates the LastMeasurementAt field of the current sensor
// only if the provided timestamp is newer than the existing one.
// This handles historical measurements that may arrive out of order.
func (s *SensorServiceImpl) UpdateLastMeasurementIfNewer(ctx context.Context, timestamp time.Time) error {
	current, err := s.repo.FindCurrent(ctx)
	if err != nil {
		// No current sensor = nothing to update
		return nil
	}

	// Update only if the timestamp is newer than the existing one
	if current.LastMeasurementAt == nil || timestamp.After(*current.LastMeasurementAt) {
		current.LastMeasurementAt = &timestamp
		return s.repo.Save(ctx, current)
	}

	return nil // Nothing to do, the existing timestamp is more recent
}
