package glucosemeasurement

import "time"

// GlucoseMeasurement represents a glucose measurement from the LibreView API.
//
// Fields ending with "mmol" represent values in mmol/L
// Fields ending with "mgdl" represent values in mg/dL
//
// TrendArrow and TrendMessage are pointers because they are absent in historical data
// (only present in current measurements from /llu/connections endpoint)
type GlucoseMeasurement struct {
	// Timestamps
	FactoryTimestamp time.Time // Timestamp from the sensor (factory time)
	Timestamp        time.Time // Real timestamp (phone time), stored in UTC

	// Glucose values
	Value          float64 // Glucose value in mmol/L
	ValueInMgPerDl int     // Glucose value in mg/dL

	// Trend indicators (optional - nil for historical data)
	TrendArrow   *int    // 1-5: direction indicator (1=⬇️⬇️, 2=⬇️, 3=➡️, 4=⬆️, 5=⬆️⬆️)
	TrendMessage *string // Textual trend message (rarely used)

	// Status indicators
	MeasurementColor int  // 1=🟢 normal, 2=🟠 warning, 3=🔴 critical
	GlucoseUnits     int  // 0=mmol/L, 1=mg/dL
	IsHigh           bool // Above high threshold
	IsLow            bool // Below low threshold
	Type             int  // 0=historical, 1=current measurement
}
