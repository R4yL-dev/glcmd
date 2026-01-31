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
	sensorHistoryLast  string
	sensorHistoryStart string
	sensorHistoryEnd   string
	sensorHistoryLimit int
)

var sensorHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show sensor history",
	Long: `Display historical sensor information.

By default, shows the last 50 sensors. Use flags to filter by activation date.

Examples:
  glcli sensor history              # Last 50 sensors
  glcli sensor history --last 3m    # Sensors activated in last 3 months
  glcli sensor history --last 6m    # Sensors activated in last 6 months
  glcli sensor history --start 2025-01-01 --end 2025-06-01
  glcli sensor history --limit 10   # Change the limit`,
	Run: runSensorHistory,
}

func runSensorHistory(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := cli.SensorHistoryParams{
		Limit: sensorHistoryLimit,
	}

	now := time.Now()

	if sensorHistoryLast != "" {
		duration, err := parseDuration(sensorHistoryLast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid duration '%s': %v\n", sensorHistoryLast, err)
			os.Exit(1)
		}
		start := now.Add(-duration)
		params.Start = &start
		params.End = &now

		if !cmd.Flags().Changed("limit") {
			params.Limit = 1000
		}
	} else {
		if sensorHistoryStart != "" {
			start, err := parseDate(sensorHistoryStart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid start date '%s': %v\n", sensorHistoryStart, err)
				os.Exit(1)
			}
			params.Start = &start
		}

		if sensorHistoryEnd != "" {
			end, err := parseDate(sensorHistoryEnd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid end date '%s': %v\n", sensorHistoryEnd, err)
				os.Exit(1)
			}
			if len(sensorHistoryEnd) == 10 {
				end = end.Add(24*time.Hour - time.Second)
			}
			params.End = &end
		} else if params.Start != nil {
			params.End = &now
		}
	}

	result, err := client.GetSensorHistory(ctx, params)
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
	sensorHistoryCmd.Flags().StringVar(&sensorHistoryLast, "last", "", "Relative period (e.g., 24h, 7d, 2w, 3m)")
	sensorHistoryCmd.Flags().StringVar(&sensorHistoryStart, "start", "", "Start date (YYYY-MM-DD)")
	sensorHistoryCmd.Flags().StringVar(&sensorHistoryEnd, "end", "", "End date (YYYY-MM-DD)")
	sensorHistoryCmd.Flags().IntVar(&sensorHistoryLimit, "limit", 50, "Maximum number of sensors")
	sensorCmd.AddCommand(sensorHistoryCmd)
}
