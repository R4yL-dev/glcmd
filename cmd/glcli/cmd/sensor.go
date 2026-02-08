package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var sensorCmd = &cobra.Command{
	Use:   "sensor",
	Short: "Show current sensor information",
	Long: `Display information about the currently active sensor.

Shows serial number, days elapsed, days remaining, and expiration date.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sensor, err := client.GetLatestSensor(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			output, err := cli.FormatJSON(sensor)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else {
			fmt.Println(cli.FormatSensor(sensor))
		}
	},
}

func init() {
	rootCmd.AddCommand(sensorCmd)
}
