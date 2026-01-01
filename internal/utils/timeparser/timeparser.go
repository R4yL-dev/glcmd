package timeparser

import (
	"fmt"
	"time"
)

// LibreView API timestamp format: "1/1/2026 2:52:27 PM"
// Go reference time: "Mon Jan 2 15:04:05 MST 2006"
const libreViewLayout = "1/2/2006 3:04:05 PM"

// ParseLibreViewTimestamp parses a timestamp string from LibreView API
// and returns a time.Time in UTC.
//
// Format expected: "M/d/yyyy h:mm:ss tt" (example: "1/1/2026 2:52:27 PM")
//
// The input timestamp is assumed to be in the local timezone of the user's device.
// The returned time.Time is converted to UTC for consistent storage.
func ParseLibreViewTimestamp(timestamp string) (time.Time, error) {
	if timestamp == "" {
		return time.Time{}, fmt.Errorf("timestamp is empty")
	}

	// Parse using the LibreView format
	t, err := time.Parse(libreViewLayout, timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp '%s': %w", timestamp, err)
	}

	// Convert to UTC for consistent storage
	// Note: time.Parse assumes local timezone by default
	// We convert to UTC to avoid DST issues and timezone inconsistencies
	return t.UTC(), nil
}

// MustParseLibreViewTimestamp is like ParseLibreViewTimestamp but panics on error.
// Useful for test data initialization.
func MustParseLibreViewTimestamp(timestamp string) time.Time {
	t, err := ParseLibreViewTimestamp(timestamp)
	if err != nil {
		panic(err)
	}
	return t
}
