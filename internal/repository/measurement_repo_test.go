package repository

import (
	"context"
	"testing"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

func TestMeasurementRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMeasurementRepository(db)

	measurement := &domain.GlucoseMeasurement{
		Timestamp:      time.Now().UTC(),
		Value:          5.5,
		ValueInMgPerDl: 99,
	}

	err := repo.Save(context.Background(), measurement)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify save
	latest, err := repo.FindLatest(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve: %v", err)
	}

	if latest.Value != 5.5 {
		t.Errorf("expected Value = 5.5, got %f", latest.Value)
	}
}

func TestMeasurementRepository_Save_DuplicateTimestamp(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMeasurementRepository(db)

	timestamp := time.Now().UTC()

	m1 := &domain.GlucoseMeasurement{
		Timestamp:      timestamp,
		Value:          5.5,
		ValueInMgPerDl: 99,
	}

	m2 := &domain.GlucoseMeasurement{
		Timestamp:      timestamp, // Same timestamp!
		Value:          6.0,       // Different value
		ValueInMgPerDl: 108,
	}

	// Save first measurement
	if err := repo.Save(context.Background(), m1); err != nil {
		t.Fatalf("failed to save m1: %v", err)
	}

	// Save duplicate (should be ignored due to ON CONFLICT DO NOTHING)
	if err := repo.Save(context.Background(), m2); err != nil {
		t.Fatalf("failed to save m2: %v", err)
	}

	// Verify only one measurement exists (the first one)
	all, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve all: %v", err)
	}

	if len(all) != 1 {
		t.Fatalf("expected 1 measurement (duplicate ignored), got %d", len(all))
	}

	// Verify it's the first measurement's value
	if all[0].Value != 5.5 {
		t.Errorf("expected Value = 5.5 (first measurement), got %f", all[0].Value)
	}
}

func TestMeasurementRepository_FindLatest(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMeasurementRepository(db)

	now := time.Now().UTC()

	measurements := []*domain.GlucoseMeasurement{
		{Timestamp: now.Add(-2 * time.Hour), Value: 4.0, ValueInMgPerDl: 72},
		{Timestamp: now.Add(-1 * time.Hour), Value: 5.0, ValueInMgPerDl: 90},
		{Timestamp: now, Value: 6.0, ValueInMgPerDl: 108}, // Latest
	}

	for _, m := range measurements {
		if err := repo.Save(context.Background(), m); err != nil {
			t.Fatalf("failed to save measurement: %v", err)
		}
	}

	latest, err := repo.FindLatest(context.Background())
	if err != nil {
		t.Fatalf("failed to find latest: %v", err)
	}

	if latest.Value != 6.0 {
		t.Errorf("expected latest Value = 6.0, got %f", latest.Value)
	}
}

func TestMeasurementRepository_FindLatest_NoData(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMeasurementRepository(db)

	_, err := repo.FindLatest(context.Background())
	if err != persistence.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMeasurementRepository_FindByTimeRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMeasurementRepository(db)

	now := time.Now().UTC()

	measurements := []*domain.GlucoseMeasurement{
		{Timestamp: now.Add(-3 * time.Hour), Value: 4.0, ValueInMgPerDl: 72}, // Outside range
		{Timestamp: now.Add(-2 * time.Hour), Value: 5.0, ValueInMgPerDl: 90}, // In range
		{Timestamp: now.Add(-1 * time.Hour), Value: 6.0, ValueInMgPerDl: 108}, // In range
		{Timestamp: now, Value: 7.0, ValueInMgPerDl: 126},                     // In range
		{Timestamp: now.Add(1 * time.Hour), Value: 8.0, ValueInMgPerDl: 144},  // Outside range
	}

	for _, m := range measurements {
		if err := repo.Save(context.Background(), m); err != nil {
			t.Fatalf("failed to save measurement: %v", err)
		}
	}

	// Query range: -2h to now
	start := now.Add(-2 * time.Hour)
	end := now

	results, err := repo.FindByTimeRange(context.Background(), start, end)
	if err != nil {
		t.Fatalf("failed to find by time range: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 measurements in range, got %d", len(results))
	}

	// Verify results are in descending order (newest first)
	if len(results) >= 2 {
		if results[0].Timestamp.Before(results[1].Timestamp) {
			t.Error("expected results in descending order (newest first)")
		}
	}
}

func TestMeasurementRepository_FindByTimeRange_EmptyRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMeasurementRepository(db)

	now := time.Now().UTC()

	// Save a measurement
	m := &domain.GlucoseMeasurement{
		Timestamp:      now,
		Value:          5.5,
		ValueInMgPerDl: 99,
	}
	if err := repo.Save(context.Background(), m); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Query range that doesn't include the measurement
	start := now.Add(-2 * time.Hour)
	end := now.Add(-1 * time.Hour)

	results, err := repo.FindByTimeRange(context.Background(), start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 measurements in empty range, got %d", len(results))
	}
}
