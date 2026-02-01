package api

import (
	"encoding/json"
	"net/http"

	"github.com/R4yL-dev/glcmd/internal/daemon"
	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/service"
)

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	Total   int64 `json:"total"`
	HasMore bool  `json:"hasMore"`
}

// MeasurementListResponse represents a paginated list of measurements
type MeasurementListResponse struct {
	Data       []*domain.GlucoseMeasurement `json:"data"`
	Pagination PaginationMetadata           `json:"pagination"`
}

// MeasurementResponse represents a single measurement response
type MeasurementResponse struct {
	Data *domain.GlucoseMeasurement `json:"data"`
}

// StatisticsResponse represents statistics response
type StatisticsResponse struct {
	Data StatisticsData `json:"data"`
}

// StatisticsData contains the statistics information
type StatisticsData struct {
	Period      PeriodInfo                `json:"period"`
	Statistics  service.MeasurementStats  `json:"statistics"`
	TimeInRange *TimeInRangeData          `json:"timeInRange,omitempty"`
	Distribution DistributionData         `json:"distribution"`
}

// PeriodInfo contains the time period for statistics
type PeriodInfo struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TimeInRangeData contains time in range metrics
type TimeInRangeData struct {
	TargetLowMgDl  int     `json:"targetLowMgDl"`
	TargetHighMgDl int     `json:"targetHighMgDl"`
	TargetLow      float64 `json:"targetLow"`
	TargetHigh     float64 `json:"targetHigh"`
	InRange        float64 `json:"inRange"`
	BelowRange     float64 `json:"belowRange"`
	AboveRange     float64 `json:"aboveRange"`
}

// DistributionData contains distribution by color
type DistributionData struct {
	Low    int `json:"low"`
	Normal int `json:"normal"`
	High   int `json:"high"`
}

// SensorsResponse represents sensors response
type SensorsResponse struct {
	Data SensorsData `json:"data"`
}

// SensorsData contains current sensor and history
type SensorsData struct {
	Current *SensorResponse   `json:"current"`
	History []*SensorResponse `json:"history"`
}

// SensorResponse represents a sensor with calculated fields
type SensorResponse struct {
	SerialNumber      string   `json:"serialNumber"`
	Activation        string   `json:"activation"`
	ExpiresAt         string   `json:"expiresAt"`
	EndedAt           *string  `json:"endedAt,omitempty"`
	LastMeasurementAt *string  `json:"lastMeasurementAt,omitempty"`
	SensorType        int      `json:"sensorType"`
	DurationDays      int      `json:"durationDays"`
	DaysRemaining     *float64 `json:"daysRemaining,omitempty"`
	DaysElapsed       float64  `json:"daysElapsed"`
	ActualDays        *float64 `json:"actualDays,omitempty"`
	DaysPastExpiry    *float64 `json:"daysPastExpiry,omitempty"`
	IsActive          bool     `json:"isActive"`
	Status            string   `json:"status"`
	IsExpired         bool     `json:"isExpired"`
	IsUnresponsive    bool     `json:"isUnresponsive"`
}

// SensorListResponse represents a paginated list of sensors
type SensorListResponse struct {
	Data       []*SensorResponse  `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// SensorStatisticsResponse represents sensor statistics response
type SensorStatisticsResponse struct {
	Data SensorStatisticsData `json:"data"`
}

// SensorStatisticsData contains sensor lifecycle statistics
type SensorStatisticsData struct {
	Statistics service.SensorStats `json:"statistics"`
	Current    *SensorResponse     `json:"current,omitempty"`
}

// NewSensorResponse creates a SensorResponse from a domain.SensorConfig
func NewSensorResponse(s *domain.SensorConfig) *SensorResponse {
	resp := &SensorResponse{
		SerialNumber:   s.SerialNumber,
		Activation:     s.Activation.Format("2006-01-02T15:04:05Z"),
		ExpiresAt:      s.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		SensorType:     s.SensorType,
		DurationDays:   s.DurationDays,
		DaysElapsed:    s.ElapsedDays(),
		IsActive:       s.IsActive(),
		Status:         string(s.Status()),
		IsExpired:      s.IsExpired(),
		IsUnresponsive: s.IsUnresponsive(),
	}

	if s.EndedAt != nil {
		endedAtStr := s.EndedAt.Format("2006-01-02T15:04:05Z")
		resp.EndedAt = &endedAtStr
		resp.ActualDays = s.ActualDays()
	} else {
		remaining := s.RemainingDays()
		resp.DaysRemaining = &remaining
	}

	if s.LastMeasurementAt != nil {
		lastMeasurementAtStr := s.LastMeasurementAt.Format("2006-01-02T15:04:05Z")
		resp.LastMeasurementAt = &lastMeasurementAtStr
	}

	// Add days past expiry if sensor is expired
	if s.IsExpired() {
		daysPast := s.DaysPastExpiry()
		resp.DaysPastExpiry = &daysPast
	}

	return resp
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// HealthResponse represents health endpoint response
type HealthResponse struct {
	Data daemon.HealthStatus `json:"data"`
}

// MetricsResponse represents metrics endpoint response
type MetricsResponse struct {
	Data MetricsData `json:"data"`
}

// MetricsData contains runtime and system metrics
type MetricsData struct {
	Uptime     string      `json:"uptime"`
	Goroutines int         `json:"goroutines"`
	Memory     MemoryStats `json:"memory"`
	Runtime    RuntimeInfo `json:"runtime"`
	Process    ProcessInfo `json:"process"`
}

// MemoryStats contains memory statistics
type MemoryStats struct {
	AllocMB      uint64 `json:"allocMB"`
	TotalAllocMB uint64 `json:"totalAllocMB"`
	SysMB        uint64 `json:"sysMB"`
	NumGC        uint32 `json:"numGC"`
}

// RuntimeInfo contains Go runtime information
type RuntimeInfo struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// ProcessInfo contains process information
type ProcessInfo struct {
	PID int `json:"pid"`
}

// newPaginationMetadata creates pagination metadata
func newPaginationMetadata(limit, offset int, total int64) PaginationMetadata {
	hasMore := int64(offset+limit) < total
	return PaginationMetadata{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: hasMore,
	}
}
