package runner

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/client"
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

	var allResults []assertion.Result
	for _, a := range suite.Assertions {
		if *server != "" {
			applyServerOverride(&a, *server)
		}
		for trial := 1; trial <= *trials; trial++ {
			r := runAssertion(a, *fixture, *timeout, *docker)
			r.Trial = trial
			allResults = append(allResults, r)
		}
	}

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
	for _, a := range suite.Assertions {
		if *server != "" {
			applyServerOverride(&a, *server)
		}
		r := runAssertion(a, *fixture, *timeout, *docker)
		allResults = append(allResults, r)
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start the MCP server.
	serverCmd := a.Server.Command
	serverArgs := a.Server.Args

	// Substitute {{fixture}} in args.
	if fixture != "" {
		for i, arg := range serverArgs {
			serverArgs[i] = strings.ReplaceAll(arg, "{{fixture}}", fixture)
		}
	}

	// Convert env map to slice.
	var envSlice []string
	for k, v := range a.Server.Env {
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

	mcpClient, err := client.NewStdioMCPClient(serverCmd, envSlice, serverArgs...)
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

	// Run setup steps.
	for _, step := range a.Setup {
		stepArgs := substituteFixture(step.Args, fixture)
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

	// Snapshot files for file_unchanged assertions.
	snapshots := make(map[string]string)
	for _, path := range a.Assert.Expect.FileUnchanged {
		p := strings.ReplaceAll(path, "{{fixture}}", fixture)
		if data, err := os.ReadFile(p); err == nil {
			snapshots[p] = string(data)
		}
	}

	// Run the assertion tool call.
	assertArgs := substituteFixture(a.Assert.Args, fixture)
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

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
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

func substituteFixture(args map[string]any, fixture string) map[string]any {
	if fixture == "" {
		return args
	}
	out := make(map[string]any, len(args))
	for k, v := range args {
		if s, ok := v.(string); ok {
			out[k] = strings.ReplaceAll(s, "{{fixture}}", fixture)
		} else {
			out[k] = v
		}
	}
	return out
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
