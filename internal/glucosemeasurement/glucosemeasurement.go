package glucosemeasurement

import (
	"encoding/json"
	"fmt"

	"github.com/R4yL-dev/glcmd/internal/colors"
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

func NewGlucoseMeasurement(data []byte) (*GlucoseMeasurement, error) {
	var gm GlucoseMeasurement

	if err := gm.parseJSON(data); err != nil {
		return nil, err
	}
	return &gm, nil
}

func (gm *GlucoseMeasurement) FactoryTimestamp() string {
	return gm.factoryTimestamp
}

func (gm *GlucoseMeasurement) Timestamp() string {
	return gm.timestamp
}

func (gm *GlucoseMeasurement) ValueInMgPerDl() int {
	return gm.valueInMgPerDl
}

func (gm *GlucoseMeasurement) TrendArrow() rune {
	if sym, ok := trendSymbols[gm.trendArrow]; ok {
		return sym
	}
	return '?'
}

func (gm *GlucoseMeasurement) ToString() string {
	text := fmt.Sprintf("🩸 %s%.1f(mmo/L) %c", colors.Bold, gm.value, gm.TrendArrow())
	return colorize(text, gm.measurementColor)
}

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
			} `json:"glucoseMeasurement"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if len(tmp.Data) == 0 {
		return fmt.Errorf("cannot parse glucose measurement: API returned empty data array")
	}

	gm.factoryTimestamp = tmp.Data[0].GlucoseMeasurement.FactoryTimestamp
	gm.timestamp = tmp.Data[0].GlucoseMeasurement.Timestamp
	gm.valueInMgPerDl = tmp.Data[0].GlucoseMeasurement.ValueInMgPerDl
	gm.trendArrow = tmp.Data[0].GlucoseMeasurement.TrendArrow
	gm.measurementColor = tmp.Data[0].GlucoseMeasurement.MeasurementColor
	gm.glucoseUnits = tmp.Data[0].GlucoseMeasurement.GlucoseUnits
	gm.value = tmp.Data[0].GlucoseMeasurement.Value
	gm.isHigh = tmp.Data[0].GlucoseMeasurement.IsHigh
	gm.isLow = tmp.Data[0].GlucoseMeasurement.IsLow

	return nil
}

func colorize(text string, colorCode int) string {
	color, ok := measurementColors[colorCode]
	if !ok {
		color = colors.Reset
	}
	return fmt.Sprintf("%s%s%s", color, text, colors.Reset)
}
