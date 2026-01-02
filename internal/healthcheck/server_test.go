package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Mock health status for testing
type mockHealthStatus struct {
	Status            string    `json:"status"`
	Timestamp         time.Time `json:"timestamp"`
	Uptime            string    `json:"uptime"`
	ConsecutiveErrors int       `json:"consecutiveErrors"`
	LastFetchError    string    `json:"lastFetchError"`
	LastFetchTime     time.Time `json:"lastFetchTime"`
}

func TestHandleHealth_Healthy(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{
			Status:            "healthy",
			Timestamp:         time.Now(),
			Uptime:            "1h30m",
			ConsecutiveErrors: 0,
			LastFetchError:    "",
			LastFetchTime:     time.Now(),
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Should return 200 OK for healthy status
	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", w.Code)
	}

	// Verify content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", contentType)
	}

	// Parse response
	var response mockHealthStatus
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status = healthy, got %s", response.Status)
	}

	if response.ConsecutiveErrors != 0 {
		t.Errorf("expected consecutiveErrors = 0, got %d", response.ConsecutiveErrors)
	}
}

func TestHandleHealth_Degraded(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{
			Status:            "degraded",
			Timestamp:         time.Now(),
			Uptime:            "2h",
			ConsecutiveErrors: 3,
			LastFetchError:    "network timeout",
			LastFetchTime:     time.Now().Add(-5 * time.Minute),
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Should return 503 Service Unavailable for degraded status
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status code 503, got %d", w.Code)
	}

	var response mockHealthStatus
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if response.Status != "degraded" {
		t.Errorf("expected status = degraded, got %s", response.Status)
	}

	if response.ConsecutiveErrors != 3 {
		t.Errorf("expected consecutiveErrors = 3, got %d", response.ConsecutiveErrors)
	}

	if response.LastFetchError != "network timeout" {
		t.Errorf("expected lastFetchError = 'network timeout', got %s", response.LastFetchError)
	}
}

func TestHandleHealth_Unhealthy(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{
			Status:            "unhealthy",
			Timestamp:         time.Now(),
			Uptime:            "3h",
			ConsecutiveErrors: 5,
			LastFetchError:    "authentication failed",
			LastFetchTime:     time.Now().Add(-30 * time.Minute),
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Should return 503 Service Unavailable for unhealthy status
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status code 503, got %d", w.Code)
	}

	var response mockHealthStatus
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if response.Status != "unhealthy" {
		t.Errorf("expected status = unhealthy, got %s", response.Status)
	}

	if response.ConsecutiveErrors != 5 {
		t.Errorf("expected consecutiveErrors = 5, got %d", response.ConsecutiveErrors)
	}
}

func TestHandleHealth_MethodNotAllowed(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{Status: "healthy"}
	})

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			server.handleHealth(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status code 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestHandleMetrics(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{Status: "healthy"}
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	// Should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", w.Code)
	}

	// Verify content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", contentType)
	}

	// Parse response
	var metrics map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &metrics); err != nil {
		t.Fatalf("failed to parse metrics JSON: %v", err)
	}

	// Verify expected fields are present
	if _, ok := metrics["uptime"]; !ok {
		t.Error("expected 'uptime' field in metrics")
	}

	if _, ok := metrics["goroutines"]; !ok {
		t.Error("expected 'goroutines' field in metrics")
	}

	if _, ok := metrics["memory"]; !ok {
		t.Error("expected 'memory' field in metrics")
	}

	if _, ok := metrics["runtime"]; !ok {
		t.Error("expected 'runtime' field in metrics")
	}

	if _, ok := metrics["process"]; !ok {
		t.Error("expected 'process' field in metrics")
	}
}

func TestHandleMetrics_MethodNotAllowed(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{Status: "healthy"}
	})

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/metrics", nil)
			w := httptest.NewRecorder()

			server.handleMetrics(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status code 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestHandleMetrics_MemoryFields(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{Status: "healthy"}
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	var metrics map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &metrics); err != nil {
		t.Fatalf("failed to parse metrics JSON: %v", err)
	}

	// Verify memory subfields
	memory, ok := metrics["memory"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'memory' to be a map")
	}

	expectedMemoryFields := []string{"allocMB", "totalAllocMB", "sysMB", "numGC"}
	for _, field := range expectedMemoryFields {
		if _, ok := memory[field]; !ok {
			t.Errorf("expected memory field '%s' to be present", field)
		}
	}
}

func TestHandleMetrics_RuntimeFields(t *testing.T) {
	server := NewServer(8080, func() interface{} {
		return mockHealthStatus{Status: "healthy"}
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	var metrics map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &metrics); err != nil {
		t.Fatalf("failed to parse metrics JSON: %v", err)
	}

	// Verify runtime subfields
	runtime, ok := metrics["runtime"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'runtime' to be a map")
	}

	expectedRuntimeFields := []string{"version", "os", "arch"}
	for _, field := range expectedRuntimeFields {
		if _, ok := runtime[field]; !ok {
			t.Errorf("expected runtime field '%s' to be present", field)
		}
	}
}

func TestNewServer(t *testing.T) {
	port := 9090
	callbackCalled := false

	server := NewServer(port, func() interface{} {
		callbackCalled = true
		return mockHealthStatus{Status: "healthy"}
	})

	if server == nil {
		t.Fatal("expected server to be created, got nil")
	}

	if server.port != port {
		t.Errorf("expected port = %d, got %d", port, server.port)
	}

	// Trigger callback
	server.getHealthStatus()

	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}
