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

	// runAndCollect loads and runs all assertions, returning the raw results.
	// Unlike Run(), it does not print results or exit on failure.
	runAndCollect := func() []assertion.Result {
		suite, err := assertion.LoadSuite(*suiteDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading suite: %v\n", err)
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

	// Snapshot mtimes of all YAML files in the suite directory.
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
	// status change notifications and unified diffs for flipped assertions.
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

	// Track results from the previous iteration for diff computation.
	prevResults := make(map[string]assertion.Result)

	// Initial run.
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

	// Poll loop.
	for {
		time.Sleep(*interval)

		currentMtimes, err := snapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
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

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}
