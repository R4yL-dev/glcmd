package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// Mock implementations

type MockSensorRepository struct {
	FindActiveFunc           func(ctx context.Context) (*domain.SensorConfig, error)
	SaveFunc                 func(ctx context.Context, s *domain.SensorConfig) error
	UpdateActiveStatusFunc   func(ctx context.Context, serial string, active bool) error
	FindAllFunc              func(ctx context.Context) ([]*domain.SensorConfig, error)
	FindBySerialNumberFunc   func(ctx context.Context, serial string) (*domain.SensorConfig, error)
}

func (m *MockSensorRepository) FindActive(ctx context.Context) (*domain.SensorConfig, error) {
	if m.FindActiveFunc != nil {
		return m.FindActiveFunc(ctx)
	}
	return nil, persistence.ErrNotFound
}

func (m *MockSensorRepository) Save(ctx context.Context, s *domain.SensorConfig) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, s)
	}
	return nil
}

func (m *MockSensorRepository) UpdateActiveStatus(ctx context.Context, serial string, active bool) error {
	if m.UpdateActiveStatusFunc != nil {
		return m.UpdateActiveStatusFunc(ctx, serial, active)
	}
	return nil
}

func (m *MockSensorRepository) FindAll(ctx context.Context) ([]*domain.SensorConfig, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx)
	}
	return []*domain.SensorConfig{}, nil
}

func (m *MockSensorRepository) FindBySerialNumber(ctx context.Context, serial string) (*domain.SensorConfig, error) {
	if m.FindBySerialNumberFunc != nil {
		return m.FindBySerialNumberFunc(ctx, serial)
	}
	return nil, persistence.ErrNotFound
}

type MockUnitOfWork struct {
	ExecuteInTransactionFunc func(ctx context.Context, fn func(txCtx context.Context) error) error
}

func (m *MockUnitOfWork) ExecuteInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	if m.ExecuteInTransactionFunc != nil {
		return m.ExecuteInTransactionFunc(ctx, fn)
	}
	// Default: just execute the function directly (no real transaction)
	return fn(ctx)
}

// Tests

func TestSensorService_HandleSensorChange_FirstSensor(t *testing.T) {
	mockRepo := &MockSensorRepository{
		FindActiveFunc: func(ctx context.Context) (*domain.SensorConfig, error) {
			// No active sensor exists
			return nil, persistence.ErrNotFound
		},
		SaveFunc: func(ctx context.Context, s *domain.SensorConfig) error {
			if s.SerialNumber != "FIRST_SENSOR" {
				t.Errorf("expected SerialNumber = FIRST_SENSOR, got %s", s.SerialNumber)
			}
			return nil
		},
	}

	mockUoW := &MockUnitOfWork{}

	service := NewSensorService(mockRepo, mockUoW, slog.Default())

	newSensor := &domain.SensorConfig{
		SerialNumber: "FIRST_SENSOR",
		IsActive:     true,
		DetectedAt:   time.Now().UTC(),
	}

	err := service.HandleSensorChange(context.Background(), newSensor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSensorService_HandleSensorChange_SensorChanged(t *testing.T) {
	oldSensor := &domain.SensorConfig{
		SerialNumber: "OLD_SENSOR",
		IsActive:     true,
	}

	deactivateCalled := false
	saveCalled := false

	mockRepo := &MockSensorRepository{
		FindActiveFunc: func(ctx context.Context) (*domain.SensorConfig, error) {
			return oldSensor, nil
		},
		UpdateActiveStatusFunc: func(ctx context.Context, serial string, active bool) error {
			if serial != "OLD_SENSOR" {
				t.Errorf("expected serial = OLD_SENSOR, got %s", serial)
			}
			if active != false {
				t.Error("expected active = false")
			}
			deactivateCalled = true
			return nil
		},
		SaveFunc: func(ctx context.Context, s *domain.SensorConfig) error {
			if s.SerialNumber != "NEW_SENSOR" {
				t.Errorf("expected SerialNumber = NEW_SENSOR, got %s", s.SerialNumber)
			}
			saveCalled = true
			return nil
		},
	}

	mockUoW := &MockUnitOfWork{}

	service := NewSensorService(mockRepo, mockUoW, slog.Default())

	newSensor := &domain.SensorConfig{
		SerialNumber: "NEW_SENSOR",
		IsActive:     true,
		DetectedAt:   time.Now().UTC(),
	}

	err := service.HandleSensorChange(context.Background(), newSensor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deactivateCalled {
		t.Error("expected old sensor to be deactivated")
	}

	if !saveCalled {
		t.Error("expected new sensor to be saved")
	}
}

func TestSensorService_HandleSensorChange_SameSensor(t *testing.T) {
	existingSensor := &domain.SensorConfig{
		SerialNumber: "SAME_SENSOR",
		IsActive:     true,
	}

	deactivateCalled := false

	mockRepo := &MockSensorRepository{
		FindActiveFunc: func(ctx context.Context) (*domain.SensorConfig, error) {
			return existingSensor, nil
		},
		UpdateActiveStatusFunc: func(ctx context.Context, serial string, active bool) error {
			deactivateCalled = true
			return nil
		},
		SaveFunc: func(ctx context.Context, s *domain.SensorConfig) error {
			// Should still save (upsert)
			return nil
		},
	}

	mockUoW := &MockUnitOfWork{}

	service := NewSensorService(mockRepo, mockUoW, slog.Default())

	sameSensor := &domain.SensorConfig{
		SerialNumber: "SAME_SENSOR", // Same serial number
		IsActive:     true,
		DetectedAt:   time.Now().UTC(),
	}

	err := service.HandleSensorChange(context.Background(), sameSensor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT deactivate if same sensor
	if deactivateCalled {
		t.Error("expected old sensor to NOT be deactivated (same sensor)")
	}
}

func TestSensorService_HandleSensorChange_TransactionRollback(t *testing.T) {
	oldSensor := &domain.SensorConfig{
		SerialNumber: "OLD_SENSOR",
		IsActive:     true,
	}

	mockRepo := &MockSensorRepository{
		FindActiveFunc: func(ctx context.Context) (*domain.SensorConfig, error) {
			return oldSensor, nil
		},
		UpdateActiveStatusFunc: func(ctx context.Context, serial string, active bool) error {
			// Deactivation succeeds
			return nil
		},
		SaveFunc: func(ctx context.Context, s *domain.SensorConfig) error {
			// Save fails - should trigger rollback
			return errors.New("database error")
		},
	}

	transactionExecuted := false
	mockUoW := &MockUnitOfWork{
		ExecuteInTransactionFunc: func(ctx context.Context, fn func(txCtx context.Context) error) error {
			transactionExecuted = true
			err := fn(ctx)
			// Simulate rollback on error
			if err != nil {
				return err
			}
			return nil
		},
	}

	service := NewSensorService(mockRepo, mockUoW, slog.Default())

	newSensor := &domain.SensorConfig{
		SerialNumber: "NEW_SENSOR",
		IsActive:     true,
	}

	err := service.HandleSensorChange(context.Background(), newSensor)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !transactionExecuted {
		t.Error("expected transaction to be executed")
	}
}
