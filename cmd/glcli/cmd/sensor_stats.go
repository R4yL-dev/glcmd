package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var sensorStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show sensor lifecycle statistics",
	Long: `Display sensor lifecycle statistics.

Shows total sensors, average/min/max duration, and comparison with expected duration.

Examples:
  glcli sensor stats          # Sensor statistics
  glcli sensor stats --json   # JSON output`,
	Run: runSensorStats,
}

func runSensorStats(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.GetSensorStatistics(ctx)
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
	sensorCmd.AddCommand(sensorStatsCmd)
}
