package glucosemeasurement

type GlucoseMeasurement struct {
	factoryTimestamp string
	timestamp        string
	valueInMgPerDl   int
	trendArrow       int
	measurementColor int
	glucoseUnits     int
	value            float64
	isHigh           bool
	isLow            bool
}
