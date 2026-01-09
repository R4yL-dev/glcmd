package domain

import "time"

// SensorConfig represents glucose sensor information from the LibreView API.
// Source: /llu/connections → data[0].sensor
type SensorConfig struct {
	// Database fields
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`

	SerialNumber string    `gorm:"type:varchar(50);uniqueIndex:idx_serial;not null" json:"serialNumber"` // sn: Serial number of the sensor
	Activation   time.Time `gorm:"type:datetime;not null;index:idx_activation" json:"activation"`        // a: Activation timestamp
	ExpiresAt    time.Time `gorm:"type:datetime;not null" json:"expiresAt"`                              // Calculated: Activation + DurationDays
	EndedAt      *time.Time `gorm:"type:datetime" json:"endedAt"`                                        // When sensor was replaced (nil = current sensor)
	SensorType   int       `gorm:"type:integer;not null" json:"sensorType"`                              // pt: Sensor type (4 = Libre 3 Plus)
	DurationDays int       `gorm:"type:integer;not null" json:"durationDays"`                            // Expected duration in days (15 for Libre 3 Plus)
	DetectedAt   time.Time `gorm:"type:datetime;not null" json:"detectedAt"`                             // When this sensor was first detected by the daemon
}

// TableName specifies the table name for GORM.
func (SensorConfig) TableName() string {
	return "sensor_configs"
}

// SensorDurationDays returns the expected duration in days for a given sensor type.
func SensorDurationDays(sensorType int) int {
	switch sensorType {
	case 0:
		return 14 // Libre 1
	case 3:
		return 14 // Libre 2
	case 4:
		return 15 // Libre 3 Plus
	default:
		return 14
	}
}

// IsActive returns true if the sensor is currently active (not ended).
func (s *SensorConfig) IsActive() bool {
	return s.EndedAt == nil
}

// RemainingDays returns the number of days remaining until the sensor expires.
// Returns 0 if the sensor has already expired or ended.
func (s *SensorConfig) RemainingDays() float64 {
	if s.EndedAt != nil {
		return 0
	}
	remaining := time.Until(s.ExpiresAt).Hours() / 24
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ElapsedDays returns the number of days since the sensor was activated.
func (s *SensorConfig) ElapsedDays() float64 {
	end := time.Now()
	if s.EndedAt != nil {
		end = *s.EndedAt
	}
	return end.Sub(s.Activation).Hours() / 24
}

// ActualDays returns the actual duration the sensor was used.
// Returns nil if the sensor is still active.
func (s *SensorConfig) ActualDays() *float64 {
	if s.EndedAt == nil {
		return nil
	}
	days := s.EndedAt.Sub(s.Activation).Hours() / 24
	return &days
}
