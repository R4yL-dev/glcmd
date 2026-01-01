package libreclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Test with nil httpClient (should create default)
	client := NewClient(nil)
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be set")
	}

	if client.baseURL != BaseURL {
		t.Errorf("expected baseURL = %s, got %s", BaseURL, client.baseURL)
	}

	// Test with custom httpClient
	customClient := &http.Client{Timeout: 5 * time.Second}
	client = NewClient(customClient)
	if client.httpClient != customClient {
		t.Error("expected custom httpClient to be used")
	}
}

func TestAuthenticate_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/llu/auth/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify request body
		var creds AuthCredentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if creds.Email != "test@example.com" {
			t.Errorf("expected email = 'test@example.com', got %s", creds.Email)
		}

		// Send response
		response := AuthResponse{}
		response.Data.User.ID = "test-user-123"
		response.Data.AuthTicket.Token = "test-token-456"
		response.Data.AuthTicket.Expires = 1234567890
		response.Data.AuthTicket.Duration = 3600

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.baseURL = server.URL

	ctx := context.Background()
	token, userID, accountID, err := client.Authenticate(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "test-token-456" {
		t.Errorf("expected token = 'test-token-456', got %s", token)
	}

	if userID != "test-user-123" {
		t.Errorf("expected userID = 'test-user-123', got %s", userID)
	}

	if accountID == "" {
		t.Error("expected accountID to be set")
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid credentials"}`))
	}))
	defer server.Close()

	client := NewClient(nil)
	client.baseURL = server.URL

	ctx := context.Background()
	_, _, _, err := client.Authenticate(ctx, "wrong@example.com", "wrongpass")
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}

	// Should be an AuthError
	if _, ok := err.(*AuthError); !ok {
		t.Errorf("expected AuthError, got %T", err)
	}
}

func TestGetConnections_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/llu/connections" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify auth headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing or invalid Authorization header")
		}

		response := ConnectionsResponse{}
		response.Data = append(response.Data, struct {
			PatientID string `json:"patientId"`
			GlucoseMeasurement struct {
				ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
				Value            float64 `json:"Value"`
				TrendArrow       int     `json:"TrendArrow"`
				TrendMessage     string  `json:"TrendMessage"`
				MeasurementColor int     `json:"MeasurementColor"`
				GlucoseUnits     int     `json:"GlucoseUnits"`
				Timestamp        string  `json:"Timestamp"`
				IsHigh           bool    `json:"isHigh"`
				IsLow            bool    `json:"isLow"`
			} `json:"glucoseMeasurement"`
		}{
			PatientID: "patient-123",
		})
		response.Data[0].GlucoseMeasurement.Value = 5.5
		response.Data[0].GlucoseMeasurement.ValueInMgPerDl = 100

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.baseURL = server.URL

	ctx := context.Background()
	result, err := client.GetConnections(ctx, "test-token", "test-account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Errorf("expected 1 data entry, got %d", len(result.Data))
	}

	if result.Data[0].PatientID != "patient-123" {
		t.Errorf("expected patientID = 'patient-123', got %s", result.Data[0].PatientID)
	}

	if result.Data[0].GlucoseMeasurement.Value != 5.5 {
		t.Errorf("expected Value = 5.5, got %f", result.Data[0].GlucoseMeasurement.Value)
	}
}

func TestGetGraph_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/llu/connections/patient-123/graph"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: %s, expected %s", r.URL.Path, expectedPath)
		}

		response := GraphResponse{}
		response.Data.Connection.GlucoseMeasurement.Value = 6.2
		response.Data.Connection.Sensor.SN = "ABC123"
		response.Data.GraphData = append(response.Data.GraphData, struct {
			FactoryTimestamp string  `json:"FactoryTimestamp"`
			Timestamp        string  `json:"Timestamp"`
			ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
			Value            float64 `json:"Value"`
			MeasurementColor int     `json:"MeasurementColor"`
			GlucoseUnits     int     `json:"GlucoseUnits"`
			IsHigh           bool    `json:"isHigh"`
			IsLow            bool    `json:"isLow"`
			Type             int     `json:"type"`
		}{
			FactoryTimestamp: "1/1/2026 1:00:00 PM",
			Timestamp:        "1/1/2026 2:00:00 PM",
			Value:            5.8,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.baseURL = server.URL

	ctx := context.Background()
	result, err := client.GetGraph(ctx, "test-token", "test-account", "patient-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.Connection.GlucoseMeasurement.Value != 6.2 {
		t.Errorf("expected current Value = 6.2, got %f", result.Data.Connection.GlucoseMeasurement.Value)
	}

	if result.Data.Connection.Sensor.SN != "ABC123" {
		t.Errorf("expected SN = 'ABC123', got %s", result.Data.Connection.Sensor.SN)
	}

	if len(result.Data.GraphData) != 1 {
		t.Errorf("expected 1 graph data point, got %d", len(result.Data.GraphData))
	}

	if result.Data.GraphData[0].Value != 5.8 {
		t.Errorf("expected historical Value = 5.8, got %f", result.Data.GraphData[0].Value)
	}
}

func TestContextCancellation(t *testing.T) {
	// Create server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.baseURL = server.URL

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, _, err := client.Authenticate(ctx, "test@example.com", "password")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}

	// Should be a NetworkError
	if _, ok := err.(*NetworkError); !ok {
		t.Errorf("expected NetworkError, got %T", err)
	}
}
