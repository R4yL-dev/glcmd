package domain

import "time"

// DeviceInfo represents patient device information and configuration.
// Source: /llu/connections → data[0].patientDevice
type DeviceInfo struct {
	// Database fields
	ID        uint      `gorm:"primaryKey" json:"-"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`

	DeviceID         string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"deviceId"`       // did: Device ID
	DeviceTypeID     int       `gorm:"type:integer;not null" json:"deviceTypeId"`                    // dtid: Device type ID (40068 for Libre 3?)
	AppVersion       string    `gorm:"type:varchar(50)" json:"appVersion"`                           // v: LibreLink app version (e.g., "3.6.5")
	AlarmsEnabled    bool      `gorm:"type:boolean;not null;default:false" json:"alarmsEnabled"`     // alarms: Whether alarms are enabled

	// Threshold configuration (in mg/dL)
	HighLimit        int       `gorm:"type:integer" json:"highLimit"`                                // hl: High glucose limit threshold
	LowLimit         int       `gorm:"type:integer" json:"lowLimit"`                                 // ll: Low glucose limit threshold
	FixedLowThreshold int      `gorm:"type:integer" json:"fixedLowThreshold"`                        // fixedLowThreshold: Fixed low threshold value

	// Additional metadata
	LastUpdate       time.Time `gorm:"type:datetime" json:"lastUpdate"`                              // u: Last update timestamp (Unix)
	LimitEnabled     bool      `gorm:"type:boolean;not null;default:false" json:"limitEnabled"`      // l: Whether limits are enabled
}

// TableName specifies the table name for GORM.
func (DeviceInfo) TableName() string {
	return "device_info"
}

// FixedLowAlarmValues represents fixed alarm threshold values in both units.
// Source: /llu/connections → data[0].patientDevice.fixedLowAlarmValues
// Note: This is not persisted to the database, it's a transient value from the API
type FixedLowAlarmValues struct {
	MgPerDl  int     // mgdl: Threshold in mg/dL
	MmolPerL float64 // mmoll: Threshold in mmol/L
}
