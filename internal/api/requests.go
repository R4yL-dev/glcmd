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

// parseTimeRange parses optional start/end query parameters as RFC3339 timestamps
// and validates that end is after start when both are provided.
func parseTimeRange(r *http.Request) (start, end *time.Time, err error) {
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		startTime, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return nil, nil, NewValidationError("invalid start time format (use RFC3339)")
		}
		start = &startTime
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		endTime, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return nil, nil, NewValidationError("invalid end time format (use RFC3339)")
		}
		end = &endTime
	}

	if start != nil && end != nil {
		if end.Before(*start) {
			return nil, nil, NewValidationError("end time must be after start time")
		}
	}

	return start, end, nil
}

// parseGlucoseFilters parses filter parameters from query string
func parseGlucoseFilters(r *http.Request) (repository.GlucoseFilters, error) {
	filters := repository.GlucoseFilters{}

	start, end, err := parseTimeRange(r)
	if err != nil {
		return filters, err
	}
	filters.StartTime = start
	filters.EndTime = end

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

	start, end, err := parseTimeRange(r)
	if err != nil {
		return filters, err
	}
	filters.StartTime = start
	filters.EndTime = end

	return filters, nil
}

// parseStatisticsParams parses and validates statistics request parameters.
// Returns nil for start/end if not provided (all time query).
// Both parameters must be provided together or not at all.
func parseStatisticsParams(r *http.Request) (start, end *time.Time, err error) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	// Both must be provided together
	if (startStr == "" && endStr != "") || (startStr != "" && endStr == "") {
		return nil, nil, NewValidationError("both start and end must be provided, or neither")
	}

	// Both empty = all time
	if startStr == "" && endStr == "" {
		return nil, nil, nil
	}

	return parseTimeRange(r)
}
