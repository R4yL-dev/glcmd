package cmd

import (
	"os"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"

	// Global flags
	jsonOutput bool
	apiURL     string

	// Shared client (initialized in PersistentPreRun)
	client *cli.Client
)

var rootCmd = &cobra.Command{
	Use:   "glcli",
	Short: "Glucose monitoring CLI",
	Long: `glcli - Glucose monitoring CLI

A command-line interface for querying glucose readings and sensor
information from a glcore API server.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		client = cli.NewClient(apiURL)
	},
	// When called without subcommand, run glucose
	Run: func(cmd *cobra.Command, args []string) {
		glucoseCmd.Run(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Get default API URL from environment
	defaultAPIURL := os.Getenv("GLCMD_API_URL")
	if defaultAPIURL == "" {
		defaultAPIURL = "http://localhost:8080"
	}

	// Global persistent flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON (for scripting)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", defaultAPIURL, "API server URL")
}
