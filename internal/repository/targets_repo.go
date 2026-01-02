package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// TargetsRepositoryGORM is the GORM implementation of TargetsRepository.
// This is a singleton repository - only one targets record is expected.
type TargetsRepositoryGORM struct {
	db *gorm.DB
}

// NewTargetsRepository creates a new TargetsRepository.
func NewTargetsRepository(db *gorm.DB) *TargetsRepositoryGORM {
	return &TargetsRepositoryGORM{db: db}
}

// Save creates or updates glucose targets (singleton).
func (r *TargetsRepositoryGORM) Save(ctx context.Context, t *domain.GlucoseTargets) error {
	db := txOrDefault(ctx, r.db)

	// For singleton, we always update the first record or create if doesn't exist
	var existing domain.GlucoseTargets
	result := db.First(&existing)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// No record exists, create new one
			return db.Create(t).Error
		}
		return result.Error
	}

	// Record exists, update it
	t.ID = existing.ID // Preserve the ID
	return db.Save(t).Error
}

// Find returns the glucose targets (only one record expected).
func (r *TargetsRepositoryGORM) Find(ctx context.Context) (*domain.GlucoseTargets, error) {
	db := txOrDefault(ctx, r.db)

	var targets domain.GlucoseTargets
	result := db.First(&targets)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, persistence.ErrNotFound
		}
		return nil, result.Error
	}

	return &targets, nil
}
