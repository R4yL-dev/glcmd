package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// DeviceRepositoryGORM is the GORM implementation of DeviceRepository.
// This is a singleton repository - only one device record is expected.
type DeviceRepositoryGORM struct {
	db *gorm.DB
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *gorm.DB) *DeviceRepositoryGORM {
	return &DeviceRepositoryGORM{db: db}
}

// Save creates or updates device info (singleton).
func (r *DeviceRepositoryGORM) Save(ctx context.Context, d *domain.DeviceInfo) error {
	db := txOrDefault(ctx, r.db)

	// Upsert: Update all fields on conflict with device_id
	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "device_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"device_type_id", "app_version", "alarms_enabled", "high_limit",
			"low_limit", "fixed_low_threshold", "last_update", "limit_enabled",
			"updated_at",
		}),
	}).Create(d)

	return result.Error
}

// Find returns the device info (only one record expected).
func (r *DeviceRepositoryGORM) Find(ctx context.Context) (*domain.DeviceInfo, error) {
	db := txOrDefault(ctx, r.db)

	var device domain.DeviceInfo
	result := db.First(&device)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, persistence.ErrNotFound
		}
		return nil, result.Error
	}

	return &device, nil
}
