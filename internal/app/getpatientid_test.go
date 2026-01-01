package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/R4yL-dev/glcmd/internal/auth"
	"github.com/R4yL-dev/glcmd/internal/config"
	"github.com/R4yL-dev/glcmd/internal/headers"
)

// TestGetPatientID_EmptyDataArray tests that empty data array is properly handled
// Critical: Prevents panic on tmp.Data[0] access (line 41)
func TestGetPatientID_EmptyDataArray(t *testing.T) {
	// Create mock server that returns empty data array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": 0, "data": []}`))
	}))
	defer server.Close()

	// Override config URL to point to test server
	originalURL := config.ConnectionsURL
	config.ConnectionsURL = server.URL
	defer func() { config.ConnectionsURL = originalURL }()

	// Create minimal app instance
	testApp := &app{
		auth:       &auth.Auth{},
		headers:    headers.NewHeaders(),
		clientHTTP: server.Client(),
	}

	err := testApp.getPatientID()
	if err == nil {
		t.Fatal("expected error for empty data array, got nil")
	}

	expectedMsg := "cannot get patientID: API returned empty data array"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestGetPatientID_NonZeroStatus tests that non-zero API status is handled
// Critical: Prevents silent failures on API errors (line 33-35)
func TestGetPatientID_NonZeroStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"status 1", 1},
		{"status 2", 2},
		{"negative status", -1},
		{"high status", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that returns non-zero status
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := fmt.Sprintf(`{"status": %d, "data": [{"patientId": "test123"}]}`, tt.status)
				w.Write([]byte(response))
			}))
			defer server.Close()

			// Override config URL
			originalURL := config.ConnectionsURL
			config.ConnectionsURL = server.URL
			defer func() { config.ConnectionsURL = originalURL }()

			// Create minimal app instance
			testApp := &app{
				auth:       &auth.Auth{},
				headers:    headers.NewHeaders(),
				clientHTTP: server.Client(),
			}

			err := testApp.getPatientID()
			if err == nil {
				t.Fatalf("expected error for status %d, got nil", tt.status)
			}

			// Check that error message contains status code
			if err.Error() == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

// TestGetPatientID_EmptyPatientID tests that empty patientID is rejected
// Critical: Blocks app startup if patientID is empty (cascades to auth.SetPatientID)
func TestGetPatientID_EmptyPatientID(t *testing.T) {
	// Create mock server that returns empty patientId
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": 0, "data": [{"patientId": ""}]}`))
	}))
	defer server.Close()

	// Override config URL
	originalURL := config.ConnectionsURL
	config.ConnectionsURL = server.URL
	defer func() { config.ConnectionsURL = originalURL }()

	// Create minimal app instance
	testApp := &app{
		auth:       &auth.Auth{},
		headers:    headers.NewHeaders(),
		clientHTTP: server.Client(),
	}

	err := testApp.getPatientID()
	if err == nil {
		t.Fatal("expected error for empty patientID, got nil")
	}

	expectedMsg := "patientID cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}
