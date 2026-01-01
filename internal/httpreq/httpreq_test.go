package httpreq

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSetMethod_InvalidMethod tests that invalid HTTP methods are rejected
// Critical: Prevents security issues by only allowing GET/POST (lines 13-21)
func TestSetMethod_InvalidMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
		{"PATCH", http.MethodPatch},
		{"HEAD", http.MethodHead},
		{"OPTIONS", http.MethodOptions},
		{"empty string", ""},
		{"invalid method", "INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &httpReq{}
			err := req.SetMethod(tt.method)
			if err == nil {
				t.Fatalf("expected error for method %q, got nil", tt.method)
			}
		})
	}
}

// TestSetUrl_DNSLookupFailure tests that URLs failing DNS lookup are rejected
// Critical: Prevents blocking/timeout on bad URLs (lines 36-40 in utils.go)
func TestSetUrl_DNSLookupFailure(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"non-existent domain", "https://this-domain-absolutely-does-not-exist-12345.com"},
		{"invalid TLD", "https://invalid.invalidtld999"},
		{"localhost without TLD", "https://localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &httpReq{}
			err := req.SetUrl(tt.url)
			if err == nil {
				t.Fatalf("expected error for URL %q, got nil", tt.url)
			}
		})
	}
}

// TestSetUrl_InvalidScheme tests that non-HTTP(S) schemes are rejected
// Critical: Prevents invalid protocol usage (lines 14-16 in utils.go)
func TestSetUrl_InvalidScheme(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		scheme string
	}{
		{"ftp", "ftp://example.com", "ftp"},
		{"file", "file:///etc/passwd", "file"},
		{"ssh", "ssh://example.com", "ssh"},
		{"ws", "ws://example.com", "ws"},
		{"wss", "wss://example.com", "wss"},
		{"no scheme", "example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &httpReq{}
			err := req.SetUrl(tt.url)
			if err == nil {
				t.Fatalf("expected error for scheme %q (URL: %q), got nil", tt.scheme, tt.url)
			}
		})
	}
}

// TestDo_Non2xxStatusCode tests that non-2xx HTTP status codes return errors
// Critical: Prevents silent failures on API errors (lines 173-175)
func TestDo_Non2xxStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		statusText string
	}{
		{"400 Bad Request", http.StatusBadRequest, "400 Bad Request"},
		{"401 Unauthorized", http.StatusUnauthorized, "401 Unauthorized"},
		{"403 Forbidden", http.StatusForbidden, "403 Forbidden"},
		{"404 Not Found", http.StatusNotFound, "404 Not Found"},
		{"500 Internal Server Error", http.StatusInternalServerError, "500 Internal Server Error"},
		{"503 Service Unavailable", http.StatusServiceUnavailable, "503 Service Unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that returns error status
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("Error response"))
			}))
			defer server.Close()

			// Create HTTP request
			req, err := NewHttpReq("GET", server.URL, nil, http.Header{}, server.Client())
			if err != nil {
				t.Fatalf("unexpected error creating request: %v", err)
			}

			// Execute request
			_, err = req.Do()
			if err == nil {
				t.Fatalf("expected error for status %d, got nil", tt.statusCode)
			}

			// Verify error message mentions HTTP status
			if err.Error() == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

// TestDo_EmptyResponseBody tests that empty response bodies are handled
// Critical: Prevents crashes on JSON parsing of empty responses (lines 182-184)
func TestDo_EmptyResponseBody(t *testing.T) {
	// Create mock server that returns empty body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Don't write anything - empty body
	}))
	defer server.Close()

	// Create HTTP request
	req, err := NewHttpReq("GET", server.URL, nil, http.Header{}, server.Client())
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}

	// Execute request
	_, err = req.Do()
	if err == nil {
		t.Fatal("expected error for empty response body, got nil")
	}

	expectedMsg := "empty response"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}
