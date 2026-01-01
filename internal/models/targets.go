package models

// GlucoseTargets represents global glucose target thresholds.
// Source: /llu/connections → data[0].targetHigh, targetLow, uom
//
// These are used for calculating "Time In Range" statistics.
type GlucoseTargets struct {
	TargetHigh int // targetHigh: High target threshold (in mg/dL)
	TargetLow  int // targetLow: Low target threshold (in mg/dL)
	UnitOfMeasure int // uom: Unit of measurement (0=mmol/L, 1=mg/dL)
}
