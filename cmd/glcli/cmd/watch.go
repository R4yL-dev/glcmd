package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var (
	onlyFlag    string
	verboseFlag bool
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Stream real-time events (glucose measurements, sensor changes)",
	Long: `Stream events from glcore in real-time using Server-Sent Events (SSE).

By default, streams all event types (glucose, sensor).
Keepalive events are hidden by default. Use --verbose to show them.

Examples:
  glcli watch                  # All events
  glcli watch --only glucose   # Glucose only
  glcli watch --only sensor    # Sensor changes only
  glcli watch --json           # JSON output for scripting
  glcli watch --verbose        # Show keepalive events`,
	Run: runWatch,
}

func init() {
	watchCmd.Flags().StringVar(&onlyFlag, "only", "", "Filter by event type (glucose, sensor)")
	watchCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show keepalive events")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) {
	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nDisconnecting...")
		cancel()
	}()

	// Parse type filter
	var types []string
	if onlyFlag != "" {
		types = []string{onlyFlag}
	}

	// Connect to SSE stream
	events, errors := client.Stream(ctx, types)

	if !jsonOutput {
		fmt.Println("Watching for events... (Ctrl+C to stop)")
		fmt.Println()
	}

	// Process events
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			formatEvent(event, jsonOutput, verboseFlag)
		case err, ok := <-errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		case <-ctx.Done():
			return
		}
	}
}

func formatEvent(event cli.SSEEvent, jsonMode bool, verbose bool) {
	// Filter keepalives if not verbose
	if event.Type == "keepalive" && !verbose {
		return
	}

	if jsonMode {
		// JSON mode: output raw event
		output := map[string]interface{}{
			"type":      event.Type,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		// Parse data if present
		if len(event.Data) > 0 && string(event.Data) != "{}" {
			var data interface{}
			if err := json.Unmarshal(event.Data, &data); err == nil {
				output["data"] = data
			} else {
				output["data"] = string(event.Data)
			}
		}

		jsonBytes, _ := json.Marshal(output)
		fmt.Println(string(jsonBytes))
		return
	}

	// Human-readable mode
	switch event.Type {
	case "glucose":
		formatGlucoseEvent(event.Data)
	case "sensor":
		formatSensorEvent(event.Data)
	case "keepalive":
		// Only shown if verbose (already filtered above)
		fmt.Printf("[%s] · keepalive\n", time.Now().Format("15:04:05"))
	default:
		fmt.Printf("[%s] Unknown event type: %s\n", time.Now().Format("15:04:05"), event.Type)
	}
}

func formatGlucoseEvent(data []byte) {
	var reading cli.GlucoseReading
	if err := json.Unmarshal(data, &reading); err != nil {
		fmt.Printf("[%s] Failed to parse glucose event\n", time.Now().Format("15:04:05"))
		return
	}

	timestamp := time.Now().Format("15:04:05")
	trend := cli.TrendArrowText(reading.TrendArrow)

	// Build status indicator
	status := "🟢"
	if reading.IsLow {
		status = "🟡 LOW"
	} else if reading.IsHigh {
		status = "🔴 HIGH"
	}

	if trend != "" {
		fmt.Printf("[%s] 🩸 %.1f mmol/L (%d mg/dL) %s %s\n",
			timestamp, reading.Value, reading.ValueInMgPerDl, trend, status)
	} else {
		fmt.Printf("[%s] 🩸 %.1f mmol/L (%d mg/dL) %s\n",
			timestamp, reading.Value, reading.ValueInMgPerDl, status)
	}
}

func formatSensorEvent(data []byte) {
	var sensor cli.SensorInfo
	if err := json.Unmarshal(data, &sensor); err != nil {
		fmt.Printf("[%s] Failed to parse sensor event\n", time.Now().Format("15:04:05"))
		return
	}

	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] 🔋 New sensor detected: %s\n", timestamp, sensor.SerialNumber)
	fmt.Printf("         Activation: %s\n", formatDateTime(sensor.Activation))
	fmt.Printf("         Expires: %s (%d days)\n", formatDateTime(sensor.ExpiresAt), sensor.DurationDays)
}

func formatDateTime(isoTimestamp string) string {
	// Parse and reformat for readability
	t, err := time.Parse(time.RFC3339, isoTimestamp)
	if err != nil {
		return isoTimestamp
	}
	return t.Local().Format("2006-01-02 15:04")
}
