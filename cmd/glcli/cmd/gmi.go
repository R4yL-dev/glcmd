package cmd

import (
	"github.com/spf13/cobra"
)

var gmiCmd = &cobra.Command{
	Use:   "gmi",
	Short: "Show Glucose Management Indicator (alias for 'glucose gmi')",
	Long: `Display the Glucose Management Indicator (GMI) for 7, 14, 30 and 90 days.

This is an alias for 'glcli glucose gmi'. See 'glcli glucose gmi --help' for details.`,
	Run: func(cmd *cobra.Command, args []string) {
		runGlucoseGmi(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(gmiCmd)
}
