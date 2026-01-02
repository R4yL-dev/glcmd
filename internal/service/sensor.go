package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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

// GetActiveSensor returns the currently active sensor.
func (s *SensorServiceImpl) GetActiveSensor(ctx context.Context) (*domain.SensorConfig, error) {
	return s.repo.FindActive(ctx)
}

// GetAllSensors returns all sensors.
func (s *SensorServiceImpl) GetAllSensors(ctx context.Context) ([]*domain.SensorConfig, error) {
	return s.repo.FindAll(ctx)
}

// HandleSensorChange handles sensor change detection and deactivation of old sensor.
//
// This method implements the business logic for sensor changes:
// 1. Check for existing active sensor
// 2. If serial number changed, deactivate old sensor and log the change
// 3. Save new sensor as active
//
// All operations are executed within a transaction to ensure atomicity.
func (s *SensorServiceImpl) HandleSensorChange(ctx context.Context, newSensor *domain.SensorConfig) error {
	return s.uow.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Check for existing active sensor
		currentSensor, err := s.repo.FindActive(txCtx)
		if err != nil && !errors.Is(err, persistence.ErrNotFound) {
			return fmt.Errorf("failed to find active sensor: %w", err)
		}

		// 2. If sensor changed, deactivate old one
		if currentSensor != nil && currentSensor.SerialNumber != newSensor.SerialNumber {
			s.logger.Info("sensor change detected",
				"oldSerial", currentSensor.SerialNumber,
				"newSerial", newSensor.SerialNumber,
				"oldActivation", currentSensor.Activation,
				"newActivation", newSensor.Activation,
			)

			err = s.repo.UpdateActiveStatus(txCtx, currentSensor.SerialNumber, false)
			if err != nil {
				return fmt.Errorf("failed to deactivate old sensor: %w", err)
			}

			s.logger.Info("old sensor deactivated",
				"serialNumber", currentSensor.SerialNumber,
			)
		}

		// 3. Save new sensor (will be active)
		if err := s.repo.Save(txCtx, newSensor); err != nil {
			return fmt.Errorf("failed to save new sensor: %w", err)
		}

		if currentSensor == nil || currentSensor.SerialNumber != newSensor.SerialNumber {
			s.logger.Info("new sensor activated",
				"serialNumber", newSensor.SerialNumber,
				"activation", newSensor.Activation,
			)
		}

		return nil
	})
}
