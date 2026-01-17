package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/R4yL-dev/glcmd/internal/cli"
)

var version = "dev"

func main() {
	// Get API URL from environment or use default
	apiURL := os.Getenv("GLCMD_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	// Determine command and arguments
	command, args := parseCommand(os.Args[1:])

	switch command {
	case "glucose", "":
		runGlucose(apiURL, args)
	case "sensor":
		runSensor(apiURL, args)
	case "version":
		fmt.Printf("glcli %s\n", version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// parseCommand determines the command and remaining args.
// If no command is given or first arg is a flag, defaults to "glucose".
func parseCommand(args []string) (command string, remaining []string) {
	if len(args) == 0 {
		return "", nil
	}

	first := args[0]

	// Check for help/version flags
	if first == "-h" || first == "--help" {
		return "help", nil
	}
	if first == "-v" || first == "--version" {
		// Tricky: -v could be --version or --verbose for glucose
		// We check if it's alone (version) or with other args (verbose)
		// Actually, let's use -V for version to avoid ambiguity
		// For now, --version is version, -v is verbose for glucose
		if first == "--version" {
			return "version", nil
		}
		// -v alone or with args = glucose verbose
		return "", args
	}

	// If first arg starts with -, it's a flag -> glucose command
	if strings.HasPrefix(first, "-") {
		return "", args
	}

	// Otherwise, first arg is a command
	return first, args[1:]
}

func runGlucose(apiURL string, args []string) {
	fs := flag.NewFlagSet("glucose", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	verbose := fs.Bool("v", false, "Show detailed output (status, time)")
	fs.BoolVar(verbose, "verbose", false, "Show detailed output (status, time)")
	fs.Parse(args)

	client := cli.NewClient(apiURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reading, err := client.GetLatestGlucose(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		output, err := cli.FormatJSON(reading)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(output)
	} else if *verbose {
		fmt.Println(cli.FormatGlucose(reading))
	} else {
		fmt.Println(cli.FormatGlucoseShort(reading))
	}
}

func runSensor(apiURL string, args []string) {
	fs := flag.NewFlagSet("sensor", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	client := cli.NewClient(apiURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sensor, err := client.GetCurrentSensor(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		output, err := cli.FormatJSON(sensor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(output)
	} else {
		fmt.Println(cli.FormatSensor(sensor))
	}
}

func printUsage() {
	fmt.Print(`glcli - Glucose monitoring CLI

Usage:
  glcli [flags]              Show current glucose (default command)
  glcli sensor [flags]       Show current sensor information
  glcli version              Show version information
  glcli help                 Show this help message

Flags:
  -v, --verbose   Show detailed output (status, time)
  --json          Output as JSON (for scripting)

Environment Variables:
  GLCMD_API_URL   API server URL (default: http://localhost:8080)

Examples:
  glcli                  # Quick glucose check
  glcli -v               # Detailed glucose info
  glcli --json           # JSON output for scripting
  glcli sensor           # Show sensor info
`)
}
