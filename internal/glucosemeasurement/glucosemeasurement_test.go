package glucosemeasurement

import (
	"fmt"
	"strings"
	"testing"

	"github.com/R4yL-dev/glcmd/internal/colors"
)

// TestNewGlucoseMeasurement_EmptyDataArray tests that empty data array is properly handled
// Critical: Prevents panic on API returning empty response (line 79)
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

// TestGetTrendArrowSymbol_InvalidCode tests fallback for trend arrow codes outside 1-5
// Critical: Prevents silent data corruption
func TestGetTrendArrowSymbol_InvalidCode(t *testing.T) {
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

			arrow := gm.GetTrendArrowSymbol()
			// TrendArrow will be nil for code 0, and invalid for negative/out of range
			// In all cases, we expect '?'
			if arrow != '?' {
				t.Errorf("expected fallback '?' for invalid trend code %d, got %c", tt.trendCode, arrow)
			}
		})
	}
}

// TestGetTrendArrowSymbol_ValidCodes tests that valid trend codes (1-5) map to correct arrows
// Critical: Ensures correct directional glucose information
func TestGetTrendArrowSymbol_ValidCodes(t *testing.T) {
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

			arrow := gm.GetTrendArrowSymbol()
			if arrow != tt.expected {
				t.Errorf("expected arrow %c for code %d, got %c", tt.expected, tt.code, arrow)
			}

			// Verify TrendArrow pointer is set
			if gm.TrendArrow == nil {
				t.Error("expected TrendArrow to be set, got nil")
			} else if *gm.TrendArrow != tt.code {
				t.Errorf("expected TrendArrow = %d, got %d", tt.code, *gm.TrendArrow)
			}
		})
	}
}

// TestColorize_InvalidColorCode tests fallback for invalid color codes
// Critical: Prevents format errors in output
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

// TestNewGlucoseMeasurement_PublicFields tests that fields are now public and accessible
func TestNewGlucoseMeasurement_PublicFields(t *testing.T) {
	validJSON := createGlucoseJSON(7.5, 3, 1)
	gm, err := NewGlucoseMeasurement([]byte(validJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test public field access
	if gm.Value != 7.5 {
		t.Errorf("expected Value = 7.5, got %f", gm.Value)
	}

	if gm.ValueInMgPerDl != 135 {
		t.Errorf("expected ValueInMgPerDl = 135, got %d", gm.ValueInMgPerDl)
	}

	if gm.MeasurementColor != 1 {
		t.Errorf("expected MeasurementColor = 1, got %d", gm.MeasurementColor)
	}

	if gm.Type != 1 {
		t.Errorf("expected Type = 1, got %d", gm.Type)
	}

	// Verify timestamps are parsed and in UTC
	if gm.Timestamp.IsZero() {
		t.Error("expected Timestamp to be parsed, got zero time")
	}

	if gm.FactoryTimestamp.IsZero() {
		t.Error("expected FactoryTimestamp to be parsed, got zero time")
	}
}

// TestNewGlucoseMeasurement_TrendArrowPointer tests pointer behavior for TrendArrow
func TestNewGlucoseMeasurement_TrendArrowPointer(t *testing.T) {
	t.Run("TrendArrow is nil when code is 0", func(t *testing.T) {
		jsonWith0 := createGlucoseJSON(7.5, 0, 1)
		gm, err := NewGlucoseMeasurement([]byte(jsonWith0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if gm.TrendArrow != nil {
			t.Errorf("expected TrendArrow to be nil for code 0, got %v", *gm.TrendArrow)
		}

		// GetTrendArrowSymbol should return '?' for nil
		symbol := gm.GetTrendArrowSymbol()
		if symbol != '?' {
			t.Errorf("expected '?' for nil TrendArrow, got %c", symbol)
		}
	})

	t.Run("TrendArrow is set when code is valid", func(t *testing.T) {
		jsonWith3 := createGlucoseJSON(7.5, 3, 1)
		gm, err := NewGlucoseMeasurement([]byte(jsonWith3))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if gm.TrendArrow == nil {
			t.Fatal("expected TrendArrow to be set, got nil")
		}

		if *gm.TrendArrow != 3 {
			t.Errorf("expected TrendArrow = 3, got %d", *gm.TrendArrow)
		}
	})
}

// TestNewGlucoseMeasurement_RealAPIExample tests parsing with real API format
func TestNewGlucoseMeasurement_RealAPIExample(t *testing.T) {
	// Real example from exploration/dumps/connections.json
	realJSON := `{
		"data": [{
			"glucoseMeasurement": {
				"FactoryTimestamp": "1/1/2026 1:52:27 PM",
				"Timestamp": "1/1/2026 2:52:27 PM",
				"Value": 5.1,
				"ValueInMgPerDl": 91,
				"TrendArrow": 3,
				"TrendMessage": null,
				"MeasurementColor": 1,
				"GlucoseUnits": 0,
				"isHigh": false,
				"isLow": false,
				"type": 1
			}
		}]
	}`

	gm, err := NewGlucoseMeasurement([]byte(realJSON))
	if err != nil {
		t.Fatalf("failed to parse real API example: %v", err)
	}

	// Verify values
	if gm.Value != 5.1 {
		t.Errorf("expected Value = 5.1, got %f", gm.Value)
	}

	if gm.ValueInMgPerDl != 91 {
		t.Errorf("expected ValueInMgPerDl = 91, got %d", gm.ValueInMgPerDl)
	}

	if gm.TrendArrow == nil || *gm.TrendArrow != 3 {
		t.Errorf("expected TrendArrow = 3, got %v", gm.TrendArrow)
	}

	if gm.Type != 1 {
		t.Errorf("expected Type = 1, got %d", gm.Type)
	}

	// Verify timestamp parsing
	if gm.Timestamp.IsZero() {
		t.Error("expected parsed timestamp, got zero")
	}
}

// Helper function to create valid glucose measurement JSON with LibreView timestamp format
func createGlucoseJSON(value float64, trendArrow, measurementColor int) string {
	return fmt.Sprintf(`{
		"data": [{
			"glucoseMeasurement": {
				"FactoryTimestamp": "1/1/2026 12:00:00 PM",
				"Timestamp": "1/1/2026 1:00:00 PM",
				"ValueInMgPerDl": 135,
				"TrendArrow": %d,
				"MeasurementColor": %d,
				"GlucoseUnits": 1,
				"Value": %.1f,
				"isHigh": false,
				"isLow": false,
				"TrendMessage": null,
				"type": 1
			}
		}]
	}`, trendArrow, measurementColor, value)
}
