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
	Server   string
	Docker   string
	Update   bool
	Timeout  time.Duration
}

// SnapshotResult holds the outcome of a snapshot operation.
type SnapshotResult struct {
	Matched int
	Changed int
	New     int
}

// SnapshotCore runs the snapshot logic and returns a result summary.
func SnapshotCore(opts SnapshotOpts) (*SnapshotResult, error) {
	suite, err := assertion.LoadSuite(opts.SuiteDir)
	if err != nil {
		return nil, err
	}

	// Load existing snapshots.
	sf, err := report.LoadSnapshots(opts.SuiteDir)
	if err != nil {
		return nil, err
	}

	// Index existing snapshots by name.
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

		// Run the tool call and capture the response.
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

// Snapshot runs assertions and compares/updates snapshots.
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

	if *update {
		report.PrintSnapshotSummary(1, result.Matched, result.Changed, result.New)
	} else {
		report.PrintSnapshotSummary(0, result.Matched, result.Changed, result.New)
	}

	return err
}

// runAndCapture executes a single assertion and returns the raw response text.
func runAndCapture(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return "", false, fmt.Errorf("start server: %w", err)
	}
	defer mcpClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		return "", false, fmt.Errorf("initialize: %w", err)
	}

	// Run setup steps.
	for _, step := range a.Setup {
		stepArgs := substituteFixture(step.Args, fixture)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		if _, err := mcpClient.CallTool(ctx, req); err != nil {
			return "", false, fmt.Errorf("setup %s: %w", step.Tool, err)
		}
	}

	// Call the assertion tool.
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
