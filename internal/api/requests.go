package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/R4yL-dev/glcmd/internal/repository"
)

const (
	defaultLimit  = 100
	maxLimit      = 1000
	defaultOffset = 0
)

// parsePaginationParams parses limit and offset from query parameters
func parsePaginationParams(r *http.Request) (limit, offset int, err error) {
	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limit = defaultLimit
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, NewValidationError("invalid limit parameter")
		}
		if limit < 1 {
			return 0, 0, NewValidationError("limit must be at least 1")
		}
		if limit > maxLimit {
			return 0, 0, NewValidationError(fmt.Sprintf("limit must not exceed %d", maxLimit))
		}
	}

	// Parse offset
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offset = defaultOffset
	} else {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return 0, 0, NewValidationError("invalid offset parameter")
		}
		if offset < 0 {
			return 0, 0, NewValidationError("offset must be non-negative")
		}
	}

	return limit, offset, nil
}

// parseMeasurementFilters parses filter parameters from query string
func parseMeasurementFilters(r *http.Request) (repository.MeasurementFilters, error) {
	filters := repository.MeasurementFilters{}

	// Parse start time
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		startTime, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return filters, NewValidationError("invalid start time format (use RFC3339)")
		}
		filters.StartTime = &startTime
	}

	// Parse end time
	if endStr := r.URL.Query().Get("end"); endStr != "" {
		endTime, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return filters, NewValidationError("invalid end time format (use RFC3339)")
		}
		filters.EndTime = &endTime
	}

	// Validate time range
	if filters.StartTime != nil && filters.EndTime != nil {
		if filters.EndTime.Before(*filters.StartTime) {
			return filters, NewValidationError("end time must be after start time")
		}
	}

	// Parse color filter
	if colorStr := r.URL.Query().Get("color"); colorStr != "" {
		color, err := strconv.Atoi(colorStr)
		if err != nil {
			return filters, NewValidationError("invalid color parameter")
		}
		if color < 1 || color > 3 {
			return filters, NewValidationError("color must be 1 (normal), 2 (warning), or 3 (critical)")
		}
		filters.Color = &color
	}

	// Parse type filter
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		measurementType, err := strconv.Atoi(typeStr)
		if err != nil {
			return filters, NewValidationError("invalid type parameter")
		}
		if measurementType < 0 || measurementType > 1 {
			return filters, NewValidationError("type must be 0 (historical) or 1 (current)")
		}
		filters.Type = &measurementType
	}

	return filters, nil
}

// parseSensorFilters parses filter parameters for sensor queries
func parseSensorFilters(r *http.Request) (repository.SensorFilters, error) {
	filters := repository.SensorFilters{}

	if startStr := r.URL.Query().Get("start"); startStr != "" {
		startTime, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return filters, NewValidationError("invalid start time format (use RFC3339)")
		}
		filters.StartTime = &startTime
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		endTime, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return filters, NewValidationError("invalid end time format (use RFC3339)")
		}
		filters.EndTime = &endTime
	}

	if filters.StartTime != nil && filters.EndTime != nil {
		if filters.EndTime.Before(*filters.StartTime) {
			return filters, NewValidationError("end time must be after start time")
		}
	}

	return filters, nil
}

// parseStatisticsParams parses and validates statistics request parameters.
// Returns nil for start/end if not provided (all time query).
// Both parameters must be provided together or not at all.
func parseStatisticsParams(r *http.Request) (start, end *time.Time, err error) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	// Both empty = all time
	if startStr == "" && endStr == "" {
		return nil, nil, nil
	}

	// Both must be provided together
	if (startStr == "" && endStr != "") || (startStr != "" && endStr == "") {
		return nil, nil, NewValidationError("both start and end must be provided, or neither")
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return nil, nil, NewValidationError("invalid start time format (use RFC3339)")
	}

	// Parse end time
	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return nil, nil, NewValidationError("invalid end time format (use RFC3339)")
	}

	// Validate time range
	if endTime.Before(startTime) {
		return nil, nil, NewValidationError("end time must be after start time")
	}

	return &startTime, &endTime, nil
}
