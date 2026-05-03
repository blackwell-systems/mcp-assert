// mcp-assert is a deterministic correctness testing tool for MCP servers.
//
// The CLI uses a simple subcommand dispatch: the first argument selects the
// command (run, ci, matrix, audit, etc.), and remaining arguments are passed
// to that command's flag parser. Each command is implemented in the runner
// package (commands.go and its siblings).
package main

import (
	"fmt"
	"os"

	"github.com/blackwell-systems/mcp-assert/internal/runner"
)

// Version is set at build time via -ldflags. Defaults to "dev" for local builds.
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	// Command registry: maps subcommand names to their entry points.
	// Adding a new command is a single line here plus the implementation
	// in the runner package.
	commands := map[string]func([]string) error{
		"run":       runner.Run,
		"matrix":    runner.Matrix,
		"ci":        runner.CI,
		"init":      runner.Init,
		"coverage":  runner.Coverage,
		"snapshot":  runner.Snapshot,
		"generate":  runner.Generate,
		"watch":     runner.Watch,
		"audit":     runner.Audit,
		"fuzz":      runner.Fuzz,
		"intercept": runner.Intercept,
		"lint":      runner.Lint,
	}

	cmd := os.Args[1]

	if fn, ok := commands[cmd]; ok {
		if err := fn(os.Args[2:]); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Special commands that don't follow the standard error-returning pattern.
	switch cmd {
	case "--version", "version":
		fmt.Printf("mcp-assert %s\n", Version)
	case "--help", "-h", "help":
		printUsage()
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

// printUsage writes the top-level help text to stdout.
func printUsage() {
	fmt.Print(`mcp-assert - Deterministic correctness testing for MCP servers

Usage:
  mcp-assert audit    --server <cmd> [--output <dir>] [--docker <image>] [--include-writes]
  mcp-assert fuzz     --server <cmd> [--runs 50] [--seed 42] [--tool <name>]
  mcp-assert init     [dir] [--server <cmd>] [--fixture <dir>]
  mcp-assert run      --suite <dir> [--server <cmd>] [--fixture <dir>] [--trials N] [--fix]
  mcp-assert matrix   --suite <dir> --languages <lang:server,...>
  mcp-assert ci       --suite <dir> [--server <cmd>] [--threshold N] [--fail-on-regression] [--fix]
  mcp-assert coverage --suite <dir> --server <cmd> [--coverage-json <path>]
  mcp-assert generate --server <cmd> --output <dir> [--fixture <dir>]
  mcp-assert snapshot --suite <dir> [--update] [--server <cmd>] [--fixture <dir>]
  mcp-assert watch    --suite <dir> [--server <cmd>] [--interval <duration>]
  mcp-assert intercept --server <cmd> --trajectory <yaml>
  mcp-assert lint      --server <cmd> [--json] [--threshold N] [--call-tools]

Commands:
  lint      Static schema analysis: check descriptions, types, examples, response sizes
  audit     Zero-config quality audit: connect, discover tools, test each one
  fuzz      Adversarial input testing: throw bad inputs at every tool, find crashes
  init      Scaffold a template, or generate a complete suite with --server
  run       Run assertions against an MCP server
  matrix    Run assertions across multiple language servers
  ci        Run assertions with CI-specific output and exit codes
  coverage  Show which server tools have assertions and which don't
  generate  Auto-generate stub assertions from a server's tools/list
  snapshot  Capture/compare tool response snapshots (like jest --updateSnapshot)
  watch     Rerun assertions when YAML files change (polling, no dependencies)
  intercept Proxy stdio, capture tool calls, validate trajectory assertions

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
  --fix                  Scan nearby positions when position-sensitive assertions fail

`)
}
