package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TrendArrowText returns emoji + text for trend arrow
func TrendArrowText(arrow *int) string {
	if arrow == nil {
		return ""
	}

	switch *arrow {
	case 1:
		return "⬇️⬇️ Falling Rapidly"
	case 2:
		return "⬇️ Falling"
	case 3:
		return "➡️ Stable"
	case 4:
		return "⬆️ Rising"
	case 5:
		return "⬆️⬆️ Rising Rapidly"
	default:
		return "? Unknown"
	}
}

// FormatGlucoseShort formats a glucose reading as a single line
func FormatGlucoseShort(g *GlucoseReading) string {
	trend := TrendArrowText(g.TrendArrow)
	if trend != "" {
		return fmt.Sprintf("%.1f mmol/L (%d mg/dL) %s", g.Value, g.ValueInMgPerDl, trend)
	}
	return fmt.Sprintf("%.1f mmol/L (%d mg/dL)", g.Value, g.ValueInMgPerDl)
}

// FormatGlucose formats a glucose reading with full details
func FormatGlucose(g *GlucoseReading) string {
	var sb strings.Builder

	// Main value line with trend
	trend := TrendArrowText(g.TrendArrow)
	if trend != "" {
		sb.WriteString(fmt.Sprintf("Glucose: %.1f mmol/L (%d mg/dL) %s\n",
			g.Value, g.ValueInMgPerDl, trend))
	} else {
		sb.WriteString(fmt.Sprintf("Glucose: %.1f mmol/L (%d mg/dL)\n",
			g.Value, g.ValueInMgPerDl))
	}

	// Status line
	status := "Normal"
	if g.IsLow {
		status = "LOW"
	} else if g.IsHigh {
		status = "HIGH"
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	// Timestamp
	sb.WriteString(fmt.Sprintf("Time: %s", g.Timestamp.Local().Format("15:04:05")))

	return sb.String()
}

// FormatSensor formats sensor info for human display
// Priority: most important info first (remaining time for active, status for problematic)
func FormatSensor(s *SensorInfo) string {
	var sb strings.Builder

	// Format expiration datetime (2006-01-02T15:04:05Z -> 2006-01-02 15:04)
	expiresDateTime := formatDateTime(s.ExpiresAt)

	switch s.Status {
	case "running":
		// Active sensor: remaining time is most important
		if s.DaysRemaining != nil {
			sb.WriteString(fmt.Sprintf("%.1f days remaining | %.1f days active\n",
				*s.DaysRemaining, s.DaysElapsed))
		}
		sb.WriteString(fmt.Sprintf("Sensor: %s\n", s.SerialNumber))
		sb.WriteString(fmt.Sprintf("Expires: %s", expiresDateTime))

	case "unresponsive":
		// Unresponsive: alert first, then time info
		if s.LastMeasurementAt != nil {
			lastDateTime := formatDateTime(*s.LastMeasurementAt)
			sb.WriteString(fmt.Sprintf("Unresponsive since %s\n", lastDateTime))
		} else {
			sb.WriteString("Unresponsive\n")
		}
		if s.DaysRemaining != nil {
			sb.WriteString(fmt.Sprintf("%.1f days remaining | %.1f days active\n",
				*s.DaysRemaining, s.DaysElapsed))
		}
		sb.WriteString(fmt.Sprintf("Sensor: %s\n", s.SerialNumber))
		sb.WriteString(fmt.Sprintf("Expires: %s", expiresDateTime))

	case "expired":
		// Expired: status with datetime
		sb.WriteString(fmt.Sprintf("Expired: %s\n", expiresDateTime))
		sb.WriteString(fmt.Sprintf("Sensor: %s", s.SerialNumber))

	case "ended":
		// Ended: show when it ended
		sb.WriteString(fmt.Sprintf("Ended | %.1f days used\n", s.DaysElapsed))
		sb.WriteString(fmt.Sprintf("Sensor: %s", s.SerialNumber))

	default:
		// Fallback
		sb.WriteString(fmt.Sprintf("Sensor: %s\n", s.SerialNumber))
		if s.DaysRemaining != nil {
			sb.WriteString(fmt.Sprintf("%.1f days remaining | %.1f days active",
				*s.DaysRemaining, s.DaysElapsed))
		} else {
			sb.WriteString(fmt.Sprintf("%.1f days active", s.DaysElapsed))
		}
	}

	return sb.String()
}

// formatDateTime converts ISO timestamp to readable format (2006-01-02 15:04)
func formatDateTime(isoTimestamp string) string {
	// Input: 2006-01-02T15:04:05Z -> Output: 2006-01-02 15:04
	if len(isoTimestamp) >= 16 {
		return isoTimestamp[:10] + " " + isoTimestamp[11:16]
	}
	if len(isoTimestamp) >= 10 {
		return isoTimestamp[:10]
	}
	return isoTimestamp
}

// FormatJSON formats any value as indented JSON
func FormatJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatMeasurementTable formats a list of measurements as a table
func FormatMeasurementTable(measurements []GlucoseReading, total int) string {
	if len(measurements) == 0 {
		return "No measurements found"
	}

	var sb strings.Builder

	// Table header
	sb.WriteString("┌─────────────────────┬───────────────┬──────────────────┬────────┐\n")
	sb.WriteString("│ Date                │ mmol/L (mg/dL)│ Trend            │ Status │\n")
	sb.WriteString("├─────────────────────┼───────────────┼──────────────────┼────────┤\n")

	// Table rows
	for _, m := range measurements {
		date := m.Timestamp.Local().Format("02/01 15:04")
		glucose := fmt.Sprintf("%.1f (%d)", m.Value, m.ValueInMgPerDl)
		trend := formatTrendShort(m.TrendArrow)
		status := formatStatus(m.IsLow, m.IsHigh)

		sb.WriteString(fmt.Sprintf("│ %-19s │ %-13s │ %-16s │ %-6s │\n",
			date, glucose, trend, status))
	}

	// Table footer
	sb.WriteString("└─────────────────────┴───────────────┴──────────────────┴────────┘\n")

	// Summary line
	if total > len(measurements) {
		sb.WriteString(fmt.Sprintf("Showing %d of %d measurements\n", len(measurements), total))
	} else {
		sb.WriteString(fmt.Sprintf("Showing %d measurements\n", len(measurements)))
	}

	// Legend for status symbols
	sb.WriteString("Status: ✓ Normal | ⚠ LOW | ⚠ HIGH")

	return sb.String()
}

// formatTrendShort returns a short trend representation for table display
func formatTrendShort(arrow *int) string {
	if arrow == nil {
		return "-"
	}

	switch *arrow {
	case 1:
		return "⬇️⬇️ Falling Fast"
	case 2:
		return "⬇️  Falling"
	case 3:
		return "➡️  Stable"
	case 4:
		return "⬆️  Rising"
	case 5:
		return "⬆️⬆️ Rising Fast"
	default:
		return "?"
	}
}

// formatStatus returns a status indicator
func formatStatus(isLow, isHigh bool) string {
	if isLow {
		return "⚠ LOW"
	}
	if isHigh {
		return "⚠ HIGH"
	}
	return "✓"
}

// FormatStatistics formats statistics data for display
func FormatStatistics(stats *StatisticsData) string {
	var sb strings.Builder

	// Header - format the period dates
	periodLabel := fmt.Sprintf("%s to %s", formatDateShort(stats.Period.Start), formatDateShort(stats.Period.End))
	sb.WriteString(fmt.Sprintf("📊 Glucose Statistics (%s)\n", periodLabel))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Summary section
	sb.WriteString("📈 Summary\n")
	sb.WriteString(fmt.Sprintf("   Measurements: %d\n", stats.Statistics.Count))
	sb.WriteString(fmt.Sprintf("   Average:      %.1f mmol/L (%.0f mg/dL)\n",
		stats.Statistics.Average, stats.Statistics.AverageMgDl))
	sb.WriteString(fmt.Sprintf("   Range:        %.1f - %.1f mmol/L (%d - %d mg/dL)\n",
		stats.Statistics.Min, stats.Statistics.Max,
		stats.Statistics.MinMgDl, stats.Statistics.MaxMgDl))
	sb.WriteString(fmt.Sprintf("   Std Dev:      %.1f mmol/L\n", stats.Statistics.StdDev))
	sb.WriteString("\n")

	// Distribution section - calculate percentages from counts
	total := stats.Statistics.Count
	lowPct := 0.0
	normalPct := 0.0
	highPct := 0.0
	if total > 0 {
		lowPct = float64(stats.Distribution.Low) / float64(total) * 100
		normalPct = float64(stats.Distribution.Normal) / float64(total) * 100
		highPct = float64(stats.Distribution.High) / float64(total) * 100
	}

	sb.WriteString("📊 Distribution\n")
	sb.WriteString(fmt.Sprintf("   🟢 Normal:    %d (%.1f%%)\n", stats.Distribution.Normal, normalPct))
	sb.WriteString(fmt.Sprintf("   🟡 Low:       %d (%.1f%%)\n", stats.Distribution.Low, lowPct))
	sb.WriteString(fmt.Sprintf("   🔴 High:      %d (%.1f%%)\n", stats.Distribution.High, highPct))
	sb.WriteString("\n")

	// Time in Range section - use data from TimeInRange object if available
	sb.WriteString("🎯 Time in Range\n")
	if stats.TimeInRange != nil {
		sb.WriteString(fmt.Sprintf("   %s %.1f%%\n",
			formatProgressBar(stats.TimeInRange.InRange, 24), stats.TimeInRange.InRange))
		sb.WriteString(fmt.Sprintf("   ⬇️  Below: %.1f%%  |  ⬆️  Above: %.1f%%\n",
			stats.TimeInRange.BelowRange, stats.TimeInRange.AboveRange))
		sb.WriteString(fmt.Sprintf("   Target: %d-%d mg/dL",
			stats.TimeInRange.TargetLowMgDl, stats.TimeInRange.TargetHighMgDl))
	} else {
		sb.WriteString("   No glucose targets configured")
	}

	return sb.String()
}

// formatDateShort extracts just the date part from an ISO timestamp
func formatDateShort(isoDate string) string {
	if len(isoDate) >= 10 {
		return isoDate[:10]
	}
	return isoDate
}

// formatProgressBar creates a simple progress bar
func formatProgressBar(percent float64, width int) string {
	filled := int(percent / 100.0 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}
