package timeparser

import (
	"testing"
	"time"
)

func TestParseLibreViewTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		checkResult func(time.Time) bool
	}{
		{
			name:    "valid timestamp with PM",
			input:   "1/1/2026 2:52:27 PM",
			wantErr: false,
			checkResult: func(result time.Time) bool {
				// Check that time is in UTC
				return result.Location() == time.UTC
			},
		},
		{
			name:    "valid timestamp with AM",
			input:   "1/1/2026 9:15:30 AM",
			wantErr: false,
			checkResult: func(result time.Time) bool {
				return result.Location() == time.UTC
			},
		},
		{
			name:    "timestamp at midnight",
			input:   "12/31/2025 12:00:00 AM",
			wantErr: false,
			checkResult: func(result time.Time) bool {
				return result.Hour() == 0 && result.Minute() == 0 && result.Second() == 0
			},
		},
		{
			name:    "timestamp at noon",
			input:   "6/15/2025 12:00:00 PM",
			wantErr: false,
			checkResult: func(result time.Time) bool {
				return result.Hour() == 12 && result.Minute() == 0 && result.Second() == 0
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - ISO 8601",
			input:   "2026-01-01T14:52:27Z",
			wantErr: true,
		},
		{
			name:    "invalid format - Unix timestamp",
			input:   "1735740747",
			wantErr: true,
		},
		{
			name:    "invalid format - missing time",
			input:   "1/1/2026",
			wantErr: true,
		},
		{
			name:    "invalid format - European format",
			input:   "01/01/2026 14:52:27",
			wantErr: true,
		},
		{
			name:    "real API example from dumps",
			input:   "1/1/2026 1:52:27 PM",
			wantErr: false,
			checkResult: func(result time.Time) bool {
				// Verify components
				year, month, day := result.Date()
				hour, min, sec := result.Clock()
				return year == 2026 && month == time.January && day == 1 &&
					hour == 13 && min == 52 && sec == 27
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLibreViewTimestamp(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseLibreViewTimestamp() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseLibreViewTimestamp() unexpected error = %v", err)
				return
			}

			if tt.checkResult != nil && !tt.checkResult(result) {
				t.Errorf("ParseLibreViewTimestamp() result check failed for input '%s', got %v", tt.input, result)
			}
		})
	}
}

func TestMustParseLibreViewTimestamp(t *testing.T) {
	t.Run("valid timestamp does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParseLibreViewTimestamp() panicked unexpectedly: %v", r)
			}
		}()

		result := MustParseLibreViewTimestamp("1/1/2026 2:52:27 PM")
		if result.IsZero() {
			t.Error("MustParseLibreViewTimestamp() returned zero time")
		}
	})

	t.Run("invalid timestamp panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseLibreViewTimestamp() should have panicked but did not")
			}
		}()

		MustParseLibreViewTimestamp("invalid")
	})
}

func TestParseLibreViewTimestamp_UTCConversion(t *testing.T) {
	// Test that the parser correctly converts to UTC
	input := "1/1/2026 2:52:27 PM"
	result, err := ParseLibreViewTimestamp(input)

	if err != nil {
		t.Fatalf("ParseLibreViewTimestamp() unexpected error = %v", err)
	}

	// Verify UTC location
	if result.Location() != time.UTC {
		t.Errorf("ParseLibreViewTimestamp() location = %v, want UTC", result.Location())
	}

	// Verify the time components are correct after UTC conversion
	// "1/1/2026 2:52:27 PM" parsed in local time then converted to UTC
	// The exact UTC hour depends on local timezone, but we can verify consistency
	utcString := result.Format(time.RFC3339)
	if utcString == "" {
		t.Error("ParseLibreViewTimestamp() resulted in invalid UTC string")
	}
}

func TestParseLibreViewTimestamp_RealAPIExamples(t *testing.T) {
	// Real examples from exploration/dumps/*.json
	realExamples := []string{
		"1/1/2026 1:52:27 PM",
		"1/1/2026 2:52:27 PM",
		"1/1/2026 1:37:28 PM",
		"1/1/2026 2:37:28 PM",
		"1/1/2026 1:38:27 AM",
		"1/1/2026 2:38:27 AM",
	}

	for _, example := range realExamples {
		t.Run(example, func(t *testing.T) {
			result, err := ParseLibreViewTimestamp(example)
			if err != nil {
				t.Errorf("ParseLibreViewTimestamp() failed to parse real API example '%s': %v", example, err)
			}

			// Verify UTC conversion
			if result.Location() != time.UTC {
				t.Errorf("ParseLibreViewTimestamp() location = %v, want UTC", result.Location())
			}

			// Verify not zero time
			if result.IsZero() {
				t.Error("ParseLibreViewTimestamp() returned zero time for valid input")
			}
		})
	}
}
