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
	sensorHistoryPeriod string
	sensorHistoryStart  string
	sensorHistoryEnd    string
	sensorHistoryLimit  int
)

var sensorHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show sensor history",
	Long: `Display historical sensor information.

By default, shows the last 50 sensors. Use flags to filter by activation date.

Period formats:
  today   Since midnight
  Xh      Last X hours (e.g., 24h)
  Xd      Last X days (e.g., 7d, 14d, 30d)
  Xw      Last X weeks (e.g., 2w)
  Xm      Last X months (e.g., 1m, 3m)

Examples:
  glcli sensor history                # Last 50 sensors
  glcli sensor history --period 3m    # Sensors activated in last 3 months
  glcli sensor history --period 6m    # Sensors activated in last 6 months
  glcli sensor history --start 2025-01-01 --end 2025-06-01
  glcli sensor history --limit 10     # Change the limit`,
	Run: runSensorHistory,
}

func runSensorHistory(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := cli.SensorParams{
		Limit: sensorHistoryLimit,
	}

	now := time.Now()

	// Handle --period flag
	if sensorHistoryPeriod != "" {
		start, end, err := periodparser.Parse(sensorHistoryPeriod)
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
		if sensorHistoryStart != "" {
			start, err := periodparser.ParseDate(sensorHistoryStart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			params.Start = &start
		}

		if sensorHistoryEnd != "" {
			end, err := periodparser.ParseDate(sensorHistoryEnd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Set end of day if only date provided
			if len(sensorHistoryEnd) == 10 {
				end = end.Add(24*time.Hour - time.Second)
			}
			params.End = &end
		} else if params.Start != nil {
			// If start is set but not end, use now
			params.End = &now
		}
	}

	result, err := client.GetSensor(ctx, params)
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
		fmt.Println(cli.FormatSensorTable(result.Data, result.Pagination.Total))
	}
}

func init() {
	sensorHistoryCmd.Flags().StringVar(&sensorHistoryPeriod, "period", "", "Relative period (e.g., today, 24h, 7d, 2w, 1m)")
	sensorHistoryCmd.Flags().StringVar(&sensorHistoryStart, "start", "", "Start date (YYYY-MM-DD)")
	sensorHistoryCmd.Flags().StringVar(&sensorHistoryEnd, "end", "", "End date (YYYY-MM-DD)")
	sensorHistoryCmd.Flags().IntVar(&sensorHistoryLimit, "limit", 50, "Maximum number of sensors")
	sensorCmd.AddCommand(sensorHistoryCmd)
}
