package periodparser

import (
	"testing"
	"time"
)

func TestParse_Today(t *testing.T) {
	start, end, err := Parse("today")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if start == nil || end == nil {
		t.Fatal("expected non-nil start and end for 'today'")
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if !start.Equal(startOfDay) {
		t.Errorf("expected start = %v, got %v", startOfDay, *start)
	}

	// End should be close to now (within a second)
	if end.Sub(now) > time.Second {
		t.Errorf("expected end close to now, got %v", *end)
	}
}

func TestParse_All(t *testing.T) {
	start, end, err := Parse("all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if start != nil || end != nil {
		t.Errorf("expected nil start and end for 'all', got start=%v, end=%v", start, end)
	}
}

func TestParse_Duration(t *testing.T) {
	tests := []struct {
		input    string
		duration time.Duration
	}{
		{"24h", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"2w", 14 * 24 * time.Hour},
		{"1m", 30 * 24 * time.Hour},
		{"3m", 90 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			start, end, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if start == nil || end == nil {
				t.Fatal("expected non-nil start and end")
			}

			actualDuration := end.Sub(*start)
			// Allow 1 second tolerance
			if actualDuration < tt.duration-time.Second || actualDuration > tt.duration+time.Second {
				t.Errorf("expected duration %v, got %v", tt.duration, actualDuration)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"24x",
		"abc",
		"",
		"24",
		"h24",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, _, err := Parse(input)
			if err == nil {
				t.Errorf("expected error for input '%s'", input)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1h", time.Hour},
		{"24h", 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2w", 14 * 24 * time.Hour},
		{"1m", 30 * 24 * time.Hour},
		{"3m", 90 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseDuration(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, d)
			}
		})
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"24x",
		"abc",
		"",
		"today",
		"all",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDuration(input)
			if err == nil {
				t.Errorf("expected error for input '%s'", input)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2025-01-15", time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local)},
		{"2025-01-15 14:30", time.Date(2025, 1, 15, 14, 30, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseDate_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"2025/01/15",
		"01-15-2025",
		"",
		"today",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDate(input)
			if err == nil {
				t.Errorf("expected error for input '%s'", input)
			}
		})
	}
}
