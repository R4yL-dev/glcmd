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
func FormatSensor(s *SensorInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Sensor: %s\n", s.SerialNumber))

	if s.DaysRemaining != nil {
		sb.WriteString(fmt.Sprintf("Active: %.1f days elapsed, %.1f days remaining\n",
			s.DaysElapsed, *s.DaysRemaining))
	} else {
		sb.WriteString(fmt.Sprintf("Days Used: %.1f\n", s.DaysElapsed))
	}

	// Extract just the date part from ExpiresAt (format: 2006-01-02T15:04:05Z)
	expiresDate := s.ExpiresAt
	if len(expiresDate) >= 10 {
		expiresDate = expiresDate[:10]
	}
	sb.WriteString(fmt.Sprintf("Expires: %s", expiresDate))

	return sb.String()
}

// FormatJSON formats any value as indented JSON
func FormatJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
