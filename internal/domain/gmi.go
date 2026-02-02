package domain

// CalculateGMI computes the Glucose Management Indicator from average glucose in mg/dL.
// Formula: GMI(%) = 3.31 + 0.02392 × [mean glucose in mg/dL]
// Returns nil if averageMgDl <= 0.
func CalculateGMI(averageMgDl float64) *float64 {
	if averageMgDl <= 0 {
		return nil
	}
	gmi := 3.31 + 0.02392*averageMgDl
	return &gmi
}
