package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var verbose bool

var glucoseCmd = &cobra.Command{
	Use:   "glucose",
	Short: "Show current glucose reading",
	Long: `Display the latest glucose reading from the sensor.

By default, shows a compact one-line output with value and trend.
Use --verbose for detailed output including status and timestamp.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		reading, err := client.GetLatestGlucose(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			output, err := cli.FormatJSON(reading)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else if verbose {
			fmt.Println(cli.FormatGlucose(reading))
		} else {
			fmt.Println(cli.FormatGlucoseShort(reading))
		}
	},
}

func init() {
	glucoseCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output (status, time)")
	rootCmd.AddCommand(glucoseCmd)
}
