package repository

import (
	"context"
	"testing"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

func TestSensorRepository_Save_Upsert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	now := time.Now().UTC()
	sensor := &domain.SensorConfig{
		SerialNumber: "TEST_SENSOR_1",
		Activation:   now.AddDate(0, 0, -5),
		ExpiresAt:    now.AddDate(0, 0, 10),
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now,
	}

	// First save - insert
	err := repo.Save(context.Background(), sensor)
	if err != nil {
		t.Fatalf("failed to save sensor: %v", err)
	}

	// Verify insert
	retrieved, err := repo.FindBySerialNumber(context.Background(), "TEST_SENSOR_1")
	if err != nil {
		t.Fatalf("failed to retrieve sensor: %v", err)
	}

	if retrieved.DurationDays != 15 {
		t.Errorf("expected DurationDays = 15, got %d", retrieved.DurationDays)
	}

	if retrieved.SensorType != 4 {
		t.Errorf("expected SensorType = 4, got %d", retrieved.SensorType)
	}

	// Update the same sensor (upsert)
	sensor.DurationDays = 14
	sensor.SensorType = 3

	err = repo.Save(context.Background(), sensor)
	if err != nil {
		t.Fatalf("failed to update sensor: %v", err)
	}

	// Verify update
	updated, err := repo.FindBySerialNumber(context.Background(), "TEST_SENSOR_1")
	if err != nil {
		t.Fatalf("failed to retrieve updated sensor: %v", err)
	}

	if updated.DurationDays != 14 {
		t.Errorf("expected DurationDays = 14 after update, got %d", updated.DurationDays)
	}

	if updated.SensorType != 3 {
		t.Errorf("expected SensorType = 3 after update, got %d", updated.SensorType)
	}
}

func TestSensorRepository_FindCurrent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	now := time.Now().UTC()
	endedAt := now.AddDate(0, 0, -2)

	endedSensor := &domain.SensorConfig{
		SerialNumber: "ENDED_SENSOR",
		Activation:   now.AddDate(0, 0, -20),
		ExpiresAt:    now.AddDate(0, 0, -5),
		EndedAt:      &endedAt,
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now.AddDate(0, 0, -20),
	}

	currentSensor := &domain.SensorConfig{
		SerialNumber: "CURRENT_SENSOR",
		Activation:   now.AddDate(0, 0, -5),
		ExpiresAt:    now.AddDate(0, 0, 10),
		EndedAt:      nil, // Current sensor has no EndedAt
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now,
	}

	// Save both sensors
	repo.Save(context.Background(), endedSensor)
	repo.Save(context.Background(), currentSensor)

	// Find current sensor
	current, err := repo.FindCurrent(context.Background())
	if err != nil {
		t.Fatalf("failed to find current sensor: %v", err)
	}

	if current.SerialNumber != "CURRENT_SENSOR" {
		t.Errorf("expected SerialNumber = CURRENT_SENSOR, got %s", current.SerialNumber)
	}

	if current.EndedAt != nil {
		t.Error("expected EndedAt = nil for current sensor")
	}
}

func TestSensorRepository_FindCurrent_NoCurrentSensor(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	now := time.Now().UTC()
	endedAt := now.AddDate(0, 0, -2)

	endedSensor := &domain.SensorConfig{
		SerialNumber: "ENDED_SENSOR",
		Activation:   now.AddDate(0, 0, -20),
		ExpiresAt:    now.AddDate(0, 0, -5),
		EndedAt:      &endedAt,
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now.AddDate(0, 0, -20),
	}

	repo.Save(context.Background(), endedSensor)

	// Find current sensor (should not exist)
	_, err := repo.FindCurrent(context.Background())
	if err != persistence.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSensorRepository_SetEndedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	now := time.Now().UTC()
	sensor := &domain.SensorConfig{
		SerialNumber: "TEST_SENSOR",
		Activation:   now.AddDate(0, 0, -10),
		ExpiresAt:    now.AddDate(0, 0, 5),
		EndedAt:      nil,
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now.AddDate(0, 0, -10),
	}

	repo.Save(context.Background(), sensor)

	// Set ended at
	endedAt := now
	err := repo.SetEndedAt(context.Background(), "TEST_SENSOR", endedAt)
	if err != nil {
		t.Fatalf("failed to set ended_at: %v", err)
	}

	// Verify update
	updated, err := repo.FindBySerialNumber(context.Background(), "TEST_SENSOR")
	if err != nil {
		t.Fatalf("failed to retrieve sensor: %v", err)
	}

	if updated.EndedAt == nil {
		t.Fatal("expected EndedAt to be set")
	}

	// Compare times (truncate to second precision for comparison)
	if updated.EndedAt.Unix() != endedAt.Unix() {
		t.Errorf("expected EndedAt = %v, got %v", endedAt, *updated.EndedAt)
	}
}

func TestSensorRepository_SetEndedAt_NonExistent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	// Try to set ended_at on non-existent sensor
	err := repo.SetEndedAt(context.Background(), "NONEXISTENT", time.Now().UTC())
	if err != persistence.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSensorRepository_FindAll_OrderByDetectedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	now := time.Now().UTC()

	sensors := []*domain.SensorConfig{
		{
			SerialNumber: "SENSOR_1",
			Activation:   now.AddDate(0, 0, -30),
			ExpiresAt:    now.AddDate(0, 0, -15),
			SensorType:   4,
			DurationDays: 15,
			DetectedAt:   now.Add(-2 * time.Hour),
		},
		{
			SerialNumber: "SENSOR_2",
			Activation:   now.AddDate(0, 0, -20),
			ExpiresAt:    now.AddDate(0, 0, -5),
			SensorType:   4,
			DurationDays: 15,
			DetectedAt:   now.Add(-1 * time.Hour),
		},
		{
			SerialNumber: "SENSOR_3",
			Activation:   now.AddDate(0, 0, -5),
			ExpiresAt:    now.AddDate(0, 0, 10),
			SensorType:   4,
			DurationDays: 15,
			DetectedAt:   now, // Most recent
		},
	}

	for _, s := range sensors {
		repo.Save(context.Background(), s)
	}

	all, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("failed to find all: %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("expected 3 sensors, got %d", len(all))
	}

	// Verify descending order (newest first)
	if all[0].SerialNumber != "SENSOR_3" {
		t.Errorf("expected first sensor = SENSOR_3 (newest), got %s", all[0].SerialNumber)
	}
}

func TestSensorRepository_FindCurrent_ReturnsLatestWhenMultiple(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	now := time.Now().UTC()

	// Save two current sensors (should not happen in practice, but test the behavior)
	sensor1 := &domain.SensorConfig{
		SerialNumber: "SENSOR_1",
		Activation:   now.AddDate(0, 0, -10),
		ExpiresAt:    now.AddDate(0, 0, 5),
		EndedAt:      nil,
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now.Add(-1 * time.Hour),
	}

	sensor2 := &domain.SensorConfig{
		SerialNumber: "SENSOR_2",
		Activation:   now.AddDate(0, 0, -5),
		ExpiresAt:    now.AddDate(0, 0, 10),
		EndedAt:      nil,
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now, // More recent
	}

	repo.Save(context.Background(), sensor1)
	repo.Save(context.Background(), sensor2)

	// FindCurrent should return the most recently detected
	current, err := repo.FindCurrent(context.Background())
	if err != nil {
		t.Fatalf("failed to find current sensor: %v", err)
	}

	if current.SerialNumber != "SENSOR_2" {
		t.Errorf("expected SerialNumber = SENSOR_2 (most recent), got %s", current.SerialNumber)
	}
}
