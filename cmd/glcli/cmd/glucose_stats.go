package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/R4yL-dev/glcmd/internal/utils/periodparser"
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

Period formats:
  today   Since midnight (default)
  Xh      Last X hours (e.g., 24h)
  Xd      Last X days (e.g., 7d, 14d, 30d, 90d)
  Xw      Last X weeks (e.g., 2w)
  Xm      Last X months (e.g., 1m, 3m)
  all     All time

Examples:
  glcli glucose stats                 # Today's statistics
  glcli glucose stats --period 7d     # Last 7 days
  glcli glucose stats --period 3d     # Last 3 days
  glcli glucose stats --period 30d    # Last 30 days
  glcli glucose stats --period all    # All time
  glcli glucose stats --start 2025-01-01 --end 2025-01-17`,
	Run: runGlucoseStats,
}

func runGlucoseStats(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var start, end *time.Time

	// Handle custom date range
	if statsStart != "" || statsEnd != "" {
		if statsStart != "" {
			s, err := periodparser.ParseDate(statsStart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			start = &s
		}

		if statsEnd != "" {
			e, err := periodparser.ParseDate(statsEnd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Set end of day if only date provided
			if len(statsEnd) == 10 {
				e = e.Add(24*time.Hour - time.Second)
			}
			end = &e
		} else {
			now := time.Now()
			end = &now
		}
	} else {
		// Handle predefined periods with periodparser
		var err error
		start, end, err = periodparser.Parse(statsPeriod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	result, err := client.GetGlucoseStatistics(ctx, start, end)
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
	glucoseStatsCmd.Flags().StringVar(&statsPeriod, "period", "today", "Time period (today, Xh, Xd, Xw, Xm, all)")
	glucoseStatsCmd.Flags().StringVar(&statsStart, "start", "", "Start date (YYYY-MM-DD)")
	glucoseStatsCmd.Flags().StringVar(&statsEnd, "end", "", "End date (YYYY-MM-DD)")
	glucoseCmd.AddCommand(glucoseStatsCmd)
}
