package domain

import "time"

// SensorStatus represents the operational state of the sensor.
type SensorStatus string

const (
	// SensorStatusRunning indicates the sensor is active and within its lifetime.
	SensorStatusRunning SensorStatus = "running"
	// SensorStatusStopped indicates the sensor is no longer active (replaced or expired).
	SensorStatusStopped SensorStatus = "stopped"
	// SensorStatusUnresponsive indicates the sensor is not sending data (no measurement for > 20 min).
	SensorStatusUnresponsive SensorStatus = "unresponsive"
)

// UnresponsiveThreshold is the duration after which a sensor is considered unresponsive
// if no measurements have been received.
const UnresponsiveThreshold = 20 * time.Minute

// SensorConfig represents glucose sensor information from the LibreView API.
// Source: /llu/connections → data[0].sensor
type SensorConfig struct {
	// Database fields
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`

	SerialNumber      string     `gorm:"type:varchar(50);uniqueIndex:idx_serial;not null" json:"serialNumber"` // sn: Serial number of the sensor
	Activation        time.Time  `gorm:"type:datetime;not null;index:idx_activation" json:"activation"`        // a: Activation timestamp
	ExpiresAt         time.Time  `gorm:"type:datetime;not null" json:"expiresAt"`                              // Calculated: Activation + DurationDays
	EndedAt           *time.Time `gorm:"type:datetime" json:"endedAt"`                                         // When sensor was replaced (nil = current sensor)
	LastMeasurementAt *time.Time `gorm:"type:datetime" json:"lastMeasurementAt"`                               // Timestamp of the last received measurement
	SensorType        int        `gorm:"type:integer;not null" json:"sensorType"`                              // pt: Sensor type (4 = Libre 3 Plus)
	DurationDays      int        `gorm:"type:integer;not null" json:"durationDays"`                            // Expected duration in days (15 for Libre 3 Plus)
	DetectedAt        time.Time  `gorm:"type:datetime;not null" json:"detectedAt"`                             // When this sensor was first detected by the daemon
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
// For stopped sensors, this is bounded by EndedAt or ExpiresAt.
func (s *SensorConfig) ElapsedDays() float64 {
	end := time.Now()
	if s.EndedAt != nil {
		end = *s.EndedAt
	} else if end.After(s.ExpiresAt) {
		end = s.ExpiresAt
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

// Status returns the current operational status of the sensor.
//   - "stopped": Sensor has been replaced (EndedAt set) or expired (now > ExpiresAt)
//   - "unresponsive": Sensor is active but not sending data (no measurement for > 20 min)
//   - "running": Sensor is active and within its lifetime
func (s *SensorConfig) Status() SensorStatus {
	if s.EndedAt != nil {
		return SensorStatusStopped
	}
	if time.Now().After(s.ExpiresAt) {
		return SensorStatusStopped
	}
	if s.LastMeasurementAt != nil && time.Since(*s.LastMeasurementAt) > UnresponsiveThreshold {
		return SensorStatusUnresponsive
	}
	return SensorStatusRunning
}
