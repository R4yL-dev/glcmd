package auth

import (
	"testing"
	"time"
)

// TestSetToken_InvalidJWTFormat tests that invalid JWT formats are rejected
// Critical: Prevents authentication with malformed tokens (lines 54-57)
func TestSetToken_InvalidJWTFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"single part", "singlepart"},
		{"two parts", "part1.part2"},
		{"four parts", "part1.part2.part3.part4"},
		{"no dots", "nodots"},
		{"trailing dot", "part1.part2.part3."},
		{"leading dot", ".part1.part2.part3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := &authTicket{}
			err := ticket.SetToken(tt.token)
			if err == nil {
				t.Fatalf("expected error for token %q, got nil", tt.token)
			}
		})
	}
}

// TestSetExpires_PastTime tests that past expiration times are rejected
// Critical: Prevents accepting expired tokens (lines 72-74)
func TestSetExpires_PastTime(t *testing.T) {
	tests := []struct {
		name   string
		offset time.Duration
	}{
		{"1 second ago", -1 * time.Second},
		{"1 minute ago", -1 * time.Minute},
		{"1 hour ago", -1 * time.Hour},
		{"1 day ago", -24 * time.Hour},
		{"1 year ago", -365 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pastTime := time.Now().Add(tt.offset)
			ticket := &authTicket{}
			err := ticket.SetExpires(pastTime)
			if err == nil {
				t.Fatalf("expected error for expiration %s, got nil", tt.offset)
			}

			expectedMsg := "expiration time cannot be in the past"
			if err.Error() != expectedMsg {
				t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
			}
		})
	}
}

// TestSetExpires_ZeroTime tests that zero time is rejected
// Critical: Prevents uninitialized expiration times (lines 68-70)
func TestSetExpires_ZeroTime(t *testing.T) {
	ticket := &authTicket{}
	err := ticket.SetExpires(time.Time{})
	if err == nil {
		t.Fatal("expected error for zero time, got nil")
	}

	expectedMsg := "expiration time cannot be zero"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestValidate_ExpiredToken tests that expired tokens fail validation
// Critical: Prevents using expired authentication tokens (lines 99-101)
func TestValidate_ExpiredToken(t *testing.T) {
	// Create a ticket with valid JWT format but past expiration
	pastTime := time.Now().Add(-1 * time.Hour)

	ticket := &authTicket{
		token:    "valid.jwt.token",
		expires:  pastTime,
		duration: 1 * time.Hour,
	}

	err := ticket.Validate()
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}

	expectedMsg := "token has expired"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestValidate_ValidToken tests that valid tokens pass validation
// Ensures the validation logic works correctly for non-expired tokens
func TestValidate_ValidToken(t *testing.T) {
	// Create a ticket with valid JWT format and future expiration
	futureTime := time.Now().Add(1 * time.Hour)

	ticket := &authTicket{
		token:    "valid.jwt.token",
		expires:  futureTime,
		duration: 1 * time.Hour,
	}

	err := ticket.Validate()
	if err != nil {
		t.Fatalf("unexpected error for valid token: %v", err)
	}

	// Also test IsValid convenience method
	if !ticket.IsValid() {
		t.Error("expected IsValid() to return true for valid token")
	}
}

// TestSetToken_ValidJWT tests that valid JWT tokens are accepted
// Ensures the validation logic accepts properly formatted tokens
func TestSetToken_ValidJWT(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"typical JWT", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
		{"simple three part", "part1.part2.part3"},
		{"with special chars", "abc_123.def-456.ghi+789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := &authTicket{}
			err := ticket.SetToken(tt.token)
			if err != nil {
				t.Fatalf("unexpected error for valid token %q: %v", tt.token, err)
			}

			if ticket.Token() != tt.token {
				t.Errorf("expected token %q, got %q", tt.token, ticket.Token())
			}
		})
	}
}

// TestSetExpires_FutureTime tests that future expiration times are accepted
// Ensures valid expiration times work correctly
func TestSetExpires_FutureTime(t *testing.T) {
	tests := []struct {
		name   string
		offset time.Duration
	}{
		{"1 second from now", 1 * time.Second},
		{"1 minute from now", 1 * time.Minute},
		{"1 hour from now", 1 * time.Hour},
		{"1 day from now", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			futureTime := time.Now().Add(tt.offset)
			ticket := &authTicket{}
			err := ticket.SetExpires(futureTime)
			if err != nil {
				t.Fatalf("unexpected error for future time %s: %v", tt.offset, err)
			}

			// Verify the time was set (allow small delta for test execution time)
			if ticket.Expires().Before(time.Now()) {
				t.Error("expected expiration to be in the future")
			}
		})
	}
}
