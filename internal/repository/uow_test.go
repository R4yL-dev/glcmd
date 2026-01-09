package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/R4yL-dev/glcmd/internal/domain"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(
		&domain.GlucoseMeasurement{},
		&domain.SensorConfig{},
		&domain.UserPreferences{},
		&domain.DeviceInfo{},
		&domain.GlucoseTargets{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestUnitOfWork_ExecuteInTransaction_Commit(t *testing.T) {
	db := setupTestDB(t)
	uow := NewUnitOfWork(db)
	sensorRepo := NewSensorRepository(db)

	now := time.Now().UTC()
	sensor := &domain.SensorConfig{
		SerialNumber: "TEST123",
		Activation:   now.AddDate(0, 0, -5),
		ExpiresAt:    now.AddDate(0, 0, 10),
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now,
	}

	// Execute in transaction
	err := uow.ExecuteInTransaction(context.Background(), func(txCtx context.Context) error {
		return sensorRepo.Save(txCtx, sensor)
	})

	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Verify data was committed
	retrieved, err := sensorRepo.FindBySerialNumber(context.Background(), "TEST123")
	if err != nil {
		t.Fatalf("failed to retrieve sensor: %v", err)
	}

	if retrieved.SerialNumber != "TEST123" {
		t.Errorf("expected SerialNumber = TEST123, got %s", retrieved.SerialNumber)
	}
}

func TestUnitOfWork_ExecuteInTransaction_Rollback(t *testing.T) {
	db := setupTestDB(t)
	uow := NewUnitOfWork(db)
	sensorRepo := NewSensorRepository(db)

	now := time.Now().UTC()
	sensor := &domain.SensorConfig{
		SerialNumber: "TEST456",
		Activation:   now.AddDate(0, 0, -5),
		ExpiresAt:    now.AddDate(0, 0, 10),
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now,
	}

	// Execute in transaction with error
	testErr := errors.New("test error")
	err := uow.ExecuteInTransaction(context.Background(), func(txCtx context.Context) error {
		if err := sensorRepo.Save(txCtx, sensor); err != nil {
			return err
		}
		// Return error to trigger rollback
		return testErr
	})

	if err != testErr {
		t.Fatalf("expected error %v, got %v", testErr, err)
	}

	// Verify data was rolled back (not committed)
	_, err = sensorRepo.FindBySerialNumber(context.Background(), "TEST456")
	if err == nil {
		t.Error("expected sensor to not exist after rollback, but it was found")
	}
}

func TestUnitOfWork_ExecuteInTransaction_ContextPropagation(t *testing.T) {
	db := setupTestDB(t)
	uow := NewUnitOfWork(db)
	sensorRepo := NewSensorRepository(db)

	now := time.Now().UTC()
	sensor1 := &domain.SensorConfig{
		SerialNumber: "SENSOR1",
		Activation:   now.AddDate(0, 0, -10),
		ExpiresAt:    now.AddDate(0, 0, 5),
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now.Add(-1 * time.Hour),
	}
	sensor2 := &domain.SensorConfig{
		SerialNumber: "SENSOR2",
		Activation:   now.AddDate(0, 0, -5),
		ExpiresAt:    now.AddDate(0, 0, 10),
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now,
	}

	// Execute multiple operations in same transaction
	err := uow.ExecuteInTransaction(context.Background(), func(txCtx context.Context) error {
		// Both saves should use the same transaction
		if err := sensorRepo.Save(txCtx, sensor1); err != nil {
			return err
		}
		if err := sensorRepo.Save(txCtx, sensor2); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Verify both sensors were committed atomically
	all, err := sensorRepo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve sensors: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("expected 2 sensors, got %d", len(all))
	}
}

func TestUnitOfWork_ExecuteInTransaction_RollbackOnSecondError(t *testing.T) {
	db := setupTestDB(t)
	uow := NewUnitOfWork(db)
	sensorRepo := NewSensorRepository(db)

	now := time.Now().UTC()
	sensor1 := &domain.SensorConfig{
		SerialNumber: "SENSOR_A",
		Activation:   now.AddDate(0, 0, -5),
		ExpiresAt:    now.AddDate(0, 0, 10),
		SensorType:   4,
		DurationDays: 15,
		DetectedAt:   now,
	}

	testErr := errors.New("second operation failed")

	// Execute transaction where second operation fails
	err := uow.ExecuteInTransaction(context.Background(), func(txCtx context.Context) error {
		// First save succeeds
		if err := sensorRepo.Save(txCtx, sensor1); err != nil {
			return err
		}

		// Second operation fails
		return testErr
	})

	if err != testErr {
		t.Fatalf("expected error %v, got %v", testErr, err)
	}

	// Verify BOTH operations were rolled back (atomicity)
	all, err := sensorRepo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve sensors: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("expected 0 sensors after rollback, got %d", len(all))
	}
}
