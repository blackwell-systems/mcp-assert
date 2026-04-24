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
	case "init":
		if err := runner.Init(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "coverage":
		if err := runner.Coverage(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "snapshot":
		if err := runner.Snapshot(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "generate":
		if err := runner.Generate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "watch":
		if err := runner.Watch(os.Args[2:]); err != nil {
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
  mcp-assert init   [dir] [--server <cmd>] [--fixture <dir>]
  mcp-assert run    --suite <dir> [--server <cmd>] [--fixture <dir>] [--trials N]
  mcp-assert matrix --suite <dir> --languages <lang:server,...>
  mcp-assert ci     --suite <dir> [--server <cmd>] [--threshold N] [--fail-on-regression]
  mcp-assert coverage  --suite <dir> --server <cmd> [--coverage-json <path>]
  mcp-assert generate  --server <cmd> --output <dir> [--fixture <dir>]
  mcp-assert snapshot  --suite <dir> [--update] [--server <cmd>] [--fixture <dir>]
  mcp-assert watch     --suite <dir> [--server <cmd>] [--interval <duration>]

Commands:
  init      Scaffold a template, or generate a complete suite with --server
  run       Run assertions against an MCP server
  matrix    Run assertions across multiple language servers
  ci        Run assertions with CI-specific output and exit codes
  coverage  Show which server tools have assertions and which don't
  generate  Auto-generate stub assertions from a server's tools/list
  snapshot  Capture/compare tool response snapshots (like jest --updateSnapshot)
  watch     Rerun assertions when YAML files change (polling, no dependencies)

Flags:
  --suite <dir>          Directory containing assertion YAML files
  --server <cmd>         Override server command for all assertions (e.g. "agent-lsp go:gopls")
  --fixture <dir>        Fixture directory (substituted for {{fixture}} in assertions)
  --trials N             Number of trials per assertion (default: 1)
  --languages <spec>     Comma-separated lang:server pairs for matrix mode
  --threshold N          Minimum pass percentage for CI mode (default: 100)
  --docker <image>       Run MCP server inside a Docker container
  --baseline <path>      Baseline JSON for regression detection
  --save-baseline <path> Save current results as baseline
  --fail-on-regression   Exit 1 if a previously-passing assertion regresses (requires --baseline)
  --timeout <duration>   Per-assertion timeout (default: 30s)
  --json                 Output results as JSON
  --junit <path>         Write JUnit XML report to path
  --markdown <path>      Write markdown summary to path (auto-detects $GITHUB_STEP_SUMMARY in ci mode)
  --badge <path>         Write shields.io endpoint JSON to path
  --coverage-json <path> Write coverage data as JSON (coverage command)
  --interval <duration>  Polling interval for watch mode (default: 2s)

`)
}
