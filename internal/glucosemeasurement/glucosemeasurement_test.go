package glucosemeasurement

import (
	"fmt"
	"strings"
	"testing"

	"github.com/R4yL-dev/glcmd/internal/colors"
)

// TestNewGlucoseMeasurement_EmptyDataArray tests that empty data array is properly handled
// Critical: Prevents panic on API returning empty response (line 78-80)
func TestNewGlucoseMeasurement_EmptyDataArray(t *testing.T) {
	emptyData := []byte(`{"data": []}`)

	_, err := NewGlucoseMeasurement(emptyData)
	if err == nil {
		t.Fatal("expected error for empty data array, got nil")
	}

	expectedMsg := "cannot parse glucose measurement: API returned empty data array"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestNewGlucoseMeasurement_InvalidJSON tests that malformed JSON is handled
// Critical: Prevents crashes on corrupted API responses
func TestNewGlucoseMeasurement_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"completely invalid", `{invalid json}`},
		{"truncated", `{"data": [`},
		{"wrong type", `{"data": "not an array"}`},
		{"empty", ``},
		{"null", `null`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGlucoseMeasurement([]byte(tt.json))
			if err == nil {
				t.Fatalf("expected error for %s JSON, got nil", tt.name)
			}
		})
	}
}

// TestTrendArrow_InvalidCode tests fallback for trend arrow codes outside 1-5
// Critical: Prevents silent data corruption (line 46-49)
func TestTrendArrow_InvalidCode(t *testing.T) {
	tests := []struct {
		name      string
		trendCode int
	}{
		{"zero", 0},
		{"negative", -1},
		{"out of range high", 6},
		{"out of range very high", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validJSON := createGlucoseJSON(7.5, tt.trendCode, 1)
			gm, err := NewGlucoseMeasurement([]byte(validJSON))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			arrow := gm.TrendArrow()
			if arrow != '?' {
				t.Errorf("expected fallback '?' for invalid trend code %d, got %c", tt.trendCode, arrow)
			}
		})
	}
}

// TestTrendArrow_ValidCodes tests that valid trend codes (1-5) map to correct arrows
// Critical: Ensures correct directional glucose information
func TestTrendArrow_ValidCodes(t *testing.T) {
	tests := []struct {
		code     int
		expected rune
	}{
		{1, '🡓'}, // Decreasing rapidly
		{2, '🡖'}, // Decreasing
		{3, '🡒'}, // Steady
		{4, '🡕'}, // Increasing
		{5, '🡑'}, // Increasing rapidly
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			validJSON := createGlucoseJSON(7.5, tt.code, 1)
			gm, err := NewGlucoseMeasurement([]byte(validJSON))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			arrow := gm.TrendArrow()
			if arrow != tt.expected {
				t.Errorf("expected arrow %c for code %d, got %c", tt.expected, tt.code, arrow)
			}
		})
	}
}

// TestColorize_InvalidColorCode tests fallback for invalid color codes
// Critical: Prevents format errors in output (line 96-99)
func TestColorize_InvalidColorCode(t *testing.T) {
	tests := []struct {
		name      string
		colorCode int
	}{
		{"zero", 0},
		{"negative", -1},
		{"out of range", 4},
		{"very high", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validJSON := createGlucoseJSON(7.5, 3, tt.colorCode)
			gm, err := NewGlucoseMeasurement([]byte(validJSON))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := gm.ToString()
			// Should use Reset color as fallback
			if !strings.Contains(output, colors.Reset) {
				t.Errorf("expected output to contain Reset color for invalid code %d", tt.colorCode)
			}
			// Should NOT contain the invalid color code
			if strings.Contains(output, colors.Green) || strings.Contains(output, colors.Orange) || strings.Contains(output, colors.Red) {
				t.Errorf("expected no valid color for invalid code %d, but got one", tt.colorCode)
			}
		})
	}
}

// TestColorize_ValidColorCodes tests that valid color codes map to correct colors
// Critical: Ensures correct health indicators (Green=normal, Orange=warning, Red=critical)
func TestColorize_ValidColorCodes(t *testing.T) {
	tests := []struct {
		code          int
		expectedColor string
		description   string
	}{
		{1, colors.Green, "normal"},
		{2, colors.Orange, "warning"},
		{3, colors.Red, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			validJSON := createGlucoseJSON(7.5, 3, tt.code)
			gm, err := NewGlucoseMeasurement([]byte(validJSON))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := gm.ToString()
			if !strings.Contains(output, tt.expectedColor) {
				t.Errorf("expected output to contain color %q for code %d, got %q", tt.expectedColor, tt.code, output)
			}
		})
	}
}

// Helper function to create valid glucose measurement JSON
func createGlucoseJSON(value float64, trendArrow, measurementColor int) string {
	return fmt.Sprintf(`{
		"data": [{
			"glucoseMeasurement": {
				"FactoryTimestamp": "2024-01-01T12:00:00",
				"Timestamp": "2024-01-01T12:00:00",
				"ValueInMgPerDl": 135,
				"TrendArrow": %d,
				"MeasurementColor": %d,
				"GlucoseUnits": 1,
				"Value": %.1f,
				"isHigh": false,
				"isLow": false
			}
		}]
	}`, trendArrow, measurementColor, value)
}
