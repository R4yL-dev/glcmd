package cmd

import (
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show glucose statistics (alias for 'glucose stats')",
	Long: `Display glucose statistics for a time period.

This is an alias for 'glcli glucose stats'. See 'glcli glucose stats --help' for details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Delegate to glucose stats command
		runGlucoseStats(cmd, args)
	},
}

func init() {
	// Mirror the same flags as glucoseStatsCmd
	statsCmd.Flags().StringVar(&statsPeriod, "period", "today", "Time period (today, 7d, 14d, 30d, 90d, all)")
	statsCmd.Flags().StringVar(&statsStart, "start", "", "Start date (YYYY-MM-DD)")
	statsCmd.Flags().StringVar(&statsEnd, "end", "", "End date (YYYY-MM-DD)")
	rootCmd.AddCommand(statsCmd)
}
