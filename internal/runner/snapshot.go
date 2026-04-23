package runner

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

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

	suite, err := assertion.LoadSuite(*suiteDir)
	if err != nil {
		return err
	}

	// Load existing snapshots.
	sf, err := report.LoadSnapshots(*suiteDir)
	if err != nil {
		return err
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
		if *server != "" {
			applyServerOverride(&a, *server)
		}

		report.ProgressLine(i+1, total, a.Name)

		// Run the tool call and capture the response.
		text, isError, err := runAndCapture(a, *fixture, *timeout, *docker)
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

		if *update {
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

	if *update {
		sf.Snapshots = newSnapshots
		if err := report.SaveSnapshots(*suiteDir, sf); err != nil {
			return fmt.Errorf("saving snapshots: %w", err)
		}
		report.PrintSnapshotSummary(1, matched, changed, newCount)
	} else {
		report.PrintSnapshotSummary(0, matched, changed, newCount)
		if len(failures) > 0 {
			return fmt.Errorf("%d snapshot(s) failed", len(failures))
		}
	}

	return nil
}

// runAndCapture executes a single assertion and returns the raw response text.
func runAndCapture(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	serverCmd := a.Server.Command
	serverArgs := make([]string, len(a.Server.Args))
	copy(serverArgs, a.Server.Args)

	if fixture != "" {
		for i, arg := range serverArgs {
			serverArgs[i] = strings.ReplaceAll(arg, "{{fixture}}", fixture)
		}
	}

	var envSlice []string
	for k, v := range a.Server.Env {
		envSlice = append(envSlice, k+"="+v)
	}

	if dockerImage != "" {
		dockerArgs := []string{"run", "--rm", "-i"}
		if fixture != "" {
			dockerArgs = append(dockerArgs, "-v", fixture+":"+fixture)
		}
		for _, e := range envSlice {
			dockerArgs = append(dockerArgs, "-e", e)
		}
		dockerArgs = append(dockerArgs, dockerImage, serverCmd)
		dockerArgs = append(dockerArgs, serverArgs...)
		serverCmd = "docker"
		serverArgs = dockerArgs
		envSlice = nil
	}

	mcpClient, err := client.NewStdioMCPClient(serverCmd, envSlice, serverArgs...)
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
