package domain

import "time"

// SensorConfig represents glucose sensor information from the LibreView API.
// Source: /llu/connections → data[0].sensor
type SensorConfig struct {
	// Database fields
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`

	SerialNumber  string    `gorm:"type:varchar(50);uniqueIndex:idx_serial;not null" json:"serialNumber"`  // sn: Serial number of the sensor (e.g., "0JW90U63N0")
	Activation    time.Time `gorm:"type:datetime;not null;index:idx_activation" json:"activation"`         // a: Activation timestamp (Unix timestamp converted to time.Time)
	DeviceID      string    `gorm:"type:varchar(100);not null" json:"deviceId"`                            // deviceId: ID of the associated device
	SensorType    int       `gorm:"type:integer;not null" json:"sensorType"`                               // pt: Sensor type (4 = Libre 3?)
	WarrantyDays  int       `gorm:"type:integer" json:"warrantyDays"`                                      // w: Warranty/validity days remaining
	IsActive      bool      `gorm:"type:boolean;not null;default:false;index:idx_active" json:"isActive"` // s: Sensor status (false = active?)
	LowJourney    bool      `gorm:"type:boolean;not null;default:false" json:"lowJourney"`                 // lj: Unknown field (possibly related to sensor journey)

	// Additional fields for tracking
	DetectedAt    time.Time `gorm:"type:datetime;not null" json:"detectedAt"` // When this sensor was first detected by the daemon
}

// TableName specifies the table name for GORM.
func (SensorConfig) TableName() string {
	return "sensor_configs"
}
