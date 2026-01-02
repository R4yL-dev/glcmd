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

	sensor := &domain.SensorConfig{
		SerialNumber: "TEST_SENSOR_1",
		IsActive:     true,
		WarrantyDays: 10,
		DetectedAt:   time.Now().UTC(),
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

	if retrieved.WarrantyDays != 10 {
		t.Errorf("expected WarrantyDays = 10, got %d", retrieved.WarrantyDays)
	}

	// Update the same sensor (upsert)
	sensor.WarrantyDays = 5
	sensor.IsActive = false

	err = repo.Save(context.Background(), sensor)
	if err != nil {
		t.Fatalf("failed to update sensor: %v", err)
	}

	// Verify update
	updated, err := repo.FindBySerialNumber(context.Background(), "TEST_SENSOR_1")
	if err != nil {
		t.Fatalf("failed to retrieve updated sensor: %v", err)
	}

	if updated.WarrantyDays != 5 {
		t.Errorf("expected WarrantyDays = 5 after update, got %d", updated.WarrantyDays)
	}

	if updated.IsActive != false {
		t.Error("expected IsActive = false after update")
	}
}

func TestSensorRepository_FindActive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	inactiveSensor := &domain.SensorConfig{
		SerialNumber: "INACTIVE_SENSOR",
		IsActive:     false,
		DetectedAt:   time.Now().UTC(),
	}

	activeSensor := &domain.SensorConfig{
		SerialNumber: "ACTIVE_SENSOR",
		IsActive:     true,
		DetectedAt:   time.Now().UTC(),
	}

	// Save both sensors
	repo.Save(context.Background(), inactiveSensor)
	repo.Save(context.Background(), activeSensor)

	// Find active sensor
	active, err := repo.FindActive(context.Background())
	if err != nil {
		t.Fatalf("failed to find active sensor: %v", err)
	}

	if active.SerialNumber != "ACTIVE_SENSOR" {
		t.Errorf("expected SerialNumber = ACTIVE_SENSOR, got %s", active.SerialNumber)
	}

	if !active.IsActive {
		t.Error("expected IsActive = true")
	}
}

func TestSensorRepository_FindActive_NoActiveSensor(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	inactiveSensor := &domain.SensorConfig{
		SerialNumber: "INACTIVE_SENSOR",
		IsActive:     false,
		DetectedAt:   time.Now().UTC(),
	}

	repo.Save(context.Background(), inactiveSensor)

	// Find active sensor (should not exist)
	_, err := repo.FindActive(context.Background())
	if err != persistence.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSensorRepository_UpdateActiveStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	sensor := &domain.SensorConfig{
		SerialNumber: "TEST_SENSOR",
		IsActive:     true,
		DetectedAt:   time.Now().UTC(),
	}

	repo.Save(context.Background(), sensor)

	// Deactivate sensor
	err := repo.UpdateActiveStatus(context.Background(), "TEST_SENSOR", false)
	if err != nil {
		t.Fatalf("failed to update active status: %v", err)
	}

	// Verify update
	updated, err := repo.FindBySerialNumber(context.Background(), "TEST_SENSOR")
	if err != nil {
		t.Fatalf("failed to retrieve sensor: %v", err)
	}

	if updated.IsActive {
		t.Error("expected IsActive = false after deactivation")
	}
}

func TestSensorRepository_UpdateActiveStatus_NonExistent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	// Try to update non-existent sensor
	err := repo.UpdateActiveStatus(context.Background(), "NONEXISTENT", false)
	if err != persistence.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSensorRepository_FindAll_OrderByDetectedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSensorRepository(db)

	now := time.Now().UTC()

	sensors := []*domain.SensorConfig{
		{SerialNumber: "SENSOR_1", DetectedAt: now.Add(-2 * time.Hour)},
		{SerialNumber: "SENSOR_2", DetectedAt: now.Add(-1 * time.Hour)},
		{SerialNumber: "SENSOR_3", DetectedAt: now}, // Most recent
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
