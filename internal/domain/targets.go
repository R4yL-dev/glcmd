package domain

import "time"

// GlucoseTargets represents global glucose target thresholds.
// Source: /llu/connections → data[0].targetHigh, targetLow, uom
//
// These are used for calculating "Time In Range" statistics.
type GlucoseTargets struct {
	// Database fields
	ID        uint      `gorm:"primaryKey" json:"-"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`

	TargetHigh    int `gorm:"type:integer;not null" json:"targetHigh"`          // targetHigh: High target threshold (in mg/dL)
	TargetLow     int `gorm:"type:integer;not null" json:"targetLow"`           // targetLow: Low target threshold (in mg/dL)
	UnitOfMeasure int `gorm:"type:integer;not null" json:"unitOfMeasure"`       // uom: Unit of measurement (0=mmol/L, 1=mg/dL)
}

// TableName specifies the table name for GORM.
func (GlucoseTargets) TableName() string {
	return "glucose_targets"
}
