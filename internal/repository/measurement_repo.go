package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// MeasurementRepositoryGORM is the GORM implementation of MeasurementRepository.
type MeasurementRepositoryGORM struct {
	db *gorm.DB
}

// NewMeasurementRepository creates a new MeasurementRepository.
func NewMeasurementRepository(db *gorm.DB) *MeasurementRepositoryGORM {
	return &MeasurementRepositoryGORM{db: db}
}

// Save creates or ignores a measurement (duplicate timestamps are silently ignored).
func (r *MeasurementRepositoryGORM) Save(ctx context.Context, m *domain.GlucoseMeasurement) error {
	db := txOrDefault(ctx, r.db)

	// ON CONFLICT DO NOTHING - ignore duplicates based on unique timestamp
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "timestamp"}},
		DoNothing: true,
	}).Create(m)

	return result.Error
}

// FindLatest returns the most recent measurement by timestamp.
func (r *MeasurementRepositoryGORM) FindLatest(ctx context.Context) (*domain.GlucoseMeasurement, error) {
	db := txOrDefault(ctx, r.db)

	var measurement domain.GlucoseMeasurement
	result := db.Order("timestamp DESC").First(&measurement)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, persistence.ErrNotFound
		}
		return nil, result.Error
	}

	return &measurement, nil
}

// FindAll returns all measurements ordered by timestamp descending.
func (r *MeasurementRepositoryGORM) FindAll(ctx context.Context) ([]*domain.GlucoseMeasurement, error) {
	db := txOrDefault(ctx, r.db)

	var measurements []*domain.GlucoseMeasurement
	result := db.Order("timestamp DESC").Find(&measurements)

	if result.Error != nil {
		return nil, result.Error
	}

	return measurements, nil
}

// FindByTimeRange returns measurements within a time range (inclusive).
func (r *MeasurementRepositoryGORM) FindByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.GlucoseMeasurement, error) {
	db := txOrDefault(ctx, r.db)

	var measurements []*domain.GlucoseMeasurement
	result := db.
		Where("timestamp >= ? AND timestamp <= ?", start, end).
		Order("timestamp DESC").
		Find(&measurements)

	if result.Error != nil {
		return nil, result.Error
	}

	return measurements, nil
}
