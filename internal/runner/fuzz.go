package runner

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/fuzz"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// FuzzStatus classifies how a tool responded to a fuzz input.
type FuzzStatus string

const (
	FuzzPassed   FuzzStatus = "passed"   // tool responded (content or isError: true)
	FuzzCrash    FuzzStatus = "crash"    // internal error, server crash, or process died
	FuzzTimeout  FuzzStatus = "timeout"  // tool did not respond within timeout
	FuzzProtocol FuzzStatus = "protocol" // malformed response or leaked stack trace
)

// FuzzInputResult is the outcome of a single fuzz input against a tool.
type FuzzInputResult struct {
	Label      string     `json:"label"`
	Status     FuzzStatus `json:"status"`
	Detail     string     `json:"detail,omitempty"`
	DurationMS int64      `json:"duration_ms"`
}

// FuzzToolResult aggregates all fuzz runs for a single tool.
type FuzzToolResult struct {
	Tool       string            `json:"tool"`
	Runs       int               `json:"runs"`
	Passed     int               `json:"passed"`
	Failures   []FuzzInputResult `json:"failures,omitempty"`
	DurationMS int64             `json:"duration_ms"`
}

// FuzzReport is the full output of a fuzz run.
type FuzzReport struct {
	Server    string           `json:"server"`
	Transport string           `json:"transport"`
	Seed      int64            `json:"seed"`
	RunsPer   int              `json:"runs_per_tool"`
	Tools     []FuzzToolResult `json:"tools"`
	Summary   FuzzSummary      `json:"summary"`
}

// FuzzSummary holds aggregate counts across all tools.
type FuzzSummary struct {
	ToolsTested    int `json:"tools_tested"`
	TotalRuns      int `json:"total_runs"`
	TotalPassed    int `json:"total_passed"`
	TotalFailures  int `json:"total_failures"`
	Crashes        int `json:"crashes"`
	Timeouts       int `json:"timeouts"`
	ProtocolErrors int `json:"protocol_errors"`
}

// Fuzz connects to an MCP server, discovers tools, and throws adversarial
// inputs at each one. Reports which tools crash, hang, or return protocol
// errors under unexpected input.
func Fuzz(args []string) error {
	fs := flag.NewFlagSet("fuzz", flag.ExitOnError)
	var sf serverFlags
	sf.register(fs)
	runs := fs.Int("runs", 50, "Number of fuzz inputs per tool")
	seed := fs.Int64("seed", 0, "Seed for reproducible runs (0 = use current time)")
	toolFilter := fs.String("tool", "", "Fuzz only this tool (default: all tools)")
	junitPath := fs.String("junit", "", "Write JUnit XML report to path")
	markdownPath := fs.String("markdown", "", "Write markdown summary to path (auto-detects $GITHUB_STEP_SUMMARY)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if sf.server == "" {
		return fmt.Errorf("--server is required\n\nUsage: mcp-assert fuzz --server <command-or-url> [--runs 50] [--seed 42]")
	}

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	serverCfg, err := sf.serverConfig()
	if err != nil {
		return err
	}
	transportLower := strings.ToLower(sf.transport)

	if !sf.jsonOut {
		fmt.Fprintf(os.Stderr, "Connecting to server...\n")
	}
	mcpClient, initResult, err := connectAndInitialize(serverCfg, connectOpts{})
	if err != nil {
		return err
	}
	defer mcpClient.Close()

	// Discover tools.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("tools/list failed: %w", err)
	}

	serverName := initResult.ServerInfo.Name
	if serverName == "" {
		serverName = sf.server
	}

	// Filter tools if --tool is set.
	tools := toolsResult.Tools
	if *toolFilter != "" {
		var filtered []mcp.Tool
		for _, t := range tools {
			if t.Name == *toolFilter {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("tool %q not found (server has %d tools)", *toolFilter, len(tools))
		}
		tools = filtered
	}

	if !sf.jsonOut {
		fmt.Fprintf(os.Stderr, "Server: %s\n", serverName)
		fmt.Fprintf(os.Stderr, "Tools:  %d to fuzz\n", len(tools))
		fmt.Fprintf(os.Stderr, "Runs:   %d per tool\n", *runs)
		fmt.Fprintf(os.Stderr, "Seed:   %d\n\n", *seed)
	}

	// Fuzz each tool.
	var toolResults []FuzzToolResult
	summary := FuzzSummary{}

	for ti, tool := range tools {
		toolStart := time.Now()

		// Generate adversarial inputs from the tool's schema.
		// Each tool gets a derived seed so results are independent but reproducible.
		toolSeed := *seed + int64(ti)
		inputs := fuzz.GenerateInputs(tool.InputSchema, *runs, toolSeed)

		var failures []FuzzInputResult
		passed := 0

		for i, input := range inputs {
			if !sf.jsonOut {
				report.ProgressLine(ti*(*runs)+i+1, len(tools)*(*runs), fmt.Sprintf("%s [%d/%d]", tool.Name, i+1, len(inputs)))
			}

			result := fuzzSingleCall(mcpClient, tool.Name, input, sf.timeout)

			if result.Status == FuzzPassed {
				passed++
			} else {
				failures = append(failures, result)
				switch result.Status {
				case FuzzCrash:
					summary.Crashes++
				case FuzzTimeout:
					summary.Timeouts++
				case FuzzProtocol:
					summary.ProtocolErrors++
				}
			}
		}

		toolResult := FuzzToolResult{
			Tool:       tool.Name,
			Runs:       len(inputs),
			Passed:     passed,
			Failures:   failures,
			DurationMS: time.Since(toolStart).Milliseconds(),
		}
		toolResults = append(toolResults, toolResult)

		summary.ToolsTested++
		summary.TotalRuns += len(inputs)
		summary.TotalPassed += passed
		summary.TotalFailures += len(failures)
	}

	if !sf.jsonOut {
		report.ClearProgress()
	}

	fuzzReport := FuzzReport{
		Server:    serverName,
		Transport: effectiveTransport(transportLower),
		Seed:      *seed,
		RunsPer:   *runs,
		Tools:     toolResults,
		Summary:   summary,
	}

	if sf.jsonOut {
		data, err := json.MarshalIndent(fuzzReport, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: JSON marshal failed: %v\n", err)
		}
		fmt.Println(string(data))
	} else {
		report.PrintFuzzHeader(fuzzReport.Server, fuzzReport.Transport, fuzzReport.Seed)
		for _, tr := range toolResults {
			report.PrintFuzzToolResult(tr.Tool, tr.Runs, tr.Passed, convertFuzzFailures(tr.Failures))
		}
		report.PrintFuzzSummary(summary.ToolsTested, summary.TotalRuns, summary.TotalPassed, summary.TotalFailures,
			summary.Crashes, summary.Timeouts, summary.ProtocolErrors)
		fmt.Fprintf(os.Stderr, "\n  Reproduce: mcp-assert fuzz --server %q --seed %d\n\n", sf.server, *seed)
	}

	// Write structured reports.
	if *junitPath != "" || *markdownPath != "" || os.Getenv("GITHUB_STEP_SUMMARY") != "" {
		reportTools := convertFuzzToolReports(toolResults)
		if *junitPath != "" {
			if err := report.WriteFuzzJUnit(reportTools, *junitPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: JUnit: %v\n", err)
			}
		}
		if *markdownPath != "" || os.Getenv("GITHUB_STEP_SUMMARY") != "" {
			path := *markdownPath
			if path == "" {
				path = os.Getenv("GITHUB_STEP_SUMMARY")
			}
			if err := report.WriteFuzzMarkdown(reportTools, summary.ToolsTested, summary.TotalRuns, summary.TotalPassed, summary.TotalFailures, *seed, path); err != nil {
				fmt.Fprintf(os.Stderr, "warning: markdown: %v\n", err)
			}
		}
	}

	if summary.TotalFailures > 0 {
		return fmt.Errorf("%d failures across %d tools", summary.TotalFailures, summary.ToolsTested)
	}
	return nil
}

// fuzzSingleCall sends one adversarial input to a tool and classifies the result.
func fuzzSingleCall(mcpClient interface {
	CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
}, toolName string, input fuzz.InputCase, timeout time.Duration) FuzzInputResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = input.Args

	result, err := mcpClient.CallTool(ctx, req)
	duration := time.Since(start)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return FuzzInputResult{
				Label:      input.Label,
				Status:     FuzzTimeout,
				Detail:     fmt.Sprintf("timed out after %s", timeout),
				DurationMS: duration.Milliseconds(),
			}
		}
		return FuzzInputResult{
			Label:      input.Label,
			Status:     FuzzCrash,
			Detail:     err.Error(),
			DurationMS: duration.Milliseconds(),
		}
	}

	text := extractText(result)

	// Internal errors (stack traces, panics) are protocol violations.
	if result.IsError && isInternalErrorResponse(text) {
		return FuzzInputResult{
			Label:      input.Label,
			Status:     FuzzProtocol,
			Detail:     "internal error: " + truncate(text, 200),
			DurationMS: duration.Milliseconds(),
		}
	}

	// Tool responded with either content or a proper isError. That's fine.
	return FuzzInputResult{
		Label:      input.Label,
		Status:     FuzzPassed,
		DurationMS: duration.Milliseconds(),
	}
}

func convertFuzzToolReports(tools []FuzzToolResult) []report.FuzzToolReport {
	out := make([]report.FuzzToolReport, len(tools))
	for i, t := range tools {
		out[i] = report.FuzzToolReport{
			Tool:       t.Tool,
			Runs:       t.Runs,
			Passed:     t.Passed,
			Failures:   convertFuzzFailures(t.Failures),
			DurationMS: t.DurationMS,
		}
	}
	return out
}

func convertFuzzFailures(failures []FuzzInputResult) []report.FuzzFailure {
	out := make([]report.FuzzFailure, len(failures))
	for i, f := range failures {
		out[i] = report.FuzzFailure{
			Label:  f.Label,
			Status: string(f.Status),
			Detail: f.Detail,
		}
	}
	return out
}
