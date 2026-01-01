package memory

import (
	"sync"
	"testing"
	"time"

	"github.com/R4yL-dev/glcmd/internal/glucosemeasurement"
	"github.com/R4yL-dev/glcmd/internal/models"
)

// Test Glucose Measurements

func TestSaveMeasurement_Success(t *testing.T) {
	storage := New()

	measurement := &glucosemeasurement.GlucoseMeasurement{
		Timestamp: time.Now(),
		Value:     5.5,
	}

	err := storage.SaveMeasurement(measurement)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latest, err := storage.GetLatestMeasurement()
	if err != nil {
		t.Fatalf("unexpected error retrieving measurement: %v", err)
	}

	if latest.Value != 5.5 {
		t.Errorf("expected Value = 5.5, got %f", latest.Value)
	}
}

func TestSaveMeasurement_NilError(t *testing.T) {
	storage := New()

	err := storage.SaveMeasurement(nil)
	if err == nil {
		t.Fatal("expected error for nil measurement, got nil")
	}
}

func TestSaveMeasurement_Deduplication(t *testing.T) {
	storage := New()

	timestamp := time.Now()

	// Save first measurement
	measurement1 := &glucosemeasurement.GlucoseMeasurement{
		Timestamp: timestamp,
		Value:     5.5,
	}
	err := storage.SaveMeasurement(measurement1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to save duplicate with same timestamp
	measurement2 := &glucosemeasurement.GlucoseMeasurement{
		Timestamp: timestamp,
		Value:     6.0, // Different value but same timestamp
	}
	err = storage.SaveMeasurement(measurement2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have 1 measurement (duplicate ignored)
	all, _ := storage.GetAllMeasurements()
	if len(all) != 1 {
		t.Errorf("expected 1 measurement (duplicate ignored), got %d", len(all))
	}

	// Verify the first one was kept
	if all[0].Value != 5.5 {
		t.Errorf("expected first measurement to be kept (Value=5.5), got %f", all[0].Value)
	}
}

func TestGetLatestMeasurement_Empty(t *testing.T) {
	storage := New()

	_, err := storage.GetLatestMeasurement()
	if err == nil {
		t.Fatal("expected error for empty storage, got nil")
	}
}

func TestGetAllMeasurements_Multiple(t *testing.T) {
	storage := New()

	// Add 3 measurements
	for i := 0; i < 3; i++ {
		measurement := &glucosemeasurement.GlucoseMeasurement{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Value:     float64(i) + 5.0,
		}
		storage.SaveMeasurement(measurement)
	}

	all, err := storage.GetAllMeasurements()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("expected 3 measurements, got %d", len(all))
	}
}

func TestGetMeasurementsByTimeRange(t *testing.T) {
	storage := New()

	now := time.Now()

	// Add measurements at different times
	storage.SaveMeasurement(&glucosemeasurement.GlucoseMeasurement{
		Timestamp: now.Add(-2 * time.Hour),
		Value:     5.0,
	})
	storage.SaveMeasurement(&glucosemeasurement.GlucoseMeasurement{
		Timestamp: now.Add(-1 * time.Hour),
		Value:     6.0,
	})
	storage.SaveMeasurement(&glucosemeasurement.GlucoseMeasurement{
		Timestamp: now,
		Value:     7.0,
	})

	// Query for last hour
	start := now.Add(-1*time.Hour - 1*time.Minute)
	end := now.Add(1 * time.Minute)

	results, err := storage.GetMeasurementsByTimeRange(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 measurements in range, got %d", len(results))
	}
}

// Test Sensor Configuration

func TestSaveSensor_Success(t *testing.T) {
	storage := New()

	sensor := &models.SensorConfig{
		SerialNumber: "TEST123",
		IsActive:     true,
	}

	err := storage.SaveSensor(sensor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	active, err := storage.GetActiveSensor()
	if err != nil {
		t.Fatalf("unexpected error retrieving sensor: %v", err)
	}

	if active.SerialNumber != "TEST123" {
		t.Errorf("expected SerialNumber = 'TEST123', got %s", active.SerialNumber)
	}
}

func TestSaveSensor_Update(t *testing.T) {
	storage := New()

	// Save initial sensor
	sensor1 := &models.SensorConfig{
		SerialNumber: "TEST123",
		IsActive:     true,
		WarrantyDays: 60,
	}
	storage.SaveSensor(sensor1)

	// Update same sensor
	sensor2 := &models.SensorConfig{
		SerialNumber: "TEST123",
		IsActive:     true,
		WarrantyDays: 50,
	}
	storage.SaveSensor(sensor2)

	all, _ := storage.GetAllSensors()
	if len(all) != 1 {
		t.Errorf("expected 1 sensor after update, got %d", len(all))
	}

	if all[0].WarrantyDays != 50 {
		t.Errorf("expected updated WarrantyDays = 50, got %d", all[0].WarrantyDays)
	}
}

func TestSaveSensor_NilError(t *testing.T) {
	storage := New()

	err := storage.SaveSensor(nil)
	if err == nil {
		t.Fatal("expected error for nil sensor, got nil")
	}
}

func TestGetActiveSensor_Empty(t *testing.T) {
	storage := New()

	_, err := storage.GetActiveSensor()
	if err == nil {
		t.Fatal("expected error for empty storage, got nil")
	}
}

func TestGetActiveSensor_NoActive(t *testing.T) {
	storage := New()

	// Add sensor that's not marked active
	sensor := &models.SensorConfig{
		SerialNumber: "TEST123",
		IsActive:     false,
	}
	storage.SaveSensor(sensor)

	active, err := storage.GetActiveSensor()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the most recent one even if not marked active
	if active.SerialNumber != "TEST123" {
		t.Errorf("expected SerialNumber = 'TEST123', got %s", active.SerialNumber)
	}
}

// Test User Preferences

func TestSaveUserPreferences_Success(t *testing.T) {
	storage := New()

	user := &models.UserPreferences{
		UserID:    "test-user",
		FirstName: "Test",
		LastName:  "User",
	}

	err := storage.SaveUserPreferences(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := storage.GetUserPreferences()
	if err != nil {
		t.Fatalf("unexpected error retrieving user: %v", err)
	}

	if retrieved.UserID != "test-user" {
		t.Errorf("expected UserID = 'test-user', got %s", retrieved.UserID)
	}
}

func TestSaveUserPreferences_NilError(t *testing.T) {
	storage := New()

	err := storage.SaveUserPreferences(nil)
	if err == nil {
		t.Fatal("expected error for nil user preferences, got nil")
	}
}

func TestGetUserPreferences_Empty(t *testing.T) {
	storage := New()

	_, err := storage.GetUserPreferences()
	if err == nil {
		t.Fatal("expected error for empty storage, got nil")
	}
}

// Test Device Info

func TestSaveDeviceInfo_Success(t *testing.T) {
	storage := New()

	device := &models.DeviceInfo{
		DeviceID:  "test-device",
		HighLimit: 180,
		LowLimit:  70,
	}

	err := storage.SaveDeviceInfo(device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := storage.GetDeviceInfo()
	if err != nil {
		t.Fatalf("unexpected error retrieving device: %v", err)
	}

	if retrieved.DeviceID != "test-device" {
		t.Errorf("expected DeviceID = 'test-device', got %s", retrieved.DeviceID)
	}
}

func TestSaveDeviceInfo_NilError(t *testing.T) {
	storage := New()

	err := storage.SaveDeviceInfo(nil)
	if err == nil {
		t.Fatal("expected error for nil device info, got nil")
	}
}

func TestGetDeviceInfo_Empty(t *testing.T) {
	storage := New()

	_, err := storage.GetDeviceInfo()
	if err == nil {
		t.Fatal("expected error for empty storage, got nil")
	}
}

// Test Glucose Targets

func TestSaveGlucoseTargets_Success(t *testing.T) {
	storage := New()

	targets := &models.GlucoseTargets{
		TargetHigh: 180,
		TargetLow:  70,
	}

	err := storage.SaveGlucoseTargets(targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := storage.GetGlucoseTargets()
	if err != nil {
		t.Fatalf("unexpected error retrieving targets: %v", err)
	}

	if retrieved.TargetHigh != 180 {
		t.Errorf("expected TargetHigh = 180, got %d", retrieved.TargetHigh)
	}
}

func TestSaveGlucoseTargets_NilError(t *testing.T) {
	storage := New()

	err := storage.SaveGlucoseTargets(nil)
	if err == nil {
		t.Fatal("expected error for nil glucose targets, got nil")
	}
}

func TestGetGlucoseTargets_Empty(t *testing.T) {
	storage := New()

	_, err := storage.GetGlucoseTargets()
	if err == nil {
		t.Fatal("expected error for empty storage, got nil")
	}
}

// Test Thread Safety

func TestMemoryStorage_Concurrency(t *testing.T) {
	storage := New()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrently write measurements
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			measurement := &glucosemeasurement.GlucoseMeasurement{
				Timestamp: time.Now().Add(time.Duration(index) * time.Millisecond),
				Value:     float64(index),
			}
			storage.SaveMeasurement(measurement)
		}(i)
	}

	wg.Wait()

	// Verify all measurements were saved
	all, err := storage.GetAllMeasurements()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(all) != numGoroutines {
		t.Errorf("expected %d measurements, got %d", numGoroutines, len(all))
	}
}

func TestMemoryStorage_ConcurrentReadWrite(t *testing.T) {
	storage := New()

	var wg sync.WaitGroup

	// Initial measurement
	storage.SaveMeasurement(&glucosemeasurement.GlucoseMeasurement{
		Timestamp: time.Now(),
		Value:     5.0,
	})

	// Concurrent reads and writes
	for i := 0; i < 50; i++ {
		wg.Add(2)

		// Writer
		go func(index int) {
			defer wg.Done()
			measurement := &glucosemeasurement.GlucoseMeasurement{
				Timestamp: time.Now(),
				Value:     float64(index),
			}
			storage.SaveMeasurement(measurement)
		}(i)

		// Reader
		go func() {
			defer wg.Done()
			storage.GetLatestMeasurement()
			storage.GetAllMeasurements()
		}()
	}

	wg.Wait()

	// Should not panic or race
	all, _ := storage.GetAllMeasurements()
	if len(all) < 1 {
		t.Error("expected at least 1 measurement after concurrent operations")
	}
}
