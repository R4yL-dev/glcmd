package auth

import (
	"testing"
	"time"
)

// TestIsAuth_NilReceiver tests that IsAuth handles nil receiver gracefully
// Critical: Prevents panic on nil Auth pointer dereference (lines 59-61)
func TestIsAuth_NilReceiver(t *testing.T) {
	var a *Auth = nil

	// Should return false for nil receiver, not panic
	if a.IsAuth() {
		t.Error("expected IsAuth() to return false for nil receiver")
	}
}

// TestIsAuth_ValidAuth tests that IsAuth returns true for valid auth
// Ensures the validation logic works correctly
func TestIsAuth_ValidAuth(t *testing.T) {
	futureTime := time.Now().Add(1 * time.Hour)

	ticket := &authTicket{
		token:    "valid.jwt.token",
		expires:  futureTime,
		duration: 1 * time.Hour,
	}

	auth := &Auth{
		userID:    "user123",
		patientID: "patient456",
		ticket:    ticket,
	}

	if !auth.IsAuth() {
		t.Error("expected IsAuth() to return true for valid auth")
	}
}

// TestIsAuth_InvalidAuth tests various invalid auth states
func TestIsAuth_InvalidAuth(t *testing.T) {
	futureTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name string
		auth *Auth
	}{
		{
			name: "empty userID",
			auth: &Auth{
				userID:    "",
				patientID: "patient123",
				ticket: &authTicket{
					token:    "valid.jwt.token",
					expires:  futureTime,
					duration: 1 * time.Hour,
				},
			},
		},
		{
			name: "empty patientID",
			auth: &Auth{
				userID:    "user123",
				patientID: "",
				ticket: &authTicket{
					token:    "valid.jwt.token",
					expires:  futureTime,
					duration: 1 * time.Hour,
				},
			},
		},
		{
			name: "nil ticket",
			auth: &Auth{
				userID:    "user123",
				patientID: "patient123",
				ticket:    nil,
			},
		},
		{
			name: "expired ticket",
			auth: &Auth{
				userID:    "user123",
				patientID: "patient123",
				ticket: &authTicket{
					token:    "valid.jwt.token",
					expires:  time.Now().Add(-1 * time.Hour),
					duration: 1 * time.Hour,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.auth.IsAuth() {
				t.Errorf("expected IsAuth() to return false for %s", tt.name)
			}
		})
	}
}

// TestSetPatientID_EmptyString tests that empty patientID is rejected
func TestSetPatientID_EmptyString(t *testing.T) {
	auth := &Auth{}
	err := auth.SetPatientID("")

	if err == nil {
		t.Fatal("expected error for empty patientID, got nil")
	}

	expectedMsg := "patientID cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestSetPatientID_ValidString tests that valid patientID is accepted
func TestSetPatientID_ValidString(t *testing.T) {
	auth := &Auth{}
	patientID := "patient123"

	err := auth.SetPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if auth.PatientID() != patientID {
		t.Errorf("expected patientID %q, got %q", patientID, auth.PatientID())
	}
}
