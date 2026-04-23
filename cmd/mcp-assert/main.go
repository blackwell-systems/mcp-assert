package main

import (
	"fmt"
	"os"

	"github.com/blackwell-systems/mcp-assert/internal/runner"
)

var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		if err := runner.Run(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "matrix":
		if err := runner.Matrix(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "ci":
		if err := runner.CI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "--version", "version":
		fmt.Printf("mcp-assert %s\n", Version)
	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`mcp-assert - Deterministic correctness testing for MCP servers

Usage:
  mcp-assert run    --suite <dir> [--fixture <dir>] [--trials N]
  mcp-assert matrix --suite <dir> --languages <lang:server,...>
  mcp-assert ci     --suite <dir> [--threshold N] [--fail-on-regression]

Commands:
  run       Run assertions against an MCP server
  matrix    Run assertions across multiple language servers
  ci        Run assertions with CI-specific output and exit codes

Flags:
  --suite <dir>          Directory containing assertion YAML files
  --fixture <dir>        Fixture directory (substituted for {{fixture}} in assertions)
  --trials N             Number of trials per assertion (default: 1)
  --languages <spec>     Comma-separated lang:server pairs for matrix mode
  --threshold N          Minimum pass percentage for CI mode (default: 100)
  --fail-on-regression   Exit 1 if any previously-passing assertion fails
  --docker <image>       Run each assertion in a Docker container
  --timeout <duration>   Per-assertion timeout (default: 30s)
  --json                 Output results as JSON

`)
}
