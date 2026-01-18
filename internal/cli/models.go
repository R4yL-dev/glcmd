package cli

import "time"

// MeasurementListResponse represents the API response for measurements list
type MeasurementListResponse struct {
	Data       []GlucoseReading `json:"data"`
	Pagination PaginationInfo   `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// StatisticsResponse represents the API response for glucose statistics
type StatisticsResponse struct {
	Data StatisticsData `json:"data"`
}

// StatisticsData contains all statistics data
type StatisticsData struct {
	Period       StatsPeriod       `json:"period"`
	Statistics   StatsDetails      `json:"statistics"`
	Distribution StatsDistribution `json:"distribution"`
	TimeInRange  *StatsTimeInRange `json:"timeInRange,omitempty"`
}

// StatsPeriod represents the time period for statistics
type StatsPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// StatsDetails contains detailed statistics
type StatsDetails struct {
	Count          int     `json:"count"`
	Average        float64 `json:"average"`
	AverageMgDl    float64 `json:"averageMgDl"`
	Min            float64 `json:"min"`
	MinMgDl        int     `json:"minMgDl"`
	Max            float64 `json:"max"`
	MaxMgDl        int     `json:"maxMgDl"`
	StdDev         float64 `json:"stdDev"`
	LowCount       int     `json:"lowCount"`
	NormalCount    int     `json:"normalCount"`
	HighCount      int     `json:"highCount"`
	TimeInRange    float64 `json:"timeInRange"`
	TimeBelowRange float64 `json:"timeBelowRange"`
	TimeAboveRange float64 `json:"timeAboveRange"`
}

// StatsDistribution contains distribution statistics
type StatsDistribution struct {
	Low    int `json:"low"`
	Normal int `json:"normal"`
	High   int `json:"high"`
}

// StatsTimeInRange contains time in range data
type StatsTimeInRange struct {
	TargetLowMgDl  int     `json:"targetLowMgDl"`
	TargetHighMgDl int     `json:"targetHighMgDl"`
	InRange        float64 `json:"inRange"`
	BelowRange     float64 `json:"belowRange"`
	AboveRange     float64 `json:"aboveRange"`
}

// MeasurementParams contains parameters for fetching measurements
type MeasurementParams struct {
	Start *time.Time
	End   *time.Time
	Limit int
}
