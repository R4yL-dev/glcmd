package models

import "time"

// DeviceInfo represents patient device information and configuration.
// Source: /llu/connections → data[0].patientDevice
type DeviceInfo struct {
	DeviceID         string    // did: Device ID
	DeviceTypeID     int       // dtid: Device type ID (40068 for Libre 3?)
	AppVersion       string    // v: LibreLink app version (e.g., "3.6.5")
	AlarmsEnabled    bool      // alarms: Whether alarms are enabled

	// Threshold configuration (in mg/dL)
	HighLimit        int       // hl: High glucose limit threshold
	LowLimit         int       // ll: Low glucose limit threshold
	FixedLowThreshold int      // fixedLowThreshold: Fixed low threshold value

	// Additional metadata
	LastUpdate       time.Time // u: Last update timestamp (Unix)
	LimitEnabled     bool      // l: Whether limits are enabled
}

// FixedLowAlarmValues represents fixed alarm threshold values in both units.
// Source: /llu/connections → data[0].patientDevice.fixedLowAlarmValues
type FixedLowAlarmValues struct {
	MgPerDl  int     // mgdl: Threshold in mg/dL
	MmolPerL float64 // mmoll: Threshold in mmol/L
}
