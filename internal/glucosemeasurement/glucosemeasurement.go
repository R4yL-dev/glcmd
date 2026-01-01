package glucosemeasurement

import (
	"encoding/json"
	"fmt"

	"github.com/R4yL-dev/glcmd/internal/colors"
	"github.com/R4yL-dev/glcmd/internal/utils/timeparser"
)

var trendSymbols = map[int]rune{
	1: '🡓',
	2: '🡖',
	3: '🡒',
	4: '🡕',
	5: '🡑',
}

var measurementColors = map[int]string{
	1: colors.Green,
	2: colors.Orange,
	3: colors.Red,
}

// NewGlucoseMeasurement creates a new GlucoseMeasurement by parsing JSON data
// from the LibreView API /llu/connections endpoint.
func NewGlucoseMeasurement(data []byte) (*GlucoseMeasurement, error) {
	var gm GlucoseMeasurement

	if err := gm.parseJSON(data); err != nil {
		return nil, err
	}
	return &gm, nil
}

// GetTrendArrowSymbol returns the unicode symbol for the trend arrow.
// Returns '?' if trend arrow is nil or invalid.
func (gm *GlucoseMeasurement) GetTrendArrowSymbol() rune {
	if gm.TrendArrow == nil {
		return '?'
	}
	if sym, ok := trendSymbols[*gm.TrendArrow]; ok {
		return sym
	}
	return '?'
}

// ToString returns a formatted string representation of the glucose measurement
// with color coding based on measurement color.
func (gm *GlucoseMeasurement) ToString() string {
	text := fmt.Sprintf("🩸 %s%.1f(mmo/L) %c", colors.Bold, gm.Value, gm.GetTrendArrowSymbol())
	return colorize(text, gm.MeasurementColor)
}

// parseJSON parses JSON data from /llu/connections endpoint
func (gm *GlucoseMeasurement) parseJSON(data []byte) error {
	var tmp struct {
		Data []struct {
			GlucoseMeasurement struct {
				FactoryTimestamp string  `json:"FactoryTimestamp"`
				Timestamp        string  `json:"Timestamp"`
				ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
				TrendArrow       int     `json:"TrendArrow"`
				MeasurementColor int     `json:"MeasurementColor"`
				GlucoseUnits     int     `json:"GlucoseUnits"`
				Value            float64 `json:"Value"`
				IsHigh           bool    `json:"isHigh"`
				IsLow            bool    `json:"isLow"`
				TrendMessage     *string `json:"TrendMessage"`
				Type             int     `json:"type"`
			} `json:"glucoseMeasurement"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if len(tmp.Data) == 0 {
		return fmt.Errorf("cannot parse glucose measurement: API returned empty data array")
	}

	measurement := tmp.Data[0].GlucoseMeasurement

	// Parse timestamps using timeparser (converts to UTC)
	factoryTs, err := timeparser.ParseLibreViewTimestamp(measurement.FactoryTimestamp)
	if err != nil {
		return fmt.Errorf("failed to parse FactoryTimestamp: %w", err)
	}

	ts, err := timeparser.ParseLibreViewTimestamp(measurement.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse Timestamp: %w", err)
	}

	gm.FactoryTimestamp = factoryTs
	gm.Timestamp = ts
	gm.ValueInMgPerDl = measurement.ValueInMgPerDl
	gm.MeasurementColor = measurement.MeasurementColor
	gm.GlucoseUnits = measurement.GlucoseUnits
	gm.Value = measurement.Value
	gm.IsHigh = measurement.IsHigh
	gm.IsLow = measurement.IsLow
	gm.Type = measurement.Type

	// TrendArrow: use pointer to distinguish between 0 (invalid) and nil (absent)
	// Only set if non-zero (0 is not a valid trend arrow value)
	if measurement.TrendArrow != 0 {
		gm.TrendArrow = &measurement.TrendArrow
	}

	// TrendMessage: already a pointer from JSON
	gm.TrendMessage = measurement.TrendMessage

	return nil
}

func colorize(text string, colorCode int) string {
	color, ok := measurementColors[colorCode]
	if !ok {
		color = colors.Reset
	}
	return fmt.Sprintf("%s%s%s", color, text, colors.Reset)
}
