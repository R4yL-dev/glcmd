package domain

import "time"

// Glucose type constants
const (
	GlucoseTypeHistorical = 0 // Historical measurement from /graph endpoint
	GlucoseTypeCurrent    = 1 // Current measurement from /connections endpoint
)

// GlucoseColor constants
const (
	GlucoseColorNormal   = 1 // 🟢 Normal glucose levels
	GlucoseColorWarning  = 2 // 🟠 Warning - outside target range
	GlucoseColorCritical = 3 // 🔴 Critical - dangerous levels
)

// TrendArrow constants
const (
	TrendArrowFallingRapidly = 1 // ⬇️⬇️ Falling rapidly
	TrendArrowFalling        = 2 // ⬇️ Falling
	TrendArrowStable         = 3 // ➡️ Stable
	TrendArrowRising         = 4 // ⬆️ Rising
	TrendArrowRisingRapidly  = 5 // ⬆️⬆️ Rising rapidly
)

// GlucoseUnits constants
const (
	GlucoseUnitsMmolL = 0 // mmol/L (millimoles per liter)
	GlucoseUnitsMgDl  = 1 // mg/dL (milligrams per deciliter)
)

// GlucoseMeasurement represents a glucose measurement from the LibreView API.
//
// Fields ending with "mmol" represent values in mmol/L
// Fields ending with "mgdl" represent values in mg/dL
//
// TrendArrow and TrendMessage are pointers because they are absent in historical data
// (only present in current measurements from /llu/connections endpoint)
type GlucoseMeasurement struct {
	// Database fields
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`

	// Timestamps
	FactoryTimestamp time.Time `gorm:"type:datetime;not null;uniqueIndex:idx_unique_factory_ts" json:"factoryTimestamp"` // Timestamp from the sensor (factory time), used for deduplication
	Timestamp        time.Time `gorm:"type:datetime;not null;index:idx_timestamp" json:"timestamp"` // Real timestamp (phone time), stored in UTC

	// Glucose values
	Value          float64 `gorm:"type:decimal(10,2);not null" json:"value"`          // Glucose value in mmol/L
	ValueInMgPerDl int     `gorm:"type:integer;not null" json:"valueInMgPerDl"`       // Glucose value in mg/dL

	// Trend indicators (optional - nil for historical data)
	TrendArrow   *int    `gorm:"type:integer" json:"trendArrow,omitempty"`     // 1-5: direction indicator (1=⬇️⬇️, 2=⬇️, 3=➡️, 4=⬆️, 5=⬆️⬆️)
	TrendMessage *string `gorm:"type:text" json:"trendMessage,omitempty"`      // Textual trend message (rarely used)

	// Status indicators
	GlucoseColor int  `gorm:"type:integer;not null;index:idx_color;column:measurement_color" json:"measurementColor"` // 1=🟢 normal, 2=🟠 warning, 3=🔴 critical
	GlucoseUnits     int  `gorm:"type:integer;not null" json:"glucoseUnits"`                     // 0=mmol/L, 1=mg/dL
	IsHigh           bool `gorm:"type:boolean;not null;default:false" json:"isHigh"`             // Above high threshold
	IsLow            bool `gorm:"type:boolean;not null;default:false" json:"isLow"`              // Below low threshold
	Type             int  `gorm:"type:integer;not null;index:idx_type" json:"type"`              // 0=historical, 1=current measurement
}

// TableName specifies the table name for GORM.
func (GlucoseMeasurement) TableName() string {
	return "glucose_measurements"
}
