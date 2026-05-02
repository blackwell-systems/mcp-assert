package runner

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// AuditToolResult holds the outcome of calling a single tool during an audit.
type AuditToolResult struct {
	Tool        string      `json:"tool"`
	Description string      `json:"description,omitempty"`
	Status      AuditStatus `json:"status"`
	Detail      string      `json:"detail,omitempty"`
	IsError     bool        `json:"is_error"`
	HasContent  bool        `json:"has_content"`
	Duration    time.Duration `json:"-"`
	DurationMS  int64       `json:"duration_ms"`
}

// AuditStatus classifies how a tool responded during the audit.
type AuditStatus string

const (
	AuditHealthy  AuditStatus = "healthy"  // responded, no crash, proper error handling
	AuditCrash    AuditStatus = "crash"    // internal error (-32603) or server crash
	AuditTimeout  AuditStatus = "timeout"  // tool did not respond within timeout
	AuditSkipped  AuditStatus = "skipped"  // destructive tool, skipped by default
)

// AuditReport is the full output of an audit run.
type AuditReport struct {
	Server      string            `json:"server"`
	Transport   string            `json:"transport"`
	ToolCount   int               `json:"tool_count"`
	Results     []AuditToolResult `json:"results"`
	Score       int               `json:"score"`       // percentage of healthy tools
	GeneratedAt string            `json:"generated_at"` // path to generated YAML dir, if any
}

// Audit connects to an MCP server, discovers all tools, calls each one with
// schema-generated inputs, and reports which tools are healthy vs. which crash.
// Optionally generates starter YAML assertion files for CI regression testing.
func Audit(args []string) error {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	serverSpec := fs.String("server", "", "Server command (stdio) or URL (http/sse)")
	transport := fs.String("transport", "stdio", "Transport type: stdio (default), http, sse")
	headersFlag := fs.String("headers", "", "Custom headers as key=value pairs, comma-separated")
	docker := fs.String("docker", "", "Run destructive tools in fresh Docker containers (stdio only)")
	timeout := fs.Duration("timeout", 15*time.Second, "Per-tool timeout")
	output := fs.String("output", "", "Generate assertion YAML files in this directory")
	jsonOut := fs.Bool("json", false, "Output results as JSON")
	includeWrites := fs.Bool("include-writes", false, "Also call destructive/write tools (skipped by default)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *serverSpec == "" {
		return fmt.Errorf("--server is required\n\nUsage: mcp-assert audit --server <command-or-url> [--output <dir>]")
	}

	headers := parseHeadersFlag(*headersFlag)

	// Build server config from flags.
	transportLower := strings.ToLower(*transport)
	var serverCfg assertion.ServerConfig
	if transportLower == "http" || transportLower == "sse" {
		serverCfg = assertion.ServerConfig{
			Transport: transportLower,
			URL:       *serverSpec,
			Headers:   headers,
		}
	} else {
		parts := strings.Fields(*serverSpec)
		if len(parts) == 0 {
			return fmt.Errorf("--server cannot be empty")
		}
		serverCfg = assertion.ServerConfig{
			Command: parts[0],
			Args:    parts[1:],
		}
	}

	// Connect and initialize.
	if !*jsonOut {
		fmt.Fprintf(os.Stderr, "Connecting to server...\n")
	}
	mcpClient, err := createMCPClient(serverCfg, "", "")
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer mcpClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initResult, err := mcpClient.Initialize(ctx, initReq)
	if err != nil {
		if isTransportError(err) {
			return fmt.Errorf("MCP initialize failed: %w\n\nhint: the server exited immediately. Check that any required environment variables (API keys, tokens) are set", err)
		}
		return fmt.Errorf("MCP initialize failed: %w", err)
	}

	// Discover tools.
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("tools/list failed: %w", err)
	}

	serverName := initResult.ServerInfo.Name
	if serverName == "" {
		serverName = *serverSpec
	}

	if !*jsonOut {
		fmt.Fprintf(os.Stderr, "Server: %s\n", serverName)
		fmt.Fprintf(os.Stderr, "Tools:  %d discovered\n", len(toolsResult.Tools))
		fmt.Fprintf(os.Stderr, "Transport: %s\n\n", effectiveTransport(transportLower))
	}

	// Call each tool and collect results.
	var results []AuditToolResult
	for i, tool := range toolsResult.Tools {
		if !*jsonOut {
			report.ProgressLine(i+1, len(toolsResult.Tools), tool.Name)
		}

		destructive := isDestructiveTool(tool)

		if destructive && !*includeWrites && *docker == "" {
			results = append(results, AuditToolResult{
				Tool:        tool.Name,
				Description: tool.Description,
				Status:      AuditSkipped,
				Detail:      "destructive tool (use --include-writes or --docker to test)",
			})
			continue
		}

		// Destructive tools with --docker: spin up a fresh container per tool.
		if destructive && *docker != "" {
			r := auditToolInDocker(serverCfg, tool, *docker, *timeout)
			results = append(results, r)
			continue
		}

		r := auditSingleTool(mcpClient, tool, *timeout)
		results = append(results, r)
	}

	if !*jsonOut {
		report.ClearProgress()
	}

	// Build report.
	auditReport := AuditReport{
		Server:    serverName,
		Transport: effectiveTransport(transportLower),
		ToolCount: len(toolsResult.Tools),
		Results:   results,
		Score:     auditScore(results),
	}

	// Generate YAML files if --output is set.
	if *output != "" {
		if err := generateAuditYAML(toolsResult.Tools, serverCfg, *serverSpec, *output, transportLower, headers); err != nil {
			fmt.Fprintf(os.Stderr, "warning: YAML generation: %v\n", err)
		} else {
			auditReport.GeneratedAt = *output
		}
	}

	// Print report.
	if *jsonOut {
		data, err := json.MarshalIndent(auditReport, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: JSON marshal failed: %v\n", err)
		}
		fmt.Println(string(data))
	} else {
		report.PrintAuditHeader(auditReport.Server, auditReport.Transport, auditReport.Score)

		// Convert to report types.
		reportResults := make([]report.AuditToolResult, len(results))
		healthy, crashes, timeouts, skipped := 0, 0, 0, 0
		for i, r := range results {
			reportResults[i] = report.AuditToolResult{
				Tool:     r.Tool,
				Status:   string(r.Status),
				Detail:   r.Detail,
				Duration: r.Duration,
			}
			switch r.Status {
			case AuditHealthy:
				healthy++
			case AuditCrash:
				crashes++
			case AuditTimeout:
				timeouts++
			case AuditSkipped:
				skipped++
			}
		}

		report.PrintAuditResults(reportResults)
		report.PrintAuditSummary(healthy, crashes, timeouts, skipped, len(results))

		if auditReport.GeneratedAt != "" {
			report.PrintAuditNextSteps(auditReport.GeneratedAt)
		}
	}

	return nil
}

// auditToolInDocker runs a destructive tool in a fresh Docker container.
// It creates a new MCP client per tool call, so each gets an isolated environment.
func auditToolInDocker(serverCfg assertion.ServerConfig, tool mcp.Tool, dockerImage string, timeout time.Duration) AuditToolResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
	defer cancel()

	isoClient, err := createMCPClient(serverCfg, "", dockerImage)
	if err != nil {
		return auditResult(tool, AuditCrash, fmt.Sprintf("docker client: %v", err), false, false, time.Since(start))
	}
	defer isoClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := isoClient.Initialize(ctx, initReq); err != nil {
		return auditResult(tool, AuditCrash, fmt.Sprintf("docker init: %v", err), false, false, time.Since(start))
	}

	return auditSingleTool(isoClient, tool, timeout)
}

// auditSingleTool calls one tool with schema-generated args and classifies the result.
func auditSingleTool(mcpClient interface{ CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) }, tool mcp.Tool, timeout time.Duration) AuditToolResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Generate args from the tool's input schema.
	args := generateArgsFromSchema(tool.InputSchema, "")

	req := mcp.CallToolRequest{}
	req.Params.Name = tool.Name
	req.Params.Arguments = args

	result, err := mcpClient.CallTool(ctx, req)
	duration := time.Since(start)

	if err != nil {
		// Distinguish timeout from crash.
		if ctx.Err() == context.DeadlineExceeded {
			return auditResult(tool, AuditTimeout, fmt.Sprintf("timed out after %s", timeout), false, false, duration)
		}
		return auditResult(tool, AuditCrash, err.Error(), false, false, duration)
	}

	// Tool responded. Check if it used isError properly or crashed.
	text := extractText(result)

	// Internal error in the response body is a crash, not proper error handling.
	if result.IsError && isInternalErrorResponse(text) {
		return auditResult(tool, AuditCrash, "internal error: "+truncate(text, 200), true, text != "", duration)
	}

	return auditResult(tool, AuditHealthy, classifyHealthy(result.IsError, text), result.IsError, text != "", duration)
}

// classifyHealthy returns a human-readable description of a healthy tool response.
func classifyHealthy(isError bool, text string) string {
	if isError {
		return "valid error handling (isError: true)"
	}
	if text == "" {
		return "responds (empty content)"
	}
	return "responds, returns content"
}

// isInternalErrorResponse checks if the response text indicates an internal server error
// rather than a proper user-facing error. These are the bugs we're looking for.
func isInternalErrorResponse(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "internal error") ||
		strings.Contains(lower, "traceback") ||
		strings.Contains(lower, "stack trace") ||
		strings.Contains(lower, "panic:") ||
		strings.Contains(lower, "nullpointerexception") ||
		strings.Contains(lower, "segmentation fault")
}

// auditScore calculates the percentage of tested tools that are healthy.
// Skipped tools are excluded from the denominator.
func auditScore(results []AuditToolResult) int {
	tested, healthy := 0, 0
	for _, r := range results {
		if r.Status == AuditSkipped {
			continue
		}
		tested++
		if r.Status == AuditHealthy {
			healthy++
		}
	}
	if tested == 0 {
		return 100
	}
	return (healthy * 100) / tested
}

// generateAuditYAML creates starter assertion YAML files from audit results.
func generateAuditYAML(tools []mcp.Tool, serverCfg assertion.ServerConfig, serverSpec string, outputDir string, transport string, headers map[string]string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for _, tool := range tools {
		filename := sanitizeFilename(tool.Name) + ".yaml"
		path := filepath.Join(outputDir, filename)

		skip := isDestructiveTool(tool)
		stub := generateStub(tool, serverSpec, "", skip, GenerateOpts{
			Transport: transport,
			Headers:   headers,
		})

		data, err := yaml.Marshal(stub)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %s: %v\n", tool.Name, err)
			continue
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %s: %v\n", tool.Name, err)
			continue
		}
	}
	return nil
}

func effectiveTransport(t string) string {
	if t == "" {
		return "stdio"
	}
	return t
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func auditResult(tool mcp.Tool, status AuditStatus, detail string, isError, hasContent bool, duration time.Duration) AuditToolResult {
	return AuditToolResult{
		Tool:        tool.Name,
		Description: tool.Description,
		Status:      status,
		Detail:      detail,
		IsError:     isError,
		HasContent:  hasContent,
		Duration:    duration,
		DurationMS:  duration.Milliseconds(),
	}
}
