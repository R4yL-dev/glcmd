package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/R4yL-dev/glcmd/internal/repository"
)

const (
	defaultLimit = 100
	maxLimit     = 1000
	defaultOffset = 0
	maxDaysRange = 90
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

// parseStatisticsParams parses and validates statistics request parameters
func parseStatisticsParams(r *http.Request) (start, end time.Time, err error) {
	// Start time is required
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		return time.Time{}, time.Time{}, NewValidationError("start parameter is required")
	}
	start, err = time.Parse(time.RFC3339, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, NewValidationError("invalid start time format (use RFC3339)")
	}

	// End time is required
	endStr := r.URL.Query().Get("end")
	if endStr == "" {
		return time.Time{}, time.Time{}, NewValidationError("end parameter is required")
	}
	end, err = time.Parse(time.RFC3339, endStr)
	if err != nil {
		return time.Time{}, time.Time{}, NewValidationError("invalid end time format (use RFC3339)")
	}

	// Validate time range
	if end.Before(start) {
		return time.Time{}, time.Time{}, NewValidationError("end time must be after start time")
	}

	// Validate max range (90 days)
	duration := end.Sub(start)
	maxDuration := time.Duration(maxDaysRange) * 24 * time.Hour
	if duration > maxDuration {
		return time.Time{}, time.Time{}, NewValidationError(fmt.Sprintf("time range must not exceed %d days", maxDaysRange))
	}

	return start, end, nil
}
