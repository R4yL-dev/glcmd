package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// handleGetLatestGlucose handles GET /glucose/latest
func (s *Server) handleGetLatestGlucose(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	measurement, err := s.glucoseService.GetLatestMeasurement(ctx)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "No measurements found")
			return
		}
		handleError(w, err, s.logger)
		return
	}

	response := MeasurementResponse{
		Data: measurement,
	}

	if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// handleGetGlucose handles GET /glucose
func (s *Server) handleGetGlucose(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit, offset, err := parsePaginationParams(r)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	// Parse filter parameters
	filters, err := parseGlucoseFilters(r)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Get measurements and total count
	measurements, total, err := s.glucoseService.GetMeasurementsWithFilters(ctx, filters, limit, offset)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	// Build response with pagination
	response := MeasurementListResponse{
		Data:       measurements,
		Pagination: newPaginationMetadata(limit, offset, total),
	}

	if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// handleGetGlucoseStatistics handles GET /glucose/stats
func (s *Server) handleGetGlucoseStatistics(w http.ResponseWriter, r *http.Request) {
	// Parse and validate parameters (nil = all time)
	start, end, err := parseStatisticsParams(r)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	// Use longer timeout for potentially large queries
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get glucose targets for Time in Range calculation
	targets, err := s.configService.GetGlucoseTargets(ctx)
	if err != nil && !errors.Is(err, persistence.ErrNotFound) {
		handleError(w, err, s.logger)
		return
	}

	// Calculate statistics
	stats, err := s.glucoseService.GetStatistics(ctx, start, end, targets)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	// Build response with period info
	var periodInfo PeriodInfo
	if start != nil && end != nil {
		periodInfo = PeriodInfo{
			Start: start.Format(time.RFC3339),
			End:   end.Format(time.RFC3339),
		}
	} else {
		// All time - use actual data bounds from database
		if stats.FirstTimestamp != nil {
			periodInfo.Start = stats.FirstTimestamp.Format(time.RFC3339)
		}
		if stats.LastTimestamp != nil {
			periodInfo.End = stats.LastTimestamp.Format(time.RFC3339)
		}
	}

	data := StatisticsData{
		Period:     periodInfo,
		Statistics: *stats,
		Distribution: DistributionData{
			Low:    stats.LowCount,
			Normal: stats.NormalCount,
			High:   stats.HighCount,
		},
	}

	// Add Time in Range data if targets were available
	if targets != nil {
		data.TimeInRange = &TimeInRangeData{
			TargetLowMgDl:  targets.TargetLow,
			TargetHighMgDl: targets.TargetHigh,
			TargetLow:      float64(targets.TargetLow) / 18.0182,
			TargetHigh:     float64(targets.TargetHigh) / 18.0182,
			InRange:        stats.TimeInRange,
			BelowRange:     stats.TimeBelowRange,
			AboveRange:     stats.TimeAboveRange,
		}
	}

	response := StatisticsResponse{
		Data: data,
	}

	if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// handleGetSensor handles GET /sensor
// Returns a paginated list of sensors with optional filters
func (s *Server) handleGetSensor(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePaginationParams(r)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	filters, err := parseSensorFilters(r)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sensors, total, err := s.sensorService.GetSensorsWithFilters(ctx, filters, limit, offset)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	data := make([]*SensorResponse, 0, len(sensors))
	for _, sensor := range sensors {
		data = append(data, NewSensorResponse(sensor))
	}

	response := SensorListResponse{
		Data:       data,
		Pagination: newPaginationMetadata(limit, offset, total),
	}

	if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// handleGetLatestSensor handles GET /sensor/latest
// Returns the current (active) sensor
func (s *Server) handleGetLatestSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sensor, err := s.sensorService.GetCurrentSensor(ctx)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "No active sensor found")
			return
		}
		handleError(w, err, s.logger)
		return
	}

	response := LatestSensorResponse{
		Data: NewSensorResponse(sensor),
	}

	if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// handleGetSensorStatistics handles GET /sensor/stats
func (s *Server) handleGetSensorStatistics(w http.ResponseWriter, r *http.Request) {
	// Parse time range (optional)
	start, end, err := parseStatisticsParams(r)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats, err := s.sensorService.GetStatistics(ctx, start, end)
	if err != nil {
		handleError(w, err, s.logger)
		return
	}

	// Get current sensor (optional)
	var currentResp *SensorResponse
	currentSensor, err := s.sensorService.GetCurrentSensor(ctx)
	if err == nil && currentSensor != nil {
		currentResp = NewSensorResponse(currentSensor)
	}

	// Build response with period info
	var periodInfo *PeriodInfo
	if start != nil && end != nil {
		periodInfo = &PeriodInfo{
			Start: start.Format(time.RFC3339),
			End:   end.Format(time.RFC3339),
		}
	}

	response := SensorStatisticsResponse{
		Data: SensorStatisticsData{
			Period:     periodInfo,
			Statistics: *stats,
			Current:    currentResp,
		},
	}

	if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// handleHealth handles GET /health
// Returns daemon health status with appropriate HTTP status code
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Get health status from daemon
	healthStatus := s.getHealthStatus()

	// Add database health check
	healthStatus.DatabaseConnected = s.getDatabaseHealth()

	// Determine HTTP status code based on daemon and database status
	statusCode := http.StatusOK
	if !healthStatus.DatabaseConnected || healthStatus.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if healthStatus.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	} else if healthStatus.Status == "healthy" {
		statusCode = http.StatusOK
	} else {
		// Unknown status - treat as unhealthy
		statusCode = http.StatusServiceUnavailable
	}

	// Build unified response
	response := HealthResponse{
		Data: healthStatus,
	}

	if err := writeJSONResponse(w, statusCode, response); err != nil {
		s.logger.Error("failed to write health response", "error", err)
	}
}

// handleMetrics handles GET /metrics
// Returns runtime metrics including memory, goroutines, and system info
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// SSE metrics
	sseMetrics := SSEMetrics{Enabled: false, Subscribers: 0}
	if s.eventBroker != nil {
		sseMetrics = SSEMetrics{
			Enabled:     true,
			Subscribers: s.eventBroker.SubscriberCount(),
		}
	}

	metricsData := MetricsData{
		Uptime:     time.Since(s.startTime).String(),
		Goroutines: runtime.NumGoroutine(),
		Memory: MemoryStats{
			AllocMB:      m.Alloc / 1024 / 1024,
			TotalAllocMB: m.TotalAlloc / 1024 / 1024,
			SysMB:        m.Sys / 1024 / 1024,
			NumGC:        m.NumGC,
		},
		Runtime: RuntimeInfo{
			Version: runtime.Version(),
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
		Process: ProcessInfo{
			PID: os.Getpid(),
		},
		SSE: sseMetrics,
	}

	// Database pool stats
	if s.getDatabasePoolStats != nil {
		metricsData.Database = s.getDatabasePoolStats()
	}

	response := MetricsResponse{
		Data: metricsData,
	}

	if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write metrics response", "error", err)
	}
}
