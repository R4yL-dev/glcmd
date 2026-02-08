package cmd

import (
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show glucose measurement history (alias for 'glucose history')",
	Long: `Display historical glucose measurements.

This is an alias for 'glcli glucose history'. See 'glcli glucose history --help' for details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Delegate to glucose history command
		runGlucoseHistory(cmd, args)
	},
}

func init() {
	// Mirror the same flags as glucoseHistoryCmd
	historyCmd.Flags().StringVar(&historyPeriod, "period", "", "Relative period (e.g., today, 24h, 7d, 2w, 1m)")
	historyCmd.Flags().StringVar(&historyStart, "start", "", "Start date (YYYY-MM-DD)")
	historyCmd.Flags().StringVar(&historyEnd, "end", "", "End date (YYYY-MM-DD)")
	historyCmd.Flags().IntVar(&historyLimit, "limit", 50, "Maximum number of measurements")
	rootCmd.AddCommand(historyCmd)
}
