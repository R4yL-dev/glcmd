package models

import (
	"testing"
	"time"
)

// TestSensorConfig_StructInitialization tests that SensorConfig can be initialized
func TestSensorConfig_StructInitialization(t *testing.T) {
	now := time.Now()

	sensor := SensorConfig{
		SerialNumber: "0JW90U63N0",
		Activation:   now,
		DeviceID:     "test-device-id",
		SensorType:   4,
		WarrantyDays: 60,
		IsActive:     true,
		LowJourney:   false,
		DetectedAt:   now,
	}

	if sensor.SerialNumber != "0JW90U63N0" {
		t.Errorf("expected SerialNumber = '0JW90U63N0', got %s", sensor.SerialNumber)
	}

	if sensor.SensorType != 4 {
		t.Errorf("expected SensorType = 4, got %d", sensor.SensorType)
	}

	if !sensor.IsActive {
		t.Error("expected IsActive = true, got false")
	}
}

// TestUserPreferences_StructInitialization tests that UserPreferences can be initialized
func TestUserPreferences_StructInitialization(t *testing.T) {
	now := time.Now()

	user := UserPreferences{
		UserID:                "test-user-id",
		FirstName:             "John",
		LastName:              "Doe",
		Email:                 "john@example.com",
		Country:               "FR",
		AccountType:           "pat",
		DateOfBirth:           now,
		Created:               now,
		LastLogin:             now,
		UILanguage:            "fr",
		CommunicationLanguage: "fr",
		UnitOfMeasure:         0,
		DateFormat:            2,
		TimeFormat:            2,
		EmailDays:             []int{1, 3, 5},
	}

	if user.UserID != "test-user-id" {
		t.Errorf("expected UserID = 'test-user-id', got %s", user.UserID)
	}

	if user.UnitOfMeasure != 0 {
		t.Errorf("expected UnitOfMeasure = 0, got %d", user.UnitOfMeasure)
	}

	if len(user.EmailDays) != 3 {
		t.Errorf("expected EmailDays length = 3, got %d", len(user.EmailDays))
	}
}

// TestDeviceInfo_StructInitialization tests that DeviceInfo can be initialized
func TestDeviceInfo_StructInitialization(t *testing.T) {
	now := time.Now()

	device := DeviceInfo{
		DeviceID:          "test-device-id",
		DeviceTypeID:      40068,
		AppVersion:        "3.6.5",
		AlarmsEnabled:     true,
		HighLimit:         180,
		LowLimit:          70,
		FixedLowThreshold: 70,
		LastUpdate:        now,
		LimitEnabled:      true,
	}

	if device.DeviceID != "test-device-id" {
		t.Errorf("expected DeviceID = 'test-device-id', got %s", device.DeviceID)
	}

	if device.DeviceTypeID != 40068 {
		t.Errorf("expected DeviceTypeID = 40068, got %d", device.DeviceTypeID)
	}

	if device.HighLimit != 180 {
		t.Errorf("expected HighLimit = 180, got %d", device.HighLimit)
	}
}

// TestGlucoseTargets_StructInitialization tests that GlucoseTargets can be initialized
func TestGlucoseTargets_StructInitialization(t *testing.T) {
	targets := GlucoseTargets{
		TargetHigh:    180,
		TargetLow:     70,
		UnitOfMeasure: 0,
	}

	if targets.TargetHigh != 180 {
		t.Errorf("expected TargetHigh = 180, got %d", targets.TargetHigh)
	}

	if targets.TargetLow != 70 {
		t.Errorf("expected TargetLow = 70, got %d", targets.TargetLow)
	}

	if targets.UnitOfMeasure != 0 {
		t.Errorf("expected UnitOfMeasure = 0, got %d", targets.UnitOfMeasure)
	}
}

// TestFixedLowAlarmValues_StructInitialization tests that FixedLowAlarmValues can be initialized
func TestFixedLowAlarmValues_StructInitialization(t *testing.T) {
	alarms := FixedLowAlarmValues{
		MgPerDl:  60,
		MmolPerL: 3.3,
	}

	if alarms.MgPerDl != 60 {
		t.Errorf("expected MgPerDl = 60, got %d", alarms.MgPerDl)
	}

	if alarms.MmolPerL != 3.3 {
		t.Errorf("expected MmolPerL = 3.3, got %f", alarms.MmolPerL)
	}
}

// TestSensorConfig_PublicFields tests that all fields are public
func TestSensorConfig_PublicFields(t *testing.T) {
	var sensor SensorConfig

	// If this compiles, fields are public
	sensor.SerialNumber = "test"
	sensor.Activation = time.Now()
	sensor.DeviceID = "test"
	sensor.SensorType = 1
	sensor.WarrantyDays = 1
	sensor.IsActive = true
	sensor.LowJourney = false
	sensor.DetectedAt = time.Now()

	// No assertions needed - compilation success is the test
}

// TestUserPreferences_PublicFields tests that all fields are public
func TestUserPreferences_PublicFields(t *testing.T) {
	var user UserPreferences

	// If this compiles, fields are public
	user.UserID = "test"
	user.FirstName = "test"
	user.LastName = "test"
	user.Email = "test@example.com"
	user.Country = "FR"
	user.AccountType = "pat"
	user.DateOfBirth = time.Now()
	user.Created = time.Now()
	user.LastLogin = time.Now()
	user.UILanguage = "fr"
	user.CommunicationLanguage = "fr"
	user.UnitOfMeasure = 0
	user.DateFormat = 1
	user.TimeFormat = 1
	user.EmailDays = []int{1}

	// No assertions needed - compilation success is the test
}

// TestDeviceInfo_PublicFields tests that all fields are public
func TestDeviceInfo_PublicFields(t *testing.T) {
	var device DeviceInfo

	// If this compiles, fields are public
	device.DeviceID = "test"
	device.DeviceTypeID = 1
	device.AppVersion = "1.0.0"
	device.AlarmsEnabled = true
	device.HighLimit = 180
	device.LowLimit = 70
	device.FixedLowThreshold = 70
	device.LastUpdate = time.Now()
	device.LimitEnabled = true

	// No assertions needed - compilation success is the test
}

// TestGlucoseTargets_PublicFields tests that all fields are public
func TestGlucoseTargets_PublicFields(t *testing.T) {
	var targets GlucoseTargets

	// If this compiles, fields are public
	targets.TargetHigh = 180
	targets.TargetLow = 70
	targets.UnitOfMeasure = 0

	// No assertions needed - compilation success is the test
}
