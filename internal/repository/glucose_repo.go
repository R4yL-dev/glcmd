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

// GlucoseRepositoryGORM is the GORM implementation of GlucoseRepository.
type GlucoseRepositoryGORM struct {
	db *gorm.DB
}

// NewGlucoseRepository creates a new GlucoseRepository.
func NewGlucoseRepository(db *gorm.DB) *GlucoseRepositoryGORM {
	return &GlucoseRepositoryGORM{db: db}
}

// Save creates or ignores a measurement (duplicate timestamps are silently ignored).
// Returns (true, nil) if inserted, (false, nil) if duplicate was ignored.
func (r *GlucoseRepositoryGORM) Save(ctx context.Context, m *domain.GlucoseMeasurement) (bool, error) {
	db := txOrDefault(ctx, r.db)

	// ON CONFLICT DO NOTHING - ignore duplicates based on unique timestamp
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "timestamp"}},
		DoNothing: true,
	}).Create(m)

	// RowsAffected = 0 if conflict (duplicate), 1 if inserted successfully
	// No extra query needed - value provided by SQL driver
	return result.RowsAffected > 0, result.Error
}

// FindLatest returns the most recent measurement by timestamp.
func (r *GlucoseRepositoryGORM) FindLatest(ctx context.Context) (*domain.GlucoseMeasurement, error) {
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
func (r *GlucoseRepositoryGORM) FindAll(ctx context.Context) ([]*domain.GlucoseMeasurement, error) {
	db := txOrDefault(ctx, r.db)

	var measurements []*domain.GlucoseMeasurement
	result := db.Order("timestamp DESC").Find(&measurements)

	if result.Error != nil {
		return nil, result.Error
	}

	return measurements, nil
}

// FindByTimeRange returns measurements within a time range (inclusive).
func (r *GlucoseRepositoryGORM) FindByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.GlucoseMeasurement, error) {
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
func (r *GlucoseRepositoryGORM) FindWithFilters(ctx context.Context, filters GlucoseFilters, limit, offset int) ([]*domain.GlucoseMeasurement, error) {
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
func (r *GlucoseRepositoryGORM) CountWithFilters(ctx context.Context, filters GlucoseFilters) (int64, error) {
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

// parseTimestamp tries to parse a timestamp string in various formats
func parseTimestamp(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}

	// Try common timestamp formats
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, *s); err == nil {
			return &t
		}
	}

	return nil
}

// statisticsRawResult is used for scanning SQL results with string timestamps
type statisticsRawResult struct {
	Count           int64
	Average         float64
	AverageMgDl     float64
	Min             float64
	MinMgDl         int
	Max             float64
	MaxMgDl         int
	Variance        float64
	LowCount        int64
	NormalCount     int64
	HighCount       int64
	InRangeCount    int64
	BelowRangeCount int64
	AboveRangeCount int64
	FirstTimestamp  *string // SQLite returns timestamps as strings
	LastTimestamp   *string
}

// GetStatistics returns aggregated statistics computed by SQL.
func (r *GlucoseRepositoryGORM) GetStatistics(ctx context.Context, filters GlucoseStatisticsFilters) (*GlucoseStatisticsResult, error) {
	db := txOrDefault(ctx, r.db)

	// Base aggregation query
	// Variance = E[X²] - E[X]², SQRT computed in Go for SQLite compatibility
	selectClause := `
		COUNT(*) as count,
		COALESCE(AVG(value), 0) as average,
		COALESCE(AVG(value_in_mg_per_dl), 0) as average_mg_dl,
		COALESCE(MIN(value), 0) as min,
		COALESCE(MIN(value_in_mg_per_dl), 0) as min_mg_dl,
		COALESCE(MAX(value), 0) as max,
		COALESCE(MAX(value_in_mg_per_dl), 0) as max_mg_dl,
		COALESCE(ABS(AVG(value * value) - AVG(value) * AVG(value)), 0) as variance,
		COALESCE(SUM(CASE WHEN measurement_color = 1 THEN 1 ELSE 0 END), 0) as normal_count,
		COALESCE(SUM(CASE WHEN measurement_color IN (2, 3) AND is_low = 1 THEN 1 ELSE 0 END), 0) as low_count,
		COALESCE(SUM(CASE WHEN measurement_color IN (2, 3) AND is_low = 0 THEN 1 ELSE 0 END), 0) as high_count,
		MIN(timestamp) as first_timestamp,
		MAX(timestamp) as last_timestamp
	`

	// Add Time in Range columns if targets are provided
	if filters.TargetLowMgDl != nil && filters.TargetHighMgDl != nil {
		selectClause += `,
			COALESCE(SUM(CASE WHEN value_in_mg_per_dl < ? THEN 1 ELSE 0 END), 0) as below_range_count,
			COALESCE(SUM(CASE WHEN value_in_mg_per_dl > ? THEN 1 ELSE 0 END), 0) as above_range_count,
			COALESCE(SUM(CASE WHEN value_in_mg_per_dl >= ? AND value_in_mg_per_dl <= ? THEN 1 ELSE 0 END), 0) as in_range_count
		`
	}

	query := db.Model(&domain.GlucoseMeasurement{})

	// Add TIR parameters to select if targets are provided
	if filters.TargetLowMgDl != nil && filters.TargetHighMgDl != nil {
		query = query.Select(selectClause,
			*filters.TargetLowMgDl,  // below_range_count
			*filters.TargetHighMgDl, // above_range_count
			*filters.TargetLowMgDl,  // in_range_count lower bound
			*filters.TargetHighMgDl, // in_range_count upper bound
		)
	} else {
		query = query.Select(selectClause)
	}

	// Apply time filters
	if filters.StartTime != nil {
		query = query.Where("timestamp >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("timestamp <= ?", *filters.EndTime)
	}

	var raw statisticsRawResult
	if err := query.Scan(&raw).Error; err != nil {
		return nil, err
	}

	// Convert to result with parsed timestamps
	result := &GlucoseStatisticsResult{
		Count:           raw.Count,
		Average:         raw.Average,
		AverageMgDl:     raw.AverageMgDl,
		Min:             raw.Min,
		MinMgDl:         raw.MinMgDl,
		Max:             raw.Max,
		MaxMgDl:         raw.MaxMgDl,
		Variance:        raw.Variance,
		LowCount:        raw.LowCount,
		NormalCount:     raw.NormalCount,
		HighCount:       raw.HighCount,
		InRangeCount:    raw.InRangeCount,
		BelowRangeCount: raw.BelowRangeCount,
		AboveRangeCount: raw.AboveRangeCount,
	}

	// Parse timestamps (SQLite stores them as strings in various formats)
	result.FirstTimestamp = parseTimestamp(raw.FirstTimestamp)
	result.LastTimestamp = parseTimestamp(raw.LastTimestamp)

	return result, nil
}
