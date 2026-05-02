// snapshot.go implements the snapshot command: capture tool responses as golden
// files for regression detection, similar to Jest's --updateSnapshot.
//
// Two modes:
//   - Verify (default): compare current responses against saved snapshots.
//     Fails if any response changed (checksum mismatch) or is missing.
//   - Update (--update): overwrite snapshots with current responses.
//     Reports NEW (first capture), UPDATE (changed), or matched (unchanged).
//
// Snapshots are stored as a JSON file (<suite-dir>/.snapshots.json) alongside
// the assertion YAML files. Each snapshot records the tool name, response text,
// isError flag, and a SHA256 checksum for fast comparison.
//
// The flow:
//   1. Load the assertion suite and existing snapshots.
//   2. For each assertion: connect to server, run setup steps, call tool, capture text.
//   3. Compare captured text against saved snapshot (verify) or save it (update).
//   4. Report results and optionally write updated snapshots to disk.
package runner

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// SnapshotOpts configures the snapshot operation.
type SnapshotOpts struct {
	SuiteDir string
	Fixture  string
	Server   string        // override server command for all assertions
	Docker   string        // Docker image for isolation
	Update   bool          // true = overwrite snapshots, false = verify
	Timeout  time.Duration // per-assertion timeout
}

// SnapshotResult holds the outcome of a snapshot operation.
type SnapshotResult struct {
	Matched int // snapshots that match current output
	Changed int // snapshots that differ from current output
	New     int // assertions with no prior snapshot
}

// SnapshotCore runs the snapshot logic and returns a result summary.
// Used by both the Snapshot CLI command and Init (which captures baselines
// after generating stubs).
func SnapshotCore(opts SnapshotOpts) (*SnapshotResult, error) {
	suite, err := assertion.LoadSuite(opts.SuiteDir)
	if err != nil {
		return nil, err
	}

	// Load existing snapshots from <suite-dir>/.snapshots.json.
	sf, err := report.LoadSnapshots(opts.SuiteDir)
	if err != nil {
		return nil, err
	}

	// Index by name for O(1) lookup during comparison.
	savedMap := make(map[string]report.Snapshot)
	for _, s := range sf.Snapshots {
		savedMap[s.Name] = s
	}

	var newSnapshots []report.Snapshot
	matched, changed, newCount := 0, 0, 0
	var failures []string

	total := len(suite.Assertions)
	for i, a := range suite.Assertions {
		if opts.Server != "" {
			applyServerOverride(&a, opts.Server)
		}

		report.ProgressLine(i+1, total, a.Name)

		// Run the tool call and capture the response text.
		isoFixture, cleanup := isolateFixture(opts.Fixture, opts.Docker)
		text, isError, err := runAndCapture(a, isoFixture, opts.Timeout, opts.Docker)
		cleanup()
		if err != nil {
			fmt.Printf("  ERROR  %s: %v\n", a.Name, err)
			continue
		}

		snap := report.Snapshot{
			Name:     a.Name,
			Tool:     a.Assert.Tool,
			Text:     text,
			IsError:  isError,
			Checksum: report.Checksum(text),
		}

		saved, exists := savedMap[a.Name]

		if opts.Update {
			// Update mode: overwrite snapshots with current responses.
			if !exists {
				newCount++
				fmt.Printf("  NEW    %s\n", a.Name)
			} else if saved.Checksum != snap.Checksum {
				changed++
				fmt.Printf("  UPDATE %s\n", a.Name)
			} else {
				matched++
			}
			newSnapshots = append(newSnapshots, snap)
		} else {
			// Verify mode: compare current responses against saved snapshots.
			if !exists {
				fmt.Printf("  MISS   %s (no snapshot — run with --update)\n", a.Name)
				failures = append(failures, a.Name)
			} else if err := report.CompareSnapshot(saved, text, isError); err != nil {
				changed++
				fmt.Printf("  DIFF   %s: %v\n", a.Name, err)
				failures = append(failures, a.Name)
			} else {
				matched++
				fmt.Printf("  MATCH  %s\n", a.Name)
			}
			newSnapshots = append(newSnapshots, saved) // preserve existing
		}
	}
	report.ClearProgress()

	// Write updated snapshots to disk.
	if opts.Update {
		sf.Snapshots = newSnapshots
		if err := report.SaveSnapshots(opts.SuiteDir, sf); err != nil {
			return nil, fmt.Errorf("saving snapshots: %w", err)
		}
	}

	result := &SnapshotResult{
		Matched: matched,
		Changed: changed,
		New:     newCount,
	}

	if !opts.Update && len(failures) > 0 {
		return result, fmt.Errorf("%d snapshot(s) failed", len(failures))
	}

	return result, nil
}

// Snapshot is the CLI entry point for the snapshot command.
func Snapshot(args []string) error {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	fixture := fs.String("fixture", "", "Fixture directory (substituted for {{fixture}})")
	server := fs.String("server", "", "Override server command")
	docker := fs.String("docker", "", "Run MCP server inside this Docker image")
	update := fs.Bool("update", false, "Update snapshots with current outputs")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-assertion timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *suiteDir == "" {
		return fmt.Errorf("--suite is required")
	}

	result, err := SnapshotCore(SnapshotOpts{
		SuiteDir: *suiteDir,
		Fixture:  *fixture,
		Server:   *server,
		Docker:   *docker,
		Update:   *update,
		Timeout:  *timeout,
	})
	if err != nil {
		return err
	}

	if *update {
		report.PrintSnapshotSummary(1, result.Matched, result.Changed, result.New)
	} else {
		report.PrintSnapshotSummary(0, result.Matched, result.Changed, result.New)
	}

	return nil
}

// runAndCapture executes a single assertion and returns the raw response text.
// Connects to the server, runs any setup steps, calls the assertion tool, and
// extracts the text content from the result. Does not check expectations;
// that's the caller's job (compare against snapshot or pass to checker).
func runAndCapture(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, _, err := connectAndInitialize(a.Server, connectOpts{
		fixture:     fixture,
		dockerImage: dockerImage,
	})
	if err != nil {
		return "", false, err
	}
	defer mcpClient.Close()

	// Run setup steps (e.g., create a file before testing read_file).
	// Note: setup steps in snapshot mode use fixture-only substitution,
	// not the full substituteAll with captured variables. This means
	// capture/variable interpolation between setup steps is not supported
	// in snapshot mode.
	for _, step := range a.Setup {
		stepArgs := substituteFixture(step.Args, fixture)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		if _, err := mcpClient.CallTool(ctx, req); err != nil {
			return "", false, fmt.Errorf("setup %s: %w", step.Tool, err)
		}
	}

	// Call the assertion tool and capture the response.
	assertArgs := substituteFixture(a.Assert.Args, fixture)
	req := mcp.CallToolRequest{}
	req.Params.Name = a.Assert.Tool
	req.Params.Arguments = assertArgs
	result, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		return "", false, fmt.Errorf("tool %s: %w", a.Assert.Tool, err)
	}

	return extractText(result), result.IsError, nil
}
