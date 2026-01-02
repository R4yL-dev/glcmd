package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// SensorRepositoryGORM is the GORM implementation of SensorRepository.
type SensorRepositoryGORM struct {
	db *gorm.DB
}

// NewSensorRepository creates a new SensorRepository.
func NewSensorRepository(db *gorm.DB) *SensorRepositoryGORM {
	return &SensorRepositoryGORM{db: db}
}

// Save creates or updates a sensor (upsert by serial number).
func (r *SensorRepositoryGORM) Save(ctx context.Context, s *domain.SensorConfig) error {
	db := txOrDefault(ctx, r.db)

	// Upsert: Update all fields on conflict with serial_number
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "serial_number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"activation", "device_id", "sensor_type", "warranty_days",
			"is_active", "low_journey", "detected_at", "updated_at",
		}),
	}).Create(s)

	return result.Error
}

// FindBySerialNumber returns a sensor by its serial number.
func (r *SensorRepositoryGORM) FindBySerialNumber(ctx context.Context, serial string) (*domain.SensorConfig, error) {
	db := txOrDefault(ctx, r.db)

	var sensor domain.SensorConfig
	result := db.Where("serial_number = ?", serial).First(&sensor)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, persistence.ErrNotFound
		}
		return nil, result.Error
	}

	return &sensor, nil
}

// FindActive returns the currently active sensor.
func (r *SensorRepositoryGORM) FindActive(ctx context.Context) (*domain.SensorConfig, error) {
	db := txOrDefault(ctx, r.db)

	var sensor domain.SensorConfig
	result := db.Where("is_active = ?", true).First(&sensor)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, persistence.ErrNotFound
		}
		return nil, result.Error
	}

	return &sensor, nil
}

// FindAll returns all sensors ordered by detected_at descending.
func (r *SensorRepositoryGORM) FindAll(ctx context.Context) ([]*domain.SensorConfig, error) {
	db := txOrDefault(ctx, r.db)

	var sensors []*domain.SensorConfig
	result := db.Order("detected_at DESC").Find(&sensors)

	if result.Error != nil {
		return nil, result.Error
	}

	return sensors, nil
}

// UpdateActiveStatus updates the active status of a sensor.
func (r *SensorRepositoryGORM) UpdateActiveStatus(ctx context.Context, serial string, active bool) error {
	db := txOrDefault(ctx, r.db)

	result := db.Model(&domain.SensorConfig{}).
		Where("serial_number = ?", serial).
		Update("is_active", active)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return persistence.ErrNotFound
	}

	return nil
}
