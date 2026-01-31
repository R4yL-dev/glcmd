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

	// Upsert: Update fields on conflict with serial_number
	// Note: ended_at is NOT updated on conflict - it's only set via SetEndedAt
	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "serial_number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"activation", "expires_at", "sensor_type", "duration_days",
			"detected_at", "updated_at", "last_measurement_at",
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

// FindCurrent returns the current sensor (EndedAt is null).
func (r *SensorRepositoryGORM) FindCurrent(ctx context.Context) (*domain.SensorConfig, error) {
	db := txOrDefault(ctx, r.db)

	var sensor domain.SensorConfig
	result := db.Where("ended_at IS NULL").Order("detected_at DESC").First(&sensor)

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

// FindWithFilters returns sensors matching filters with pagination.
func (r *SensorRepositoryGORM) FindWithFilters(ctx context.Context, filters SensorFilters, limit, offset int) ([]*domain.SensorConfig, error) {
	db := txOrDefault(ctx, r.db)

	query := db.Model(&domain.SensorConfig{})

	if filters.StartTime != nil {
		query = query.Where("activation >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("activation <= ?", *filters.EndTime)
	}

	var sensors []*domain.SensorConfig
	result := query.
		Order("activation DESC").
		Limit(limit).
		Offset(offset).
		Find(&sensors)

	if result.Error != nil {
		return nil, result.Error
	}

	return sensors, nil
}

// CountWithFilters returns total count of sensors matching filters.
func (r *SensorRepositoryGORM) CountWithFilters(ctx context.Context, filters SensorFilters) (int64, error) {
	db := txOrDefault(ctx, r.db)

	query := db.Model(&domain.SensorConfig{})

	if filters.StartTime != nil {
		query = query.Where("activation >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("activation <= ?", *filters.EndTime)
	}

	var count int64
	result := query.Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetStatistics returns aggregated sensor lifecycle statistics computed by SQL.
func (r *SensorRepositoryGORM) GetStatistics(ctx context.Context) (*SensorStatisticsResult, error) {
	db := txOrDefault(ctx, r.db)

	selectClause := `
		COUNT(*) as total_sensors,
		COALESCE(SUM(CASE WHEN ended_at IS NOT NULL THEN 1 ELSE 0 END), 0) as ended_sensors,
		COALESCE(AVG(CASE WHEN ended_at IS NOT NULL
			THEN (julianday(ended_at) - julianday(activation)) END), 0) as avg_duration,
		COALESCE(MIN(CASE WHEN ended_at IS NOT NULL
			THEN (julianday(ended_at) - julianday(activation)) END), 0) as min_duration,
		COALESCE(MAX(CASE WHEN ended_at IS NOT NULL
			THEN (julianday(ended_at) - julianday(activation)) END), 0) as max_duration,
		COALESCE(AVG(duration_days), 0) as avg_expected
	`

	var result SensorStatisticsResult
	if err := db.Model(&domain.SensorConfig{}).Select(selectClause).Scan(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

// SetEndedAt marks a sensor as ended (replaced by a new sensor).
func (r *SensorRepositoryGORM) SetEndedAt(ctx context.Context, serial string, endedAt time.Time) error {
	db := txOrDefault(ctx, r.db)

	result := db.Model(&domain.SensorConfig{}).
		Where("serial_number = ?", serial).
		Update("ended_at", endedAt)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return persistence.ErrNotFound
	}

	return nil
}
