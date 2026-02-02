package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
	"github.com/spf13/cobra"
)

var glucoseGmiCmd = &cobra.Command{
	Use:   "gmi",
	Short: "Show Glucose Management Indicator (GMI)",
	Long: `Display the Glucose Management Indicator (GMI) for 7, 14, 30 and 90 days.

GMI is an estimated A1C derived from continuous glucose monitoring data.
Formula: GMI(%) = 3.31 + 0.02392 × [mean glucose in mg/dL]

Examples:
  glcli glucose gmi
  glcli gmi
  glcli gmi --json`,
	Run: runGlucoseGmi,
}

func runGlucoseGmi(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	periods := []struct {
		label string
		days  int
	}{
		{"7 days", 7},
		{"14 days", 14},
		{"30 days", 30},
		{"90 days", 90},
	}

	type periodResult struct {
		index  int
		result *cli.StatisticsResponse
		err    error
	}

	var wg sync.WaitGroup
	ch := make(chan periodResult, len(periods))

	for i, p := range periods {
		wg.Add(1)
		go func(idx, days int) {
			defer wg.Done()
			start := now.AddDate(0, 0, -days)
			end := now
			res, err := client.GetStatistics(ctx, &start, &end)
			ch <- periodResult{index: idx, result: res, err: err}
		}(i, p.days)
	}

	wg.Wait()
	close(ch)

	// Collect results in order
	results := make([]*cli.StatisticsResponse, len(periods))
	for pr := range ch {
		if pr.err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s stats: %v\n", periods[pr.index].label, pr.err)
			os.Exit(1)
		}
		results[pr.index] = pr.result
	}

	if jsonOutput {
		gmiResults := make([]cli.GMIPeriodResult, len(periods))
		for i, p := range periods {
			stats := results[i].Data.Statistics
			gmiResults[i] = cli.GMIPeriodResult{
				Label:        p.label,
				GMI:          stats.GMI,
				AverageMmol:  stats.Average,
				AverageMgDl:  stats.AverageMgDl,
				Measurements: stats.Count,
			}
		}
		output, err := cli.FormatJSON(gmiResults)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(output)
	} else {
		gmiResults := make([]cli.GMIPeriodResult, len(periods))
		for i, p := range periods {
			stats := results[i].Data.Statistics
			gmiResults[i] = cli.GMIPeriodResult{
				Label:        p.label,
				GMI:          stats.GMI,
				AverageMmol:  stats.Average,
				AverageMgDl:  stats.AverageMgDl,
				Measurements: stats.Count,
			}
		}
		fmt.Println(cli.FormatGMI(gmiResults))
	}
}

func init() {
	glucoseCmd.AddCommand(glucoseGmiCmd)
}
