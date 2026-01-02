package service

import (
	"context"
	"log/slog"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/repository"
)

// ConfigServiceImpl implements ConfigService.
type ConfigServiceImpl struct {
	userRepo    repository.UserRepository
	deviceRepo  repository.DeviceRepository
	targetsRepo repository.TargetsRepository
	logger      *slog.Logger
}

// NewConfigService creates a new ConfigService.
func NewConfigService(
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceRepository,
	targetsRepo repository.TargetsRepository,
	logger *slog.Logger,
) *ConfigServiceImpl {
	return &ConfigServiceImpl{
		userRepo:    userRepo,
		deviceRepo:  deviceRepo,
		targetsRepo: targetsRepo,
		logger:      logger,
	}
}

// SaveUserPreferences saves user preferences.
func (s *ConfigServiceImpl) SaveUserPreferences(ctx context.Context, u *domain.UserPreferences) error {
	if err := s.userRepo.Save(ctx, u); err != nil {
		s.logger.Error("failed to save user preferences", "error", err)
		return err
	}

	s.logger.Debug("user preferences saved", "userId", u.UserID)
	return nil
}

// GetUserPreferences returns user preferences.
func (s *ConfigServiceImpl) GetUserPreferences(ctx context.Context) (*domain.UserPreferences, error) {
	return s.userRepo.Find(ctx)
}

// SaveDeviceInfo saves device information.
func (s *ConfigServiceImpl) SaveDeviceInfo(ctx context.Context, d *domain.DeviceInfo) error {
	if err := s.deviceRepo.Save(ctx, d); err != nil {
		s.logger.Error("failed to save device info", "error", err)
		return err
	}

	s.logger.Debug("device info saved", "deviceId", d.DeviceID)
	return nil
}

// GetDeviceInfo returns device information.
func (s *ConfigServiceImpl) GetDeviceInfo(ctx context.Context) (*domain.DeviceInfo, error) {
	return s.deviceRepo.Find(ctx)
}

// SaveGlucoseTargets saves glucose targets.
func (s *ConfigServiceImpl) SaveGlucoseTargets(ctx context.Context, t *domain.GlucoseTargets) error {
	if err := s.targetsRepo.Save(ctx, t); err != nil {
		s.logger.Error("failed to save glucose targets", "error", err)
		return err
	}

	s.logger.Debug("glucose targets saved",
		"targetHigh", t.TargetHigh,
		"targetLow", t.TargetLow,
	)
	return nil
}

// GetGlucoseTargets returns glucose targets.
func (s *ConfigServiceImpl) GetGlucoseTargets(ctx context.Context) (*domain.GlucoseTargets, error) {
	return s.targetsRepo.Find(ctx)
}
