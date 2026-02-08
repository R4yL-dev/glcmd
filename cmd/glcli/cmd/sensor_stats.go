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
	sensorStatsPeriod string
	sensorStatsStart  string
	sensorStatsEnd    string
)

var sensorStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show sensor lifecycle statistics",
	Long: `Display sensor lifecycle statistics for a time period.

Shows total sensors, average/min/max duration, and comparison with expected duration.

Period formats:
  today   Since midnight
  Xh      Last X hours (e.g., 24h)
  Xd      Last X days (e.g., 7d, 14d, 30d, 90d)
  Xw      Last X weeks (e.g., 2w)
  Xm      Last X months (e.g., 1m, 3m)
  all     All time (default)

Examples:
  glcli sensor stats                 # All time statistics (default)
  glcli sensor stats --period 90d    # Last 90 days
  glcli sensor stats --period 6m     # Last 6 months
  glcli sensor stats --start 2025-01-01 --end 2025-06-01`,
	Run: runSensorStats,
}

func runSensorStats(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var start, end *time.Time

	// Handle custom date range
	if sensorStatsStart != "" || sensorStatsEnd != "" {
		if sensorStatsStart != "" {
			s, err := periodparser.ParseDate(sensorStatsStart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			start = &s
		}

		if sensorStatsEnd != "" {
			e, err := periodparser.ParseDate(sensorStatsEnd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Set end of day if only date provided
			if len(sensorStatsEnd) == 10 {
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
		start, end, err = periodparser.Parse(sensorStatsPeriod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	result, err := client.GetSensorStatistics(ctx, start, end)
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
		fmt.Println(cli.FormatSensorStats(&result.Data))
	}
}

func init() {
	sensorStatsCmd.Flags().StringVar(&sensorStatsPeriod, "period", "all", "Time period (today, Xh, Xd, Xw, Xm, all)")
	sensorStatsCmd.Flags().StringVar(&sensorStatsStart, "start", "", "Start date (YYYY-MM-DD)")
	sensorStatsCmd.Flags().StringVar(&sensorStatsEnd, "end", "", "End date (YYYY-MM-DD)")
	sensorCmd.AddCommand(sensorStatsCmd)
}
