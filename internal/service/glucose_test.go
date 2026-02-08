package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
	"github.com/R4yL-dev/glcmd/internal/repository"
)

// MockGlucoseRepository for testing
type MockGlucoseRepository struct {
	SaveFunc             func(ctx context.Context, m *domain.GlucoseMeasurement) (bool, error)
	FindLatestFunc       func(ctx context.Context) (*domain.GlucoseMeasurement, error)
	FindAllFunc          func(ctx context.Context) ([]*domain.GlucoseMeasurement, error)
	FindByTimeRangeFunc  func(ctx context.Context, start, end time.Time) ([]*domain.GlucoseMeasurement, error)
	FindWithFiltersFunc  func(ctx context.Context, filters repository.GlucoseFilters, limit, offset int) ([]*domain.GlucoseMeasurement, error)
	CountWithFiltersFunc func(ctx context.Context, filters repository.GlucoseFilters) (int64, error)
	GetStatisticsFunc    func(ctx context.Context, filters repository.GlucoseStatisticsFilters) (*repository.GlucoseStatisticsResult, error)
}

func (m *MockGlucoseRepository) Save(ctx context.Context, measurement *domain.GlucoseMeasurement) (bool, error) {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, measurement)
	}
	return true, nil
}

func (m *MockGlucoseRepository) FindLatest(ctx context.Context) (*domain.GlucoseMeasurement, error) {
	if m.FindLatestFunc != nil {
		return m.FindLatestFunc(ctx)
	}
	return nil, persistence.ErrNotFound
}

func (m *MockGlucoseRepository) FindAll(ctx context.Context) ([]*domain.GlucoseMeasurement, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx)
	}
	return []*domain.GlucoseMeasurement{}, nil
}

func (m *MockGlucoseRepository) FindByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.GlucoseMeasurement, error) {
	if m.FindByTimeRangeFunc != nil {
		return m.FindByTimeRangeFunc(ctx, start, end)
	}
	return []*domain.GlucoseMeasurement{}, nil
}

func (m *MockGlucoseRepository) FindWithFilters(ctx context.Context, filters repository.GlucoseFilters, limit, offset int) ([]*domain.GlucoseMeasurement, error) {
	if m.FindWithFiltersFunc != nil {
		return m.FindWithFiltersFunc(ctx, filters, limit, offset)
	}
	return []*domain.GlucoseMeasurement{}, nil
}

func (m *MockGlucoseRepository) CountWithFilters(ctx context.Context, filters repository.GlucoseFilters) (int64, error) {
	if m.CountWithFiltersFunc != nil {
		return m.CountWithFiltersFunc(ctx, filters)
	}
	return 0, nil
}

func (m *MockGlucoseRepository) GetStatistics(ctx context.Context, filters repository.GlucoseStatisticsFilters) (*repository.GlucoseStatisticsResult, error) {
	if m.GetStatisticsFunc != nil {
		return m.GetStatisticsFunc(ctx, filters)
	}
	return &repository.GlucoseStatisticsResult{}, nil
}

func TestGlucoseService_SaveMeasurement_Success(t *testing.T) {
	saveCalled := false

	mockRepo := &MockGlucoseRepository{
		SaveFunc: func(ctx context.Context, m *domain.GlucoseMeasurement) (bool, error) {
			saveCalled = true
			if m.Value != 7.5 {
				t.Errorf("expected Value = 7.5, got %f", m.Value)
			}
			return true, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurement := &domain.GlucoseMeasurement{
		Timestamp:      time.Now(),
		Value:          7.5,
		ValueInMgPerDl: 135,
		Type:           domain.GlucoseTypeCurrent,
	}

	inserted, err := service.SaveMeasurement(context.Background(), measurement)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !saveCalled {
		t.Error("expected Save to be called")
	}

	if !inserted {
		t.Error("expected inserted to be true")
	}
}

func TestGlucoseService_SaveMeasurement_RetryOnTransientError(t *testing.T) {
	attemptCount := 0

	mockRepo := &MockGlucoseRepository{
		SaveFunc: func(ctx context.Context, m *domain.GlucoseMeasurement) (bool, error) {
			attemptCount++
			// Fail first attempt, succeed on second
			if attemptCount == 1 {
				// Simulate a retryable error (database locked)
				return false, errors.New("database is locked")
			}
			return true, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurement := &domain.GlucoseMeasurement{
		Timestamp: time.Now(),
		Value:     6.2,
		Type:      domain.GlucoseTypeCurrent,
	}

	_, err := service.SaveMeasurement(context.Background(), measurement)
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}

	// Should have retried after first failure
	if attemptCount < 2 {
		t.Errorf("expected at least 2 attempts, got %d", attemptCount)
	}
}

func TestGlucoseService_SaveMeasurement_FailureAfterRetries(t *testing.T) {
	persistentError := errors.New("persistent database error")

	mockRepo := &MockGlucoseRepository{
		SaveFunc: func(ctx context.Context, m *domain.GlucoseMeasurement) (bool, error) {
			// Always fail
			return false, persistentError
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurement := &domain.GlucoseMeasurement{
		Timestamp: time.Now(),
		Value:     5.0,
		Type:      domain.GlucoseTypeCurrent,
	}

	_, err := service.SaveMeasurement(context.Background(), measurement)
	if err == nil {
		t.Fatal("expected error after retries, got nil")
	}
}

func TestGlucoseService_GetLatestMeasurement_Success(t *testing.T) {
	expectedMeasurement := &domain.GlucoseMeasurement{
		ID:             1,
		Timestamp:      time.Now(),
		Value:          8.5,
		ValueInMgPerDl: 153,
		Type:           domain.GlucoseTypeCurrent,
	}

	mockRepo := &MockGlucoseRepository{
		FindLatestFunc: func(ctx context.Context) (*domain.GlucoseMeasurement, error) {
			return expectedMeasurement, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurement, err := service.GetLatestMeasurement(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if measurement == nil {
		t.Fatal("expected measurement, got nil")
	}

	if measurement.Value != 8.5 {
		t.Errorf("expected Value = 8.5, got %f", measurement.Value)
	}

	if measurement.ID != 1 {
		t.Errorf("expected ID = 1, got %d", measurement.ID)
	}
}

func TestGlucoseService_GetLatestMeasurement_NotFound(t *testing.T) {
	mockRepo := &MockGlucoseRepository{
		FindLatestFunc: func(ctx context.Context) (*domain.GlucoseMeasurement, error) {
			return nil, persistence.ErrNotFound
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurement, err := service.GetLatestMeasurement(context.Background())
	if err == nil {
		t.Fatal("expected ErrNotFound, got nil")
	}

	if !errors.Is(err, persistence.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	if measurement != nil {
		t.Error("expected nil measurement when not found")
	}
}

func TestGlucoseService_GetAllMeasurements_Success(t *testing.T) {
	expectedMeasurements := []*domain.GlucoseMeasurement{
		{ID: 1, Value: 7.0, Timestamp: time.Now().Add(-2 * time.Hour)},
		{ID: 2, Value: 7.5, Timestamp: time.Now().Add(-1 * time.Hour)},
		{ID: 3, Value: 8.0, Timestamp: time.Now()},
	}

	mockRepo := &MockGlucoseRepository{
		FindAllFunc: func(ctx context.Context) ([]*domain.GlucoseMeasurement, error) {
			return expectedMeasurements, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurements, err := service.GetAllMeasurements(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(measurements) != 3 {
		t.Errorf("expected 3 measurements, got %d", len(measurements))
	}
}

func TestGlucoseService_GetAllMeasurements_Empty(t *testing.T) {
	mockRepo := &MockGlucoseRepository{
		FindAllFunc: func(ctx context.Context) ([]*domain.GlucoseMeasurement, error) {
			return []*domain.GlucoseMeasurement{}, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurements, err := service.GetAllMeasurements(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(measurements))
	}
}

func TestGlucoseService_GetMeasurementsByTimeRange_Success(t *testing.T) {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	expectedMeasurements := []*domain.GlucoseMeasurement{
		{ID: 1, Value: 6.5, Timestamp: start.Add(1 * time.Hour)},
		{ID: 2, Value: 7.0, Timestamp: start.Add(6 * time.Hour)},
		{ID: 3, Value: 7.5, Timestamp: start.Add(12 * time.Hour)},
	}

	mockRepo := &MockGlucoseRepository{
		FindByTimeRangeFunc: func(ctx context.Context, s, e time.Time) ([]*domain.GlucoseMeasurement, error) {
			// Verify correct time range passed
			if !s.Equal(start) {
				t.Errorf("expected start = %v, got %v", start, s)
			}
			if !e.Equal(end) {
				t.Errorf("expected end = %v, got %v", end, e)
			}
			return expectedMeasurements, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurements, err := service.GetMeasurementsByTimeRange(context.Background(), start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(measurements) != 3 {
		t.Errorf("expected 3 measurements, got %d", len(measurements))
	}
}

func TestGlucoseService_GetMeasurementsByTimeRange_Empty(t *testing.T) {
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	mockRepo := &MockGlucoseRepository{
		FindByTimeRangeFunc: func(ctx context.Context, s, e time.Time) ([]*domain.GlucoseMeasurement, error) {
			// No measurements in range
			return []*domain.GlucoseMeasurement{}, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	measurements, err := service.GetMeasurementsByTimeRange(context.Background(), start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(measurements))
	}
}

func TestGlucoseService_SaveMeasurement_ValidatesType(t *testing.T) {
	mockRepo := &MockGlucoseRepository{
		SaveFunc: func(ctx context.Context, m *domain.GlucoseMeasurement) (bool, error) {
			// Verify measurement type is valid
			if m.Type != domain.GlucoseTypeCurrent && m.Type != domain.GlucoseTypeHistorical {
				t.Errorf("invalid measurement type: %d", m.Type)
			}
			return true, nil
		},
	}

	service := NewGlucoseService(mockRepo, slog.Default(), nil)

	tests := []struct {
		name string
		typ  int
	}{
		{"Current measurement", domain.GlucoseTypeCurrent},
		{"Historical measurement", domain.GlucoseTypeHistorical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			measurement := &domain.GlucoseMeasurement{
				Timestamp: time.Now(),
				Value:     7.0,
				Type:      tt.typ,
			}

			_, err := service.SaveMeasurement(context.Background(), measurement)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
