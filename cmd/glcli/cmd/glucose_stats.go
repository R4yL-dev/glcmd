package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var (
	statsPeriod string
	statsStart  string
	statsEnd    string
)

var glucoseStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show glucose statistics",
	Long: `Display glucose statistics for a time period.

Available periods: today, 7d, 14d, 30d, 90d, all

Examples:
  glcli glucose stats                # Today's statistics
  glcli glucose stats --period 7d    # Last 7 days
  glcli glucose stats --period 30d   # Last 30 days
  glcli glucose stats --period all   # All time
  glcli glucose stats --start 2025-01-01 --end 2025-01-17`,
	Run: runGlucoseStats,
}

func runGlucoseStats(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var start, end *time.Time
	now := time.Now()

	// Handle custom date range
	if statsStart != "" || statsEnd != "" {
		if statsStart != "" {
			s, err := parseDate(statsStart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid start date '%s': %v\n", statsStart, err)
				os.Exit(1)
			}
			start = &s
		}

		if statsEnd != "" {
			e, err := parseDate(statsEnd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid end date '%s': %v\n", statsEnd, err)
				os.Exit(1)
			}
			// Set end of day if only date provided
			if len(statsEnd) == 10 {
				e = e.Add(24*time.Hour - time.Second)
			}
			end = &e
		} else {
			end = &now
		}
	} else {
		// Handle predefined periods
		switch statsPeriod {
		case "today", "":
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
			start = &startOfDay
			end = &now
		case "7d":
			s := now.AddDate(0, 0, -7)
			start = &s
			end = &now
		case "14d":
			s := now.AddDate(0, 0, -14)
			start = &s
			end = &now
		case "30d":
			s := now.AddDate(0, 0, -30)
			start = &s
			end = &now
		case "90d":
			s := now.AddDate(0, 0, -90)
			start = &s
			end = &now
		case "all":
			// No start/end means all time
			start = nil
			end = nil
		default:
			fmt.Fprintf(os.Stderr, "Error: invalid period '%s'. Valid options: today, 7d, 14d, 30d, 90d, all\n", statsPeriod)
			os.Exit(1)
		}
	}

	result, err := client.GetStatistics(ctx, start, end)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		output, err := cli.FormatJSON(result.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(output)
	} else {
		fmt.Println(cli.FormatStatistics(&result.Data))
	}
}

func init() {
	glucoseStatsCmd.Flags().StringVar(&statsPeriod, "period", "today", "Time period (today, 7d, 14d, 30d, 90d, all)")
	glucoseStatsCmd.Flags().StringVar(&statsStart, "start", "", "Start date (YYYY-MM-DD)")
	glucoseStatsCmd.Flags().StringVar(&statsEnd, "end", "", "End date (YYYY-MM-DD)")
	glucoseCmd.AddCommand(glucoseStatsCmd)
}
