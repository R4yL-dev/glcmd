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
	historyPeriod string
	historyStart  string
	historyEnd    string
	historyLimit  int
)

var glucoseHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show glucose measurement history",
	Long: `Display historical glucose measurements.

By default, shows the last 50 measurements. Use flags to filter by time period.

Period formats:
  today   Since midnight
  Xh      Last X hours (e.g., 24h)
  Xd      Last X days (e.g., 7d, 14d, 30d)
  Xw      Last X weeks (e.g., 2w)
  Xm      Last X months (e.g., 1m, 3m)

Examples:
  glcli glucose history                 # Last 50 measurements
  glcli glucose history --period 24h    # Last 24 hours
  glcli glucose history --period 7d     # Last 7 days
  glcli glucose history --period 2w     # Last 2 weeks
  glcli glucose history --start 2025-01-10 --end 2025-01-17
  glcli glucose history --limit 100     # Change the limit`,
	Run: runGlucoseHistory,
}

func runGlucoseHistory(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := cli.GlucoseParams{
		Limit: historyLimit,
	}

	now := time.Now()

	// Handle --period flag
	if historyPeriod != "" {
		start, end, err := periodparser.Parse(historyPeriod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		params.Start = start
		params.End = end

		// If --period is specified without explicit --limit, use API max limit to get all data
		if !cmd.Flags().Changed("limit") {
			params.Limit = 1000
		}
	} else {
		// Handle --start/--end flags
		if historyStart != "" {
			start, err := periodparser.ParseDate(historyStart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			params.Start = &start
		}

		if historyEnd != "" {
			end, err := periodparser.ParseDate(historyEnd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

	result, err := client.GetGlucose(ctx, params)
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

func init() {
	glucoseHistoryCmd.Flags().StringVar(&historyPeriod, "period", "", "Relative period (e.g., today, 24h, 7d, 2w, 1m)")
	glucoseHistoryCmd.Flags().StringVar(&historyStart, "start", "", "Start date (YYYY-MM-DD)")
	glucoseHistoryCmd.Flags().StringVar(&historyEnd, "end", "", "End date (YYYY-MM-DD)")
	glucoseHistoryCmd.Flags().IntVar(&historyLimit, "limit", 50, "Maximum number of measurements")
	glucoseCmd.AddCommand(glucoseHistoryCmd)
}
