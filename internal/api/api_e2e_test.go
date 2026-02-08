package api_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/R4yL-dev/glcmd/internal/api"
	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/repository"
	"github.com/R4yL-dev/glcmd/internal/service"
)

// setupE2ETest creates a test environment with in-memory database and API server
func setupE2ETest(t *testing.T) (http.Handler, *gorm.DB) {
	t.Helper()

	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
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
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create repositories
	measurementRepo := repository.NewGlucoseRepository(db)
	sensorRepo := repository.NewSensorRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	targetsRepo := repository.NewTargetsRepository(db)
	uow := repository.NewUnitOfWork(db)

	// Create services (nil event broker for tests)
	glucoseService := service.NewGlucoseService(measurementRepo, slog.Default(), nil)
	sensorService := service.NewSensorService(sensorRepo, uow, slog.Default(), nil)
	configService := service.NewConfigService(userRepo, deviceRepo, targetsRepo, slog.Default())

	// Create API server (nil event broker for tests)
	server := api.NewServer(
		8080,
		glucoseService,
		sensorService,
		configService,
		nil, // eventBroker
		func() daemon.HealthStatus {
			return daemon.HealthStatus{
				Status:            "healthy",
				Timestamp:         time.Now(),
				Uptime:            "1h",
				ConsecutiveErrors: 0,
				LastFetchTime:     time.Now(),
				DatabaseConnected: true,
				DataFresh:         true,
				FetchInterval:     "5m0s",
			}
		},
		func() bool { return true },
		nil, // getDatabasePoolStats
		slog.Default(),
	)

	// Return the HTTP handler from the server's httpServer field
	// We access the Handler field which contains the chi router
	return server.HTTPHandler(), db
}

// TestE2E_GetLatestMeasurement_NotFound tests getting latest measurement from empty database
func TestE2E_GetLatestMeasurement_NotFound(t *testing.T) {
	server, _ := setupE2ETest(t)

	req := httptest.NewRequest("GET", "/v1/glucose/latest", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["error"] == nil {
		t.Error("expected error field in response")
	}
}

// TestE2E_SaveAndGetMeasurement tests saving and retrieving a measurement
func TestE2E_SaveAndGetMeasurement(t *testing.T) {
	server, db := setupE2ETest(t)

	// Insert test measurement
	measurement := &domain.GlucoseMeasurement{
		Timestamp:        time.Now().UTC(),
		Value:            5.5,
		ValueInMgPerDl:   99,
		GlucoseColor: domain.GlucoseColorNormal,
		Type:             domain.GlucoseTypeCurrent,
	}
	if err := db.Create(measurement).Error; err != nil {
		t.Fatalf("failed to insert test measurement: %v", err)
	}

	// GET via API
	req := httptest.NewRequest("GET", "/v1/glucose/latest", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response api.MeasurementResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Data.Value != 5.5 {
		t.Errorf("expected value 5.5, got %f", response.Data.Value)
	}
	if response.Data.ValueInMgPerDl != 99 {
		t.Errorf("expected value 99 mg/dL, got %d", response.Data.ValueInMgPerDl)
	}
	if response.Data.GlucoseColor != domain.GlucoseColorNormal {
		t.Errorf("expected color %d, got %d", domain.GlucoseColorNormal, response.Data.GlucoseColor)
	}
}

// TestE2E_GetMeasurements_WithPagination tests pagination
func TestE2E_GetMeasurements_WithPagination(t *testing.T) {
	server, db := setupE2ETest(t)

	// Insert 5 test measurements
	for i := 0; i < 5; i++ {
		measurement := &domain.GlucoseMeasurement{
			Timestamp:        time.Now().UTC().Add(time.Duration(-i) * time.Hour),
			Value:            5.0 + float64(i)*0.1,
			ValueInMgPerDl:   90 + i,
			GlucoseColor: domain.GlucoseColorNormal,
			Type:             domain.GlucoseTypeCurrent,
		}
		if err := db.Create(measurement).Error; err != nil {
			t.Fatalf("failed to insert test measurement: %v", err)
		}
	}

	// GET first page (limit=2)
	req := httptest.NewRequest("GET", "/v1/glucose?limit=2&offset=0", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response api.MeasurementListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(response.Data))
	}

	if response.Pagination.Total != 5 {
		t.Errorf("expected total 5, got %d", response.Pagination.Total)
	}

	if !response.Pagination.HasMore {
		t.Error("expected hasMore to be true")
	}

	// GET second page (limit=2, offset=2)
	req = httptest.NewRequest("GET", "/v1/glucose?limit=2&offset=2", nil)
	w = httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Errorf("expected 2 measurements on page 2, got %d", len(response.Data))
	}

	if !response.Pagination.HasMore {
		t.Error("expected hasMore to be true (1 item left)")
	}
}

// TestE2E_GetStatistics_WithData tests statistics calculation
func TestE2E_GetStatistics_WithData(t *testing.T) {
	server, db := setupE2ETest(t)

	// Insert glucose targets (in mg/dL)
	targets := &domain.GlucoseTargets{
		TargetLow:     72,  // 4.0 mmol/L = 72 mg/dL
		TargetHigh:    126, // 7.0 mmol/L = 126 mg/dL
		UnitOfMeasure: domain.GlucoseUnitsMgDl,
	}
	if err := db.Create(targets).Error; err != nil {
		t.Fatalf("failed to insert targets: %v", err)
	}

	// Insert measurements with different colors (use UTC time)
	now := time.Now().UTC()
	measurements := []*domain.GlucoseMeasurement{
		{Timestamp: now.Add(-3 * time.Hour), Value: 5.0, ValueInMgPerDl: 90, GlucoseColor: domain.GlucoseColorNormal, Type: domain.GlucoseTypeCurrent},
		{Timestamp: now.Add(-2 * time.Hour), Value: 8.5, ValueInMgPerDl: 153, GlucoseColor: domain.GlucoseColorWarning, IsHigh: true, Type: domain.GlucoseTypeCurrent},
		{Timestamp: now.Add(-1 * time.Hour), Value: 5.2, ValueInMgPerDl: 94, GlucoseColor: domain.GlucoseColorNormal, Type: domain.GlucoseTypeCurrent},
	}
	for _, m := range measurements {
		if err := db.Create(m).Error; err != nil {
			t.Fatalf("failed to insert measurement: %v", err)
		}
	}

	// GET statistics
	start := now.Add(-4 * time.Hour).Format(time.RFC3339)
	end := now.Format(time.RFC3339)

	req := httptest.NewRequest("GET", "/v1/glucose/stats?start="+start+"&end="+end, nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.StatisticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Data.Statistics.Count != 3 {
		t.Errorf("expected count 3, got %d", response.Data.Statistics.Count)
	}

	if response.Data.Distribution.Normal != 2 {
		t.Errorf("expected 2 normal measurements, got %d", response.Data.Distribution.Normal)
	}

	if response.Data.Distribution.High != 1 {
		t.Errorf("expected 1 high measurement, got %d", response.Data.Distribution.High)
	}

	// Verify Time in Range data is present
	if response.Data.TimeInRange == nil {
		t.Error("expected TimeInRange data, got nil")
	} else {
		if response.Data.TimeInRange.TargetLowMgDl != 72 {
			t.Errorf("expected target low 72 mg/dL, got %d", response.Data.TimeInRange.TargetLowMgDl)
		}
		if response.Data.TimeInRange.TargetHighMgDl != 126 {
			t.Errorf("expected target high 126 mg/dL, got %d", response.Data.TimeInRange.TargetHighMgDl)
		}
	}
}

// TestE2E_GetStatistics_InvalidTimeRange tests validation of time range
func TestE2E_GetStatistics_InvalidTimeRange(t *testing.T) {
	server, _ := setupE2ETest(t)

	// end < start (invalid)
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	req := httptest.NewRequest("GET", "/v1/glucose/stats?start="+start+"&end="+end, nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestE2E_GetStatistics_LargeTimeRange tests that large time ranges work (no 90-day limit)
func TestE2E_GetStatistics_LargeTimeRange(t *testing.T) {
	server, _ := setupE2ETest(t)

	// Range > 90 days should now work (no limit)
	start := time.Now().UTC().Add(-100 * 24 * time.Hour).Format(time.RFC3339)
	end := time.Now().UTC().Format(time.RFC3339)

	req := httptest.NewRequest("GET", "/v1/glucose/stats?start="+start+"&end="+end, nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestE2E_GetStatistics_AllTime tests statistics without time bounds (all time)
func TestE2E_GetStatistics_AllTime(t *testing.T) {
	server, db := setupE2ETest(t)

	now := time.Now().UTC()

	// Insert test measurement
	measurement := &domain.GlucoseMeasurement{
		Timestamp:        now.Add(-1 * time.Hour),
		Value:            7.0,
		ValueInMgPerDl:   126,
		Type:             domain.GlucoseTypeCurrent,
		GlucoseColor: domain.GlucoseColorNormal,
	}
	db.Create(measurement)

	// Query without start/end (all time)
	req := httptest.NewRequest("GET", "/v1/glucose/stats", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.StatisticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Data.Statistics.Count != 1 {
		t.Errorf("expected count 1, got %d", response.Data.Statistics.Count)
	}

	// Period should contain actual data bounds for all-time queries
	if response.Data.Period.Start == "" || response.Data.Period.End == "" {
		t.Errorf("expected period with actual data bounds, got start=%s end=%s",
			response.Data.Period.Start, response.Data.Period.End)
	}
}

// TestE2E_GetSensor tests sensor listing
func TestE2E_GetSensor(t *testing.T) {
	server, db := setupE2ETest(t)

	now := time.Now().UTC()
	endedAt := now.Add(-2 * 24 * time.Hour)

	// Insert test sensors
	sensors := []*domain.SensorConfig{
		{
			SerialNumber: "SENSOR001",
			Activation:   now.Add(-22 * 24 * time.Hour),
			ExpiresAt:    now.Add(-7 * 24 * time.Hour),
			EndedAt:      &endedAt, // Ended sensor
			SensorType:   4,
			DurationDays: 15,
			DetectedAt:   now.Add(-22 * 24 * time.Hour),
		},
		{
			SerialNumber: "SENSOR002",
			Activation:   now.Add(-2 * 24 * time.Hour),
			ExpiresAt:    now.Add(13 * 24 * time.Hour),
			EndedAt:      nil, // Current sensor
			SensorType:   4,
			DurationDays: 15,
			DetectedAt:   now.Add(-2 * 24 * time.Hour),
		},
	}
	for _, s := range sensors {
		if err := db.Create(s).Error; err != nil {
			t.Fatalf("failed to insert sensor: %v", err)
		}
	}

	// GET sensors (paginated list)
	req := httptest.NewRequest("GET", "/v1/sensor", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response api.SensorListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify all sensors are returned
	if len(response.Data) != 2 {
		t.Errorf("expected 2 sensors, got %d", len(response.Data))
	}

	if response.Pagination.Total != 2 {
		t.Errorf("expected total 2, got %d", response.Pagination.Total)
	}
}

// TestE2E_GetLatestSensor tests getting the current sensor
func TestE2E_GetLatestSensor(t *testing.T) {
	server, db := setupE2ETest(t)

	now := time.Now().UTC()

	// Insert current sensor
	sensor := &domain.SensorConfig{
		SerialNumber: "SENSOR001",
		Activation:   now.Add(-2 * 24 * time.Hour),
		ExpiresAt:    now.Add(12 * 24 * time.Hour),
		EndedAt:      nil, // Current sensor
		SensorType:   4,
		DurationDays: 14,
		DetectedAt:   now.Add(-2 * 24 * time.Hour),
	}
	if err := db.Create(sensor).Error; err != nil {
		t.Fatalf("failed to insert sensor: %v", err)
	}

	// GET latest sensor
	req := httptest.NewRequest("GET", "/v1/sensor/latest", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response api.LatestSensorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Data == nil {
		t.Fatal("expected sensor data, got nil")
	}

	if response.Data.SerialNumber != "SENSOR001" {
		t.Errorf("expected SENSOR001, got %s", response.Data.SerialNumber)
	}

	if !response.Data.IsActive {
		t.Error("expected sensor to be active")
	}
}

// TestE2E_GetLatestSensor_NotFound tests getting latest sensor when none exists
func TestE2E_GetLatestSensor_NotFound(t *testing.T) {
	server, _ := setupE2ETest(t)

	req := httptest.NewRequest("GET", "/v1/sensor/latest", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// TestE2E_Health tests health endpoint
func TestE2E_Health(t *testing.T) {
	server, _ := setupE2ETest(t)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response api.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Data.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", response.Data.Status)
	}

	if !response.Data.DatabaseConnected {
		t.Error("expected database connected")
	}

	if !response.Data.DataFresh {
		t.Error("expected DataFresh = true")
	}

	if response.Data.FetchInterval != "5m0s" {
		t.Errorf("expected FetchInterval = 5m0s, got %s", response.Data.FetchInterval)
	}
}

// TestE2E_Metrics tests metrics endpoint
func TestE2E_Metrics(t *testing.T) {
	server, _ := setupE2ETest(t)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response api.MetricsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Data.Goroutines <= 0 {
		t.Errorf("expected positive goroutines count, got %d", response.Data.Goroutines)
	}

	if response.Data.Runtime.Version == "" {
		t.Error("expected runtime version, got empty string")
	}

	if response.Data.Process.PID <= 0 {
		t.Errorf("expected positive PID, got %d", response.Data.Process.PID)
	}

	// SSE metrics (broker is nil in tests)
	if response.Data.SSE.Enabled {
		t.Error("expected SSE.Enabled = false (no broker)")
	}

	if response.Data.SSE.Subscribers != 0 {
		t.Errorf("expected SSE.Subscribers = 0, got %d", response.Data.SSE.Subscribers)
	}

	// Database pool stats (nil callback in tests)
	if response.Data.Database != nil {
		t.Error("expected Database = nil (no callback)")
	}
}

// TestE2E_CORS_Preflight tests CORS preflight request
func TestE2E_CORS_Preflight(t *testing.T) {
	server, _ := setupE2ETest(t)

	req := httptest.NewRequest("OPTIONS", "/v1/glucose/latest", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Verify CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

// TestE2E_CORS_Headers tests CORS headers on actual request
func TestE2E_CORS_Headers(t *testing.T) {
	server, _ := setupE2ETest(t)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Verify CORS headers are present
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
