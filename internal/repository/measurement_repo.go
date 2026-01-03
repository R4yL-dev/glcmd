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

// FindWithFilters returns measurements matching filters with pagination.
func (r *MeasurementRepositoryGORM) FindWithFilters(ctx context.Context, filters MeasurementFilters, limit, offset int) ([]*domain.GlucoseMeasurement, error) {
	db := txOrDefault(ctx, r.db)

	query := db.Model(&domain.GlucoseMeasurement{})

	// Apply filters
	if filters.StartTime != nil {
		query = query.Where("timestamp >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("timestamp <= ?", *filters.EndTime)
	}
	if filters.Color != nil {
		query = query.Where("measurement_color = ?", *filters.Color)
	}
	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}

	var measurements []*domain.GlucoseMeasurement
	result := query.
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&measurements)

	if result.Error != nil {
		return nil, result.Error
	}

	return measurements, nil
}

// CountWithFilters returns total count of measurements matching filters.
func (r *MeasurementRepositoryGORM) CountWithFilters(ctx context.Context, filters MeasurementFilters) (int64, error) {
	db := txOrDefault(ctx, r.db)

	query := db.Model(&domain.GlucoseMeasurement{})

	// Apply filters
	if filters.StartTime != nil {
		query = query.Where("timestamp >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("timestamp <= ?", *filters.EndTime)
	}
	if filters.Color != nil {
		query = query.Where("measurement_color = ?", *filters.Color)
	}
	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}

	var count int64
	result := query.Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}
