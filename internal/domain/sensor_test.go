package domain

import (
	"math"
	"testing"
	"time"
)

func TestStatus_EndedAt_Stopped(t *testing.T) {
	endedAt := time.Now().Add(-1 * time.Hour)
	s := &SensorConfig{
		Activation: time.Now().Add(-10 * 24 * time.Hour),
		ExpiresAt:  time.Now().Add(5 * 24 * time.Hour),
		EndedAt:    &endedAt,
	}

	if s.Status() != SensorStatusStopped {
		t.Errorf("expected stopped, got %s", s.Status())
	}
}

func TestStatus_Expired_Stopped(t *testing.T) {
	s := &SensorConfig{
		Activation: time.Now().Add(-20 * 24 * time.Hour),
		ExpiresAt:  time.Now().Add(-5 * 24 * time.Hour),
		EndedAt:    nil,
	}

	if s.Status() != SensorStatusStopped {
		t.Errorf("expected stopped, got %s", s.Status())
	}
}

func TestStatus_Unresponsive(t *testing.T) {
	lastMeasurement := time.Now().Add(-30 * time.Minute)
	s := &SensorConfig{
		Activation:        time.Now().Add(-5 * 24 * time.Hour),
		ExpiresAt:         time.Now().Add(10 * 24 * time.Hour),
		EndedAt:           nil,
		LastMeasurementAt: &lastMeasurement,
	}

	if s.Status() != SensorStatusUnresponsive {
		t.Errorf("expected unresponsive, got %s", s.Status())
	}
}

func TestStatus_Running(t *testing.T) {
	lastMeasurement := time.Now().Add(-5 * time.Minute)
	s := &SensorConfig{
		Activation:        time.Now().Add(-5 * 24 * time.Hour),
		ExpiresAt:         time.Now().Add(10 * 24 * time.Hour),
		EndedAt:           nil,
		LastMeasurementAt: &lastMeasurement,
	}

	if s.Status() != SensorStatusRunning {
		t.Errorf("expected running, got %s", s.Status())
	}
}

func TestStatus_Running_NoLastMeasurement(t *testing.T) {
	s := &SensorConfig{
		Activation: time.Now().Add(-5 * 24 * time.Hour),
		ExpiresAt:  time.Now().Add(10 * 24 * time.Hour),
		EndedAt:    nil,
	}

	if s.Status() != SensorStatusRunning {
		t.Errorf("expected running, got %s", s.Status())
	}
}

func TestElapsedDays_Expired_BoundedToExpiresAt(t *testing.T) {
	activation := time.Now().Add(-20 * 24 * time.Hour)
	expiresAt := activation.Add(15 * 24 * time.Hour)
	s := &SensorConfig{
		Activation:   activation,
		ExpiresAt:    expiresAt,
		DurationDays: 15,
		EndedAt:      nil,
	}

	elapsed := s.ElapsedDays()
	expected := 15.0

	if math.Abs(elapsed-expected) > 0.01 {
		t.Errorf("expected ElapsedDays ≈ %.1f, got %.1f", expected, elapsed)
	}
}

func TestElapsedDays_EndedAt_UsesEndedAt(t *testing.T) {
	activation := time.Now().Add(-10 * 24 * time.Hour)
	endedAt := activation.Add(8 * 24 * time.Hour)
	s := &SensorConfig{
		Activation:   activation,
		ExpiresAt:    activation.Add(15 * 24 * time.Hour),
		DurationDays: 15,
		EndedAt:      &endedAt,
	}

	elapsed := s.ElapsedDays()
	expected := 8.0

	if math.Abs(elapsed-expected) > 0.01 {
		t.Errorf("expected ElapsedDays ≈ %.1f, got %.1f", expected, elapsed)
	}
}

func TestElapsedDays_Running_UsesNow(t *testing.T) {
	activation := time.Now().Add(-5 * 24 * time.Hour)
	s := &SensorConfig{
		Activation:   activation,
		ExpiresAt:    activation.Add(15 * 24 * time.Hour),
		DurationDays: 15,
		EndedAt:      nil,
	}

	elapsed := s.ElapsedDays()
	expected := 5.0

	if math.Abs(elapsed-expected) > 0.1 {
		t.Errorf("expected ElapsedDays ≈ %.1f, got %.1f", expected, elapsed)
	}
}
