package headers

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestBuildAuthHeader_EmptyInputs tests behavior with empty token and userID
// Critical: Ensures headers are built even with empty inputs (lines 30-36)
func TestBuildAuthHeader_EmptyInputs(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		userID string
	}{
		{"both empty", "", ""},
		{"empty token", "", "user123"},
		{"empty userID", "token123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeaders()
			h.BuildAuthHeader(tt.token, tt.userID)

			authHeader := h.AuthHeader()

			// Check Authorization header
			authValue := authHeader.Get("Authorization")
			expectedAuth := "Bearer " + tt.token
			if authValue != expectedAuth {
				t.Errorf("expected Authorization %q, got %q", expectedAuth, authValue)
			}

			// Check account-id header (SHA256 hash of userID)
			hasher := sha256.New()
			hasher.Write([]byte(tt.userID))
			hasherByte := hasher.Sum(nil)
			expectedHash := hex.EncodeToString(hasherByte)

			accountID := authHeader.Get("account-id")
			if accountID != expectedHash {
				t.Errorf("expected account-id %q, got %q", expectedHash, accountID)
			}
		})
	}
}

// TestBuildAuthHeader_ValidInputs tests behavior with valid inputs
// Ensures the SHA256 hashing and Bearer token formatting work correctly
func TestBuildAuthHeader_ValidInputs(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		userID string
	}{
		{"simple values", "abc123token", "user456"},
		{"long token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", "12345"},
		{"special chars in userID", "token", "user-id_123@example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeaders()
			h.BuildAuthHeader(tt.token, tt.userID)

			authHeader := h.AuthHeader()

			// Verify Authorization header
			authValue := authHeader.Get("Authorization")
			expectedAuth := "Bearer " + tt.token
			if authValue != expectedAuth {
				t.Errorf("expected Authorization %q, got %q", expectedAuth, authValue)
			}

			// Verify account-id is a valid hex SHA256 hash
			accountID := authHeader.Get("account-id")
			if len(accountID) != 64 { // SHA256 hex is always 64 characters
				t.Errorf("expected account-id length 64, got %d", len(accountID))
			}

			// Verify the hash matches expected value
			hasher := sha256.New()
			hasher.Write([]byte(tt.userID))
			hasherByte := hasher.Sum(nil)
			expectedHash := hex.EncodeToString(hasherByte)

			if accountID != expectedHash {
				t.Errorf("expected account-id %q, got %q", expectedHash, accountID)
			}
		})
	}
}

// TestBuildAuthHeader_HeaderCloning tests that default headers are cloned
// Ensures modifications to authHeader don't affect defaultHeader
func TestBuildAuthHeader_HeaderCloning(t *testing.T) {
	h := NewHeaders()

	// Store original default header values
	originalUserAgent := h.DefaultHeader().Get("User-Agent")

	// Build auth header
	h.BuildAuthHeader("token123", "user123")

	// Verify auth header has the default headers
	authUserAgent := h.AuthHeader().Get("User-Agent")
	if authUserAgent != originalUserAgent {
		t.Errorf("expected auth header to inherit User-Agent %q, got %q", originalUserAgent, authUserAgent)
	}

	// Modify auth header
	h.AuthHeader().Set("User-Agent", "Modified")

	// Verify default header is unchanged (was properly cloned)
	currentDefault := h.DefaultHeader().Get("User-Agent")
	if currentDefault != originalUserAgent {
		t.Error("default header was modified when it should have been cloned")
	}
}

// TestNewHeaders tests the constructor
func TestNewHeaders(t *testing.T) {
	h := NewHeaders()

	if h == nil {
		t.Fatal("expected non-nil Headers")
	}

	// Verify default header is initialized
	defaultHeader := h.DefaultHeader()
	if defaultHeader == nil {
		t.Fatal("expected non-nil default header")
	}

	// Verify it has expected default headers from config
	if defaultHeader.Get("User-Agent") == "" {
		t.Error("expected User-Agent to be set in default header")
	}

	if defaultHeader.Get("Content-Type") == "" {
		t.Error("expected Content-Type to be set in default header")
	}
}
