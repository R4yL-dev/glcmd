package periodparser

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Parse parses a period string and returns start/end times.
// Supported formats:
//   - "today" -> since midnight
//   - "all" -> nil, nil (all time)
//   - "Xh" -> last X hours
//   - "Xd" -> last X days
//   - "Xw" -> last X weeks
//   - "Xm" -> last X months
func Parse(s string) (start, end *time.Time, err error) {
	now := time.Now()

	switch s {
	case "today":
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		return &startOfDay, &now, nil
	case "all":
		return nil, nil, nil
	}

	// Try parsing as duration (Xh, Xd, Xw, Xm)
	duration, err := ParseDuration(s)
	if err != nil {
		return nil, nil, err
	}

	startTime := now.Add(-duration)
	return &startTime, &now, nil
}

// ParseDuration parses duration strings like "24h", "7d", "2w", "1m".
func ParseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)([hdwm])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid period format '%s': expected today, all, or number + unit (h=hours, d=days, w=weeks, m=months)", s)
	}

	value, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "m":
		return time.Duration(value) * 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// ParseDate parses date strings in YYYY-MM-DD or YYYY-MM-DD HH:MM format.
func ParseDate(s string) (time.Time, error) {
	// Try full datetime format first
	if t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local); err == nil {
		return t, nil
	}

	// Try date only format
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format '%s': expected YYYY-MM-DD or YYYY-MM-DD HH:MM", s)
}
