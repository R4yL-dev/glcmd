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

// SensorsData contains active sensor and all sensors
type SensorsData struct {
	Active  *domain.SensorConfig   `json:"active"`
	Sensors []*domain.SensorConfig `json:"sensors"`
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
