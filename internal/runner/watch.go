// Package runner
//
// watch.go implements the watch command: poll for YAML file changes and
// rerun assertions automatically, showing diffs when assertion status flips.
//
// The design is intentionally simple: polling via os.ReadDir + mtime comparison.
// No fsnotify, no inotify, no dependencies. This works on every OS including
// Docker containers and network-mounted filesystems where event-based watchers
// are unreliable.
//
// Limitations:
//   - Only watches the top-level suite directory (not subdirectories).
//     LoadSuite recurses one level, so files in subdirectories are executed
//     but changes to them don't trigger a rerun.
//   - No graceful shutdown: Ctrl+C terminates the process. Running MCP server
//     subprocesses from in-progress assertions may be left as orphans.
//   - Stale prevResults: if all YAML files are deleted, the previous results
//     map retains stale entries but no diffs are shown (no current results
//     to compare against).
package runner

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
)

// Watch reruns assertions when YAML files in the suite directory change.
// Runs an initial pass, then enters an infinite poll loop comparing file
// mtimes. On change: clears the screen, reruns all assertions, prints
// results and status-change diffs.
func Watch(args []string) error {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	fixture := fs.String("fixture", "", "Fixture directory (substituted for {{fixture}})")
	server := fs.String("server", "", "Override server command (e.g. 'agent-lsp go:gopls')")
	interval := fs.Duration("interval", 2*time.Second, "Polling interval for file changes")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-assertion timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *suiteDir == "" {
		return fmt.Errorf("--suite is required")
	}

	// runAndCollect loads the suite fresh (picking up file changes) and runs
	// all assertions. Returns raw results without printing or exiting on failure.
	runAndCollect := func() []assertion.Result {
		suite, err := assertion.LoadSuite(*suiteDir)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error loading suite: %v\n", err)
			return nil
		}

		totalAssertions := len(suite.Assertions)
		var results []assertion.Result
		for i, a := range suite.Assertions {
			if *server != "" {
				applyServerOverride(&a, *server)
			}
			report.ProgressLine(i+1, totalAssertions, a.Name)
			isoFixture, cleanup := isolateFixture(*fixture, "")
			r := runAssertion(a, isoFixture, *timeout, "")
			cleanup()
			results = append(results, r)
		}
		report.ClearProgress()
		return results
	}

	// snapshot captures the mtime of every .yaml/.yml file in the suite dir.
	// Used for change detection between poll iterations.
	snapshot := func() (map[string]time.Time, error) {
		mtimes := make(map[string]time.Time)
		entries, err := os.ReadDir(*suiteDir)
		if err != nil {
			return nil, fmt.Errorf("reading suite dir: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := filepath.Ext(e.Name())
			if ext != ".yaml" && ext != ".yml" {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			mtimes[e.Name()] = info.ModTime()
		}
		return mtimes, nil
	}

	// changed compares two mtime snapshots. Returns true if any file was
	// added, removed, or modified.
	changed := func(prev, curr map[string]time.Time) bool {
		if len(prev) != len(curr) {
			return true
		}
		for name, mt := range curr {
			if prev[name] != mt {
				return true
			}
		}
		return false
	}

	// printDiffs compares current results against previous results and prints
	// status change notifications (PASS->FAIL, FAIL->PASS) with unified diffs
	// for assertions that started failing.
	printDiffs := func(prevResults map[string]assertion.Result, currResults []assertion.Result) {
		for _, curr := range currResults {
			prev, seen := prevResults[curr.Name]
			if !seen || prev.Status == curr.Status {
				continue
			}
			fmt.Println(report.FormatStatusChange(curr.Name, prev.Status, curr.Status, curr.Detail))
			if prev.Status == assertion.StatusPass && curr.Status == assertion.StatusFail &&
				curr.Detail != "" {
				diff := report.FormatDiff(curr.Name, prev.Detail, curr.Detail)
				if diff != "" {
					fmt.Print(diff)
				}
			}
		}
	}

	// prevResults tracks the last run's results for diff computation.
	prevResults := make(map[string]assertion.Result)

	// Initial run: execute all assertions before entering the poll loop.
	clearScreen()
	fmt.Printf("[watch] Running assertions from %s (polling every %s)\n\n", *suiteDir, *interval)
	results := runAndCollect()
	report.PrintResults(results)
	for _, r := range results {
		prevResults[r.Name] = r
	}

	lastMtimes, err := snapshot()
	if err != nil {
		return err
	}

	// Poll loop: check for file changes, rerun on change.
	for {
		time.Sleep(*interval)

		currentMtimes, err := snapshot()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}

		if !changed(lastMtimes, currentMtimes) {
			continue
		}

		lastMtimes = currentMtimes
		clearScreen()
		fmt.Printf("[watch] Change detected, rerunning at %s\n\n", time.Now().Format("15:04:05"))
		results = runAndCollect()
		report.PrintResults(results)
		printDiffs(prevResults, results)
		for _, r := range results {
			prevResults[r.Name] = r
		}
	}
}

// clearScreen sends ANSI escape codes to clear the terminal and move the
// cursor to the top-left. Works on all modern terminals (macOS, Linux, WSL).
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}
