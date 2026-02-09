package repository

import (
	"context"
	"testing"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

func TestGlucoseRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGlucoseRepository(db)

	now := time.Now().UTC()
	measurement := &domain.GlucoseMeasurement{
		FactoryTimestamp: now,
		Timestamp:        now,
		Value:            5.5,
		ValueInMgPerDl:   99,
	}

	_, err := repo.Save(context.Background(), measurement)
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

func TestGlucoseRepository_Save_DuplicateFactoryTimestamp(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGlucoseRepository(db)

	factoryTS := time.Now().UTC()

	m1 := &domain.GlucoseMeasurement{
		FactoryTimestamp: factoryTS,
		Timestamp:        factoryTS,
		Value:            5.5,
		ValueInMgPerDl:   99,
	}

	m2 := &domain.GlucoseMeasurement{
		FactoryTimestamp: factoryTS,              // Same factory timestamp!
		Timestamp:        factoryTS.Add(time.Second), // Different timestamp
		Value:            6.0,                    // Different value
		ValueInMgPerDl:   108,
	}

	// Save first measurement
	inserted1, err := repo.Save(context.Background(), m1)
	if err != nil {
		t.Fatalf("failed to save m1: %v", err)
	}
	if !inserted1 {
		t.Error("expected first measurement to be inserted")
	}

	// Save duplicate (should be ignored due to ON CONFLICT DO NOTHING)
	inserted2, err := repo.Save(context.Background(), m2)
	if err != nil {
		t.Fatalf("failed to save m2: %v", err)
	}
	if inserted2 {
		t.Error("expected duplicate measurement to be skipped (inserted=false)")
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

func TestGlucoseRepository_FindLatest(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGlucoseRepository(db)

	now := time.Now().UTC()

	measurements := []*domain.GlucoseMeasurement{
		{FactoryTimestamp: now.Add(-2 * time.Hour), Timestamp: now.Add(-2 * time.Hour), Value: 4.0, ValueInMgPerDl: 72},
		{FactoryTimestamp: now.Add(-1 * time.Hour), Timestamp: now.Add(-1 * time.Hour), Value: 5.0, ValueInMgPerDl: 90},
		{FactoryTimestamp: now, Timestamp: now, Value: 6.0, ValueInMgPerDl: 108}, // Latest
	}

	for _, m := range measurements {
		if _, err := repo.Save(context.Background(), m); err != nil {
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

func TestGlucoseRepository_FindLatest_NoData(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGlucoseRepository(db)

	_, err := repo.FindLatest(context.Background())
	if err != persistence.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGlucoseRepository_FindByTimeRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGlucoseRepository(db)

	now := time.Now().UTC()

	measurements := []*domain.GlucoseMeasurement{
		{FactoryTimestamp: now.Add(-3 * time.Hour), Timestamp: now.Add(-3 * time.Hour), Value: 4.0, ValueInMgPerDl: 72}, // Outside range
		{FactoryTimestamp: now.Add(-2 * time.Hour), Timestamp: now.Add(-2 * time.Hour), Value: 5.0, ValueInMgPerDl: 90}, // In range
		{FactoryTimestamp: now.Add(-1 * time.Hour), Timestamp: now.Add(-1 * time.Hour), Value: 6.0, ValueInMgPerDl: 108}, // In range
		{FactoryTimestamp: now, Timestamp: now, Value: 7.0, ValueInMgPerDl: 126},                     // In range
		{FactoryTimestamp: now.Add(1 * time.Hour), Timestamp: now.Add(1 * time.Hour), Value: 8.0, ValueInMgPerDl: 144},  // Outside range
	}

	for _, m := range measurements {
		if _, err := repo.Save(context.Background(), m); err != nil {
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

func TestGlucoseRepository_FindByTimeRange_EmptyRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGlucoseRepository(db)

	now := time.Now().UTC()

	// Save a measurement
	m := &domain.GlucoseMeasurement{
		FactoryTimestamp: now,
		Timestamp:        now,
		Value:            5.5,
		ValueInMgPerDl:   99,
	}
	if _, err := repo.Save(context.Background(), m); err != nil {
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
