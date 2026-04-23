package runner

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/client"
	clienttransport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// Run executes assertions from a suite directory.
func Run(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	fixture := fs.String("fixture", "", "Fixture directory (substituted for {{fixture}})")
	server := fs.String("server", "", "Override server command (e.g. 'agent-lsp go:gopls')")
	docker := fs.String("docker", "", "Run MCP server inside this Docker image")
	trials := fs.Int("trials", 1, "Number of trials per assertion")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-assertion timeout")
	jsonOut := fs.Bool("json", false, "Output results as JSON")
	junitPath := fs.String("junit", "", "Write JUnit XML report to path")
	markdownPath := fs.String("markdown", "", "Write markdown summary to path (or $GITHUB_STEP_SUMMARY)")
	badgePath := fs.String("badge", "", "Write shields.io badge JSON to path")
	baselinePath := fs.String("baseline", "", "Baseline JSON file for regression detection")
	saveBaseline := fs.String("save-baseline", "", "Save current results as baseline to path")
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

	totalAssertions := len(suite.Assertions) * *trials
	current := 0
	var allResults []assertion.Result
	for _, a := range suite.Assertions {
		if *server != "" {
			applyServerOverride(&a, *server)
		}
		for trial := 1; trial <= *trials; trial++ {
			current++
			report.ProgressLine(current, totalAssertions, a.Name)
			r := runAssertion(a, *fixture, *timeout, *docker)
			r.Trial = trial
			allResults = append(allResults, r)
		}
	}
	report.ClearProgress()

	if *jsonOut {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))
	} else {
		report.PrintResults(allResults)
		if *trials > 1 {
			report.PrintReliability(allResults)
		}
	}

	writeReports(allResults, *junitPath, *markdownPath, *badgePath)

	if *saveBaseline != "" {
		if err := report.WriteBaseline(allResults, *saveBaseline); err != nil {
			fmt.Fprintf(os.Stderr, "warning: save-baseline: %v\n", err)
		}
	}

	// Regression detection.
	if *baselinePath != "" {
		baseline, err := report.LoadBaseline(*baselinePath)
		if err != nil {
			return fmt.Errorf("loading baseline: %w", err)
		}
		regressions := report.DetectRegressions(baseline, allResults)
		report.PrintRegressions(regressions)
		if len(regressions) > 0 {
			return fmt.Errorf("%d regression(s) detected", len(regressions))
		}
	}

	for _, r := range allResults {
		if r.Status == assertion.StatusFail {
			return fmt.Errorf("%d assertion(s) failed", countFails(allResults))
		}
	}
	return nil
}

// Matrix runs assertions across multiple language server configurations.
func Matrix(args []string) error {
	fs := flag.NewFlagSet("matrix", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	languages := fs.String("languages", "", "Comma-separated lang:server pairs")
	fixture := fs.String("fixture", "", "Fixture directory")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-assertion timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *suiteDir == "" || *languages == "" {
		return fmt.Errorf("--suite and --languages are required")
	}

	suite, err := assertion.LoadSuite(*suiteDir)
	if err != nil {
		return err
	}

	var allResults []assertion.Result
	for _, langSpec := range strings.Split(*languages, ",") {
		parts := strings.SplitN(langSpec, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid language spec %q (expected lang:server)", langSpec)
		}
		lang, server := parts[0], parts[1]

		for _, a := range suite.Assertions {
			// Override server config for matrix mode.
			a.Server.Command = "agent-lsp"
			a.Server.Args = []string{lang + ":" + server}
			r := runAssertion(a, *fixture, *timeout, "")
			r.Language = lang
			allResults = append(allResults, r)
		}
	}

	report.PrintMatrix(allResults)
	return nil
}

// CI runs assertions with CI-specific exit codes.
func CI(args []string) error {
	fs := flag.NewFlagSet("ci", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	fixture := fs.String("fixture", "", "Fixture directory")
	server := fs.String("server", "", "Override server command (e.g. 'agent-lsp go:gopls')")
	docker := fs.String("docker", "", "Run MCP server inside this Docker image")
	threshold := fs.Int("threshold", 100, "Minimum pass percentage")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-assertion timeout")
	junitPath := fs.String("junit", "", "Write JUnit XML report to path")
	markdownPath := fs.String("markdown", "", "Write markdown summary to path (or $GITHUB_STEP_SUMMARY)")
	badgePath := fs.String("badge", "", "Write shields.io badge JSON to path")
	baselinePath := fs.String("baseline", "", "Baseline JSON file for regression detection")
	saveBaseline := fs.String("save-baseline", "", "Save current results as baseline to path")
	failOnRegression := fs.Bool("fail-on-regression", false, "Exit 1 if any previously-passing assertion regresses (requires --baseline)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *suiteDir == "" {
		return fmt.Errorf("--suite is required")
	}
	if *failOnRegression && *baselinePath == "" {
		return fmt.Errorf("--fail-on-regression requires --baseline")
	}

	suite, err := assertion.LoadSuite(*suiteDir)
	if err != nil {
		return err
	}

	var allResults []assertion.Result
	for i, a := range suite.Assertions {
		if *server != "" {
			applyServerOverride(&a, *server)
		}
		report.ProgressLine(i+1, len(suite.Assertions), a.Name)
		r := runAssertion(a, *fixture, *timeout, *docker)
		allResults = append(allResults, r)
	}
	report.ClearProgress()

	report.PrintResults(allResults)

	// Auto-write GitHub Step Summary when in CI.
	mdPath := *markdownPath
	if mdPath == "" && os.Getenv("GITHUB_STEP_SUMMARY") != "" {
		mdPath = os.Getenv("GITHUB_STEP_SUMMARY")
	}
	writeReports(allResults, *junitPath, mdPath, *badgePath)

	if *saveBaseline != "" {
		if err := report.WriteBaseline(allResults, *saveBaseline); err != nil {
			fmt.Fprintf(os.Stderr, "warning: save-baseline: %v\n", err)
		}
	}

	// Regression detection.
	if *baselinePath != "" {
		baseline, err := report.LoadBaseline(*baselinePath)
		if err != nil {
			return fmt.Errorf("loading baseline: %w", err)
		}
		regressions := report.DetectRegressions(baseline, allResults)
		report.PrintRegressions(regressions)
		if *failOnRegression && len(regressions) > 0 {
			return fmt.Errorf("%d regression(s) detected", len(regressions))
		}
	}

	passed := countPasses(allResults)
	total := len(allResults)
	pct := 0
	if total > 0 {
		pct = (passed * 100) / total
	}

	if pct < *threshold {
		return fmt.Errorf("pass rate %d%% is below threshold %d%%", pct, *threshold)
	}
	return nil
}

func runAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string) assertion.Result {
	start := time.Now()

	// Trajectory assertions check a tool call sequence without calling the server.
	if len(a.Trajectory) > 0 {
		return runTrajectoryAssertion(a, fixture, start)
	}

	// Resource assertions call resources/list or resources/read instead of tools/call.
	if a.AssertResources != nil {
		return runResourceAssertion(a, fixture, timeout, dockerImage, start)
	}

	// Prompt assertions call prompts/list or prompts/get instead of tools/call.
	if a.AssertPrompts != nil {
		return runPromptAssertion(a, fixture, timeout, dockerImage, start)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: time.Since(start),
		}
	}
	defer mcpClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {  //nolint:errcheck
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("MCP initialize failed: %v", err),
			Duration: time.Since(start),
		}
	}

	// Register progress notification handler before setup so it's active for the full lifetime.
	var progressCount int32
	if a.Assert.CaptureProgress {
		mcpClient.OnNotification(func(n mcp.JSONRPCNotification) {
			if n.Method == "notifications/progress" {
				atomic.AddInt32(&progressCount, 1)
			}
		})
	}

	// Run setup steps with variable capture.
	captured := make(map[string]string) // variable_name -> captured value
	for _, step := range a.Setup {
		stepArgs := substituteAll(step.Args, fixture, captured)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		stepResult, err := mcpClient.CallTool(ctx, req)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("setup step %s failed: %v", step.Tool, err),
				Duration: time.Since(start),
			}
		}

		// Capture variables from the response.
		if len(step.Capture) > 0 && stepResult != nil {
			responseText := extractText(stepResult)
			for varName, jsonPath := range step.Capture {
				val, err := extractJSONPath(responseText, jsonPath)
				if err != nil {
					return assertion.Result{
						Name:     a.Name,
						Status:   assertion.StatusFail,
						Detail:   fmt.Sprintf("setup step %s: capture %q from %q failed: %v", step.Tool, varName, jsonPath, err),
						Duration: time.Since(start),
					}
				}
				captured[varName] = val
			}
		}
	}

	// Snapshot files for file_unchanged assertions.
	snapshots := make(map[string]string)
	for _, path := range a.Assert.Expect.FileUnchanged {
		p := strings.ReplaceAll(path, "{{fixture}}", fixture)
		if data, err := os.ReadFile(p); err == nil {
			snapshots[p] = string(data)
		}
	}

	// Run the assertion tool call.
	assertArgs := substituteAll(a.Assert.Args, fixture, captured)
	req := mcp.CallToolRequest{}
	req.Params.Name = a.Assert.Tool
	req.Params.Arguments = assertArgs
	result, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("tool call %s failed: %v", a.Assert.Tool, err),
			Duration: time.Since(start),
		}
	}

	// Extract text from result.
	resultText := extractText(result)
	isError := result.IsError

	// Check assertions (with file snapshots for file_unchanged).
	if err := assertion.CheckWithSnapshots(a.Assert.Expect, resultText, isError, snapshots); err != nil {
		detail := err.Error()
		if isError && resultText != "" {
			detail += "\n      server response: " + resultText
		}
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   detail,
			Duration: time.Since(start),
		}
	}

	// Check progress notification count if capture_progress was requested.
	if a.Assert.CaptureProgress {
		if err := assertion.CheckProgress(a.Assert.Expect, int(atomic.LoadInt32(&progressCount))); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   err.Error(),
				Duration: time.Since(start),
			}
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// createMCPClient creates the appropriate MCP client based on the server config's
// transport type. For stdio (default), it launches a subprocess. For sse/http, it
// connects to the specified URL. Docker isolation is only supported with stdio.
func createMCPClient(server assertion.ServerConfig, fixture string, dockerImage string) (client.MCPClient, error) {
	transport := strings.ToLower(server.Transport)

	switch transport {
	case "sse":
		if server.URL == "" {
			return nil, fmt.Errorf("transport %q requires a url field", transport)
		}
		return client.NewSSEMCPClient(server.URL)
	case "http":
		if server.URL == "" {
			return nil, fmt.Errorf("transport %q requires a url field", transport)
		}
		return client.NewStreamableHttpClient(server.URL)
	case "stdio", "":
		// Default: launch server as a subprocess via stdio.
		serverCmd := server.Command
		serverArgs := make([]string, len(server.Args))
		copy(serverArgs, server.Args)

		if fixture != "" {
			for i, arg := range serverArgs {
				serverArgs[i] = strings.ReplaceAll(arg, "{{fixture}}", fixture)
			}
		}

		var envSlice []string
		for k, v := range server.Env {
			envSlice = append(envSlice, k+"="+v)
		}

		// Docker isolation: wrap server command in docker run -i.
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
			envSlice = nil // env is passed via -e flags
		}

		// If client capabilities are set, use NewClient directly to pass options.
		caps := server.ClientCapabilities
		if len(caps.Roots) > 0 || caps.Sampling != nil || len(caps.Elicitation) > 0 {
			return createStdioClientWithCapabilities(serverCmd, envSlice, serverArgs, fixture, caps)
		}

		return client.NewStdioMCPClient(serverCmd, envSlice, serverArgs...)
	default:
		return nil, fmt.Errorf("unknown transport %q (expected stdio, sse, or http)", transport)
	}
}

// createStdioClientWithCapabilities creates a stdio client with mock client capabilities.
func createStdioClientWithCapabilities(
	command string,
	env []string,
	args []string,
	fixture string,
	caps assertion.ClientCapabilities,
) (client.MCPClient, error) {
	stdioTransport := clienttransport.NewStdioWithOptions(command, env, args)
	if err := stdioTransport.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start stdio transport: %w", err)
	}

	var opts []client.ClientOption

	// Roots: respond to roots/list with configured paths.
	if len(caps.Roots) > 0 {
		roots := make([]mcp.Root, 0, len(caps.Roots))
		for _, path := range caps.Roots {
			path = strings.ReplaceAll(path, "{{fixture}}", fixture)
			roots = append(roots, mcp.Root{
				URI:  "file://" + path,
				Name: filepath.Base(path),
			})
		}
		opts = append(opts, client.WithRootsHandler(&staticRootsHandler{roots: roots}))
	}

	// Sampling: respond to sampling/createMessage with a mock LLM response.
	if caps.Sampling != nil {
		text := caps.Sampling.Text
		model := caps.Sampling.Model
		if model == "" {
			model = "mock"
		}
		stopReason := caps.Sampling.StopReason
		if stopReason == "" {
			stopReason = "end_turn"
		}
		opts = append(opts, client.WithSamplingHandler(&staticSamplingHandler{
			text:       text,
			model:      model,
			stopReason: stopReason,
		}))
	}

	// Elicitation: respond with preset values.
	if len(caps.Elicitation) > 0 {
		opts = append(opts, client.WithElicitationHandler(&staticElicitationHandler{values: caps.Elicitation}))
	}

	c := client.NewClient(stdioTransport, opts...)
	// Start the client to register bidirectional request handlers (roots, sampling, elicitation).
	if err := c.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}
	return c, nil
}

// staticRootsHandler returns a fixed list of roots.
type staticRootsHandler struct {
	roots []mcp.Root
}

func (h *staticRootsHandler) ListRoots(_ context.Context, _ mcp.ListRootsRequest) (*mcp.ListRootsResult, error) {
	return &mcp.ListRootsResult{Roots: h.roots}, nil
}

// staticSamplingHandler returns a fixed mock LLM response.
type staticSamplingHandler struct {
	text       string
	model      string
	stopReason string
}

func (h *staticSamplingHandler) CreateMessage(_ context.Context, _ mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	return &mcp.CreateMessageResult{
		SamplingMessage: mcp.SamplingMessage{
			Role:    mcp.RoleAssistant,
			Content: mcp.TextContent{Type: "text", Text: h.text},
		},
		Model:      h.model,
		StopReason: h.stopReason,
	}, nil
}

// staticElicitationHandler returns preset values for server-initiated prompts.
type staticElicitationHandler struct {
	values map[string]any
}

func (h *staticElicitationHandler) Elicit(_ context.Context, _ mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
	var content any
	if c, ok := h.values["content"]; ok {
		content = c
	} else {
		content = h.values
	}
	return &mcp.ElicitationResult{
		ElicitationResponse: mcp.ElicitationResponse{
			Action:  mcp.ElicitationResponseActionAccept,
			Content: content,
		},
	}, nil
}

func extractText(result *mcp.CallToolResult) string {
	var parts []string
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// substituteAll replaces {{fixture}} and any captured variables in args.
func substituteAll(args map[string]any, fixture string, captured map[string]string) map[string]any {
	out := make(map[string]any, len(args))
	for k, v := range args {
		out[k] = substituteValue(v, fixture, captured)
	}
	return out
}

// substituteFixture replaces {{fixture}} only (backward compat for callers without captures).
func substituteFixture(args map[string]any, fixture string) map[string]any {
	return substituteAll(args, fixture, nil)
}

func substituteValue(v any, fixture string, captured map[string]string) any {
	switch val := v.(type) {
	case string:
		s := val
		if fixture != "" {
			s = strings.ReplaceAll(s, "{{fixture}}", fixture)
		}
		for name, value := range captured {
			s = strings.ReplaceAll(s, "{{"+name+"}}", value)
		}
		return s
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = substituteValue(item, fixture, captured)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = substituteValue(item, fixture, captured)
		}
		return out
	default:
		return v
	}
}

// extractJSONPath extracts a value from JSON text using a simple dot-notation path.
// Reuses the jsonPathLookup logic from the assertion checker.
func extractJSONPath(jsonText, path string) (string, error) {
	path = strings.TrimPrefix(path, "$.")
	if path == "" || path == "$" {
		return jsonText, nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return "", fmt.Errorf("response is not valid JSON: %w", err)
	}

	parts := strings.Split(path, ".")
	current := parsed
	for _, part := range parts {
		// Handle array index: "field[0]"
		if idx := strings.Index(part, "["); idx >= 0 {
			field := part[:idx]
			indexStr := strings.TrimSuffix(part[idx+1:], "]")

			obj, ok := current.(map[string]any)
			if !ok {
				return "", fmt.Errorf("expected object at %q, got %T", field, current)
			}
			arr, ok := obj[field].([]any)
			if !ok {
				return "", fmt.Errorf("expected array at %q", field)
			}
			var i int
			if _, err := fmt.Sscanf(indexStr, "%d", &i); err != nil {
				return "", fmt.Errorf("invalid array index %q", indexStr)
			}
			if i < 0 || i >= len(arr) {
				return "", fmt.Errorf("index %d out of range (len=%d)", i, len(arr))
			}
			current = arr[i]
			continue
		}

		obj, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("expected object at %q, got %T", part, current)
		}
		v, ok := obj[part]
		if !ok {
			return "", fmt.Errorf("field %q not found", part)
		}
		current = v
	}

	// Convert to string.
	switch val := current.(type) {
	case string:
		return val, nil
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val)), nil
		}
		return fmt.Sprintf("%g", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	default:
		// For objects/arrays, marshal back to JSON.
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val), nil
		}
		return string(data), nil
	}
}

func countFails(results []assertion.Result) int {
	n := 0
	for _, r := range results {
		if r.Status == assertion.StatusFail {
			n++
		}
	}
	return n
}

func countPasses(results []assertion.Result) int {
	n := 0
	for _, r := range results {
		if r.Status == assertion.StatusPass {
			n++
		}
	}
	return n
}

// writeReports writes optional structured report files. Errors are printed
// to stderr but do not fail the run — reporting is best-effort.
func writeReports(results []assertion.Result, junitPath, markdownPath, badgePath string) {
	if junitPath != "" {
		if err := report.WriteJUnit(results, junitPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: junit: %v\n", err)
		}
	}
	if markdownPath != "" {
		if err := report.WriteMarkdownSummary(results, markdownPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: markdown: %v\n", err)
		}
	}
	if badgePath != "" {
		if err := report.WriteBadge(results, badgePath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: badge: %v\n", err)
		}
	}
}

// runResourceAssertion tests MCP resources (resources/list or resources/read).
func runResourceAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	// Validate up front — avoids starting the server for a malformed assertion.
	rb := a.AssertResources
	if rb.List == nil && rb.Read == "" {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   "assert_resources requires either 'list' or 'read'",
			Duration: time.Since(start),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: time.Since(start),
		}
	}
	defer mcpClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("MCP initialize failed: %v", err),
			Duration: time.Since(start),
		}
	}

	// Run setup steps.
	captured := make(map[string]string)
	for _, step := range a.Setup {
		stepArgs := substituteAll(step.Args, fixture, captured)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		if _, err := mcpClient.CallTool(ctx, req); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("setup step %s failed: %v", step.Tool, err),
				Duration: time.Since(start),
			}
		}
	}

	var resultText string
	var isError bool

	if rb.List != nil {
		// resources/list
		listReq := mcp.ListResourcesRequest{}
		if rb.List.Cursor != "" {
			listReq.Params.Cursor = mcp.Cursor(strings.ReplaceAll(rb.List.Cursor, "{{fixture}}", fixture))
		}
		result, err := mcpClient.ListResources(ctx, listReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("resources/list failed: %v", err),
				Duration: time.Since(start),
			}
		}
		data, _ := json.Marshal(result)
		resultText = string(data)
	} else if rb.Read != "" {
		// resources/read
		uri := strings.ReplaceAll(rb.Read, "{{fixture}}", fixture)
		readReq := mcp.ReadResourceRequest{}
		readReq.Params.URI = uri
		result, err := mcpClient.ReadResource(ctx, readReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("resources/read failed for %s: %v", uri, err),
				Duration: time.Since(start),
			}
		}
		// Combine all content items into result text.
		var parts []string
		for _, c := range result.Contents {
			switch v := c.(type) {
			case mcp.TextResourceContents:
				parts = append(parts, v.Text)
			case mcp.BlobResourceContents:
				parts = append(parts, fmt.Sprintf("<blob mimeType=%q len=%d>", v.MIMEType, len(v.Blob)))
			default:
				data, _ := json.Marshal(v)
				parts = append(parts, string(data))
			}
		}
		resultText = strings.Join(parts, "\n")
	} else {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   "assert_resources requires either 'list' or 'read'",
			Duration: time.Since(start),
		}
	}

	if err := assertion.Check(rb.Expect, resultText, isError); err != nil {
		detail := err.Error()
		if isError && resultText != "" {
			detail += "\n      server response: " + resultText
		}
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   detail,
			Duration: time.Since(start),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// runPromptAssertion tests MCP prompts (prompts/list or prompts/get).
func runPromptAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	// Validate up front — avoids starting the server for a malformed assertion.
	pb := a.AssertPrompts
	if pb.List == nil && pb.Get == nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   "assert_prompts requires either 'list' or 'get'",
			Duration: time.Since(start),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: time.Since(start),
		}
	}
	defer mcpClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("MCP initialize failed: %v", err),
			Duration: time.Since(start),
		}
	}

	// Run setup steps.
	captured := make(map[string]string)
	for _, step := range a.Setup {
		stepArgs := substituteAll(step.Args, fixture, captured)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		if _, err := mcpClient.CallTool(ctx, req); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("setup step %s failed: %v", step.Tool, err),
				Duration: time.Since(start),
			}
		}
	}

	var resultText string

	if pb.List != nil {
		// prompts/list
		listReq := mcp.ListPromptsRequest{}
		if pb.List.Cursor != "" {
			listReq.Params.Cursor = mcp.Cursor(strings.ReplaceAll(pb.List.Cursor, "{{fixture}}", fixture))
		}
		result, err := mcpClient.ListPrompts(ctx, listReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("prompts/list failed: %v", err),
				Duration: time.Since(start),
			}
		}
		data, _ := json.Marshal(result)
		resultText = string(data)
	} else if pb.Get != nil {
		// prompts/get
		name := strings.ReplaceAll(pb.Get.Name, "{{fixture}}", fixture)
		for k, v := range captured {
			name = strings.ReplaceAll(name, "{{"+k+"}}", v)
		}
		// Substitute captured variables in arguments too.
		args := make(map[string]string, len(pb.Get.Arguments))
		for k, v := range pb.Get.Arguments {
			v = strings.ReplaceAll(v, "{{fixture}}", fixture)
			for varName, varVal := range captured {
				v = strings.ReplaceAll(v, "{{"+varName+"}}", varVal)
			}
			args[k] = v
		}
		getReq := mcp.GetPromptRequest{}
		getReq.Params.Name = name
		getReq.Params.Arguments = args
		result, err := mcpClient.GetPrompt(ctx, getReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("prompts/get failed for %q: %v", name, err),
				Duration: time.Since(start),
			}
		}
		// Build result text from messages.
		var parts []string
		if result.Description != "" {
			parts = append(parts, result.Description)
		}
		for _, msg := range result.Messages {
			switch c := msg.Content.(type) {
			case mcp.TextContent:
				parts = append(parts, c.Text)
			default:
				data, _ := json.Marshal(msg.Content)
				parts = append(parts, string(data))
			}
		}
		resultText = strings.Join(parts, "\n")
	}

	if err := assertion.Check(pb.Expect, resultText, false); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   err.Error(),
			Duration: time.Since(start),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// runTrajectoryAssertion checks a tool call sequence without starting an MCP server.
// The trace comes from the assertion's inline Trace field or an AuditLog file.
func runTrajectoryAssertion(a assertion.Assertion, fixture string, start time.Time) assertion.Result {
	var trace []assertion.TraceEntry

	if a.AuditLog != "" {
		// Load trace from agent-lsp JSONL audit log.
		path := strings.ReplaceAll(a.AuditLog, "{{fixture}}", fixture)
		loaded, err := assertion.LoadAuditLog(path)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("audit_log: %v", err),
				Duration: time.Since(start),
			}
		}
		trace = loaded
	} else {
		// Use inline trace from YAML.
		trace = a.Trace
	}

	if err := assertion.CheckTrajectory(a.Trajectory, trace); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   err.Error(),
			Duration: time.Since(start),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// applyServerOverride parses a "--server" string like "agent-lsp go:gopls"
// and replaces the assertion's server config.
func applyServerOverride(a *assertion.Assertion, serverSpec string) {
	parts := strings.Fields(serverSpec)
	if len(parts) == 0 {
		return
	}
	a.Server.Command = parts[0]
	a.Server.Args = parts[1:]
}
