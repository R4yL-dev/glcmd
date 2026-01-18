package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var (
	historyLast  string
	historyStart string
	historyEnd   string
	historyLimit int
)

var glucoseHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show glucose measurement history",
	Long: `Display historical glucose measurements.

By default, shows the last 50 measurements. Use flags to filter by time period.

Examples:
  glcli glucose history              # Last 50 measurements
  glcli glucose history --last 24h   # Last 24 hours
  glcli glucose history --last 7d    # Last 7 days
  glcli glucose history --start 2025-01-10 --end 2025-01-17
  glcli glucose history --limit 100  # Change the limit`,
	Run: runGlucoseHistory,
}

func runGlucoseHistory(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := cli.MeasurementParams{
		Limit: historyLimit,
	}

	// Parse time parameters
	now := time.Now()

	if historyLast != "" {
		duration, err := parseDuration(historyLast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid duration '%s': %v\n", historyLast, err)
			os.Exit(1)
		}
		start := now.Add(-duration)
		params.Start = &start
		params.End = &now

		// If --last is specified without explicit --limit, use API max limit to get all data
		if !cmd.Flags().Changed("limit") {
			params.Limit = 1000
		}
	} else {
		if historyStart != "" {
			start, err := parseDate(historyStart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid start date '%s': %v\n", historyStart, err)
				os.Exit(1)
			}
			params.Start = &start
		}

		if historyEnd != "" {
			end, err := parseDate(historyEnd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid end date '%s': %v\n", historyEnd, err)
				os.Exit(1)
			}
			// Set end of day if only date provided
			if len(historyEnd) == 10 {
				end = end.Add(24*time.Hour - time.Second)
			}
			params.End = &end
		} else if params.Start != nil {
			// If start is set but not end, use now
			params.End = &now
		}
	}

	result, err := client.GetMeasurements(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		output, err := cli.FormatJSON(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(output)
	} else {
		fmt.Println(cli.FormatMeasurementTable(result.Data, result.Pagination.Total))
	}
}

// parseDuration parses duration strings like "24h", "7d", "2w", "1m"
func parseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)([hdwm])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("expected format: number + unit (h=hours, d=days, w=weeks, m=months)")
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

// parseDate parses date strings in YYYY-MM-DD or YYYY-MM-DD HH:MM format
func parseDate(s string) (time.Time, error) {
	// Try full datetime format first
	if t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local); err == nil {
		return t, nil
	}

	// Try date only format
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("expected format: YYYY-MM-DD or YYYY-MM-DD HH:MM")
}

func init() {
	glucoseHistoryCmd.Flags().StringVar(&historyLast, "last", "", "Relative period (e.g., 24h, 7d, 2w)")
	glucoseHistoryCmd.Flags().StringVar(&historyStart, "start", "", "Start date (YYYY-MM-DD)")
	glucoseHistoryCmd.Flags().StringVar(&historyEnd, "end", "", "End date (YYYY-MM-DD)")
	glucoseHistoryCmd.Flags().IntVar(&historyLimit, "limit", 50, "Maximum number of measurements")
	glucoseCmd.AddCommand(glucoseHistoryCmd)
}
