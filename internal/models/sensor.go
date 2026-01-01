package models

import "time"

// SensorConfig represents glucose sensor information from the LibreView API.
// Source: /llu/connections → data[0].sensor
type SensorConfig struct {
	SerialNumber  string    // sn: Serial number of the sensor (e.g., "0JW90U63N0")
	Activation    time.Time // a: Activation timestamp (Unix timestamp converted to time.Time)
	DeviceID      string    // deviceId: ID of the associated device
	SensorType    int       // pt: Sensor type (4 = Libre 3?)
	WarrantyDays  int       // w: Warranty/validity days remaining
	IsActive      bool      // s: Sensor status (false = active?)
	LowJourney    bool      // lj: Unknown field (possibly related to sensor journey)

	// Additional fields for tracking
	DetectedAt    time.Time // When this sensor was first detected by the daemon
}
