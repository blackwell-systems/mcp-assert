// Package runner
// commands.go contains the top-level CLI commands: Run, Matrix, and CI.
//
// Each command follows the same pattern:
//  1. Parse flags with a dedicated FlagSet
//  2. Load the assertion suite from YAML
//  3. Execute assertions via runSuite (shared lifecycle)
//  4. Collect results and produce output (console, JSON, JUnit, badge, etc.)
//  5. Optionally detect regressions against a baseline
//  6. Exit non-zero if any assertion failed or pass rate is below threshold
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

// suiteRunOptions configures the shared suite execution lifecycle.
type suiteRunOptions struct {
	fixture      string
	server       string
	docker       string
	timeout      time.Duration
	trials       int
	progress     bool
	reuseServer  bool
}

// serverGroup is a batch of assertions that share the same server config
// and can reuse a single server process. Groups with isolate=true run each
// assertion with its own server (used for destructive tests like restart_lsp).
type serverGroup struct {
	key        string
	assertions []assertion.Assertion
	isolate    bool
}

// statefulTools lists tool names that mutate server state in ways that leak
// across assertions. These must run isolated when server reuse is enabled.
var statefulTools = map[string]bool{
	// Kills/restarts the server process.
	"restart_lsp_server": true,
	// Skill phase state carries across calls.
	"activate_skill":   true,
	"deactivate_skill": true,
	// Session state is per-server, not per-assertion.
	"create_simulation_session": true,
	"simulate_edit":             true,
	"simulate_edit_atomic":      true,
	"simulate_chain":            true,
	"evaluate_session":          true,
	"commit_session":            true,
	"discard_session":           true,
	"destroy_session":           true,
	// Modifies files on disk.
	"apply_edit":    true,
	"rename_symbol": true,
	// Runs external processes (go test, go build).
	"run_tests": true,
	"run_build": true,
	// Executes arbitrary workspace commands on the language server.
	"execute_command": true,
}

// isStateful returns true if the assertion uses a tool that mutates server
// state, making it unsafe for server reuse.
func isStateful(a assertion.Assertion) bool {
	if statefulTools[a.Assert.Tool] {
		return true
	}
	for _, step := range a.Setup {
		if statefulTools[step.Tool] {
			return true
		}
	}
	return false
}

// needsIsolation returns true if the assertion should not share a server.
// This includes destructive tests, offline-only assertions (trajectories),
// and assertions with no server command.
func needsIsolation(a assertion.Assertion) bool {
	if isStateful(a) {
		return true
	}
	// Trajectory assertions don't need a server at all.
	if len(a.Trajectory) > 0 {
		return true
	}
	// No server command means nothing to share.
	if a.Server.Command == "" {
		return true
	}
	return false
}

// groupByServer partitions assertions by ServerKey. Assertions with the same
// key share a server process when --reuse-server is enabled. Assertions that
// need isolation (destructive, trajectory, no server) run individually.
func groupByServer(assertions []assertion.Assertion) []serverGroup {
	keyOrder := make([]string, 0)
	groups := make(map[string]*serverGroup)
	var isolated []serverGroup

	for _, a := range assertions {
		if needsIsolation(a) {
			isolated = append(isolated, serverGroup{
				key:        a.Server.ServerKey() + "-isolated",
				assertions: []assertion.Assertion{a},
				isolate:    true,
			})
			continue
		}

		key := a.Server.ServerKey()
		if _, ok := groups[key]; !ok {
			keyOrder = append(keyOrder, key)
			groups[key] = &serverGroup{key: key}
		}
		groups[key].assertions = append(groups[key].assertions, a)
	}

	result := make([]serverGroup, 0, len(keyOrder)+len(isolated))
	for _, key := range keyOrder {
		result = append(result, *groups[key])
	}
	// Isolated (destructive) assertions run after shared groups.
	result = append(result, isolated...)
	return result
}

// runSuite executes every assertion in the suite, applying server overrides,
// fixture isolation, and optional multi-trial repetition. Returns the collected
// results. This is the shared core of Run and CI.
//
// When opts.reuseServer is true, assertions with the same ServerKey share a
// single server process and fixture copy. This avoids repeated cold starts
// (e.g., gopls workspace indexing) and can reduce suite time by 5-10x.
func runSuite(suite *assertion.Suite, opts suiteRunOptions) []assertion.Result {
	trials := opts.trials
	if trials <= 0 {
		trials = 1
	}

	total := len(suite.Assertions) * trials
	current := 0

	// Apply server overrides before grouping so the keys reflect the actual config.
	for i := range suite.Assertions {
		if opts.server != "" {
			applyServerOverride(&suite.Assertions[i], opts.server)
		}
	}

	if !opts.reuseServer {
		return runSuiteIsolated(suite, opts, &current, total)
	}

	groups := groupByServer(suite.Assertions)
	var results []assertion.Result
	var cleanups []func()

	for _, group := range groups {
		if group.isolate {
			// Destructive tests run with full isolation (own server + fixture).
			for _, a := range group.assertions {
				for trial := 1; trial <= trials; trial++ {
					current++
					if opts.progress {
						report.ProgressLine(current, total, a.Name)
					}
					r := runAssertionWithFixture(a, opts.fixture, opts.docker, opts.timeout)
					r.Trial = trial
					results = append(results, r)
				}
			}
			continue
		}
		groupResults, cleanup := runServerGroup(group, opts, &current, total)
		results = append(results, groupResults...)
		cleanups = append(cleanups, cleanup)
	}

	if opts.progress {
		report.ClearProgress()
	}

	// Close all shared servers AFTER all results are collected. The mcp-go
	// stdio reader goroutine may panic on close (library bug); by deferring
	// cleanup to the very end, the panic only affects exit, not results.
	for _, cleanup := range cleanups {
		cleanup()
	}

	return results
}

// runSuiteIsolated is the original per-assertion isolation strategy.
func runSuiteIsolated(suite *assertion.Suite, opts suiteRunOptions, current *int, total int) []assertion.Result {
	var results []assertion.Result
	for _, a := range suite.Assertions {
		for trial := 1; trial <= opts.trials; trial++ {
			*current++
			if opts.progress {
				report.ProgressLine(*current, total, a.Name)
			}

			r := runAssertionWithFixture(a, opts.fixture, opts.docker, opts.timeout)
			r.Trial = trial
			results = append(results, r)
		}
	}

	if opts.progress {
		report.ClearProgress()
	}

	return results
}

// sharedServer holds a reusable MCP client and its lifecycle controls.
// When the underlying process dies (e.g., restart_lsp_server test), the
// server is automatically re-created for subsequent assertions.
type sharedServer struct {
	ctx       context.Context
	cancel    context.CancelFunc
	client    client.MCPClient
	config    assertion.ServerConfig
	fixture   string
	docker    string
	timeout   time.Duration
	alive     bool
}

// ensureAlive re-creates the server if the previous one died.
func (s *sharedServer) ensureAlive() error {
	if s.alive {
		return nil
	}
	// Clean up the dead client.
	if s.client != nil {
		_ = s.client.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}

	ctx, cancel, mcpClient, err := initializedLongLivedClient(s.config, s.fixture, s.timeout, s.docker)
	if err != nil {
		return fmt.Errorf("server re-creation failed: %w", err)
	}
	s.ctx = ctx
	s.cancel = cancel
	s.client = mcpClient
	s.alive = true
	return nil
}

// close shuts down the server and releases resources.
func (s *sharedServer) close() {
	if s.client != nil {
		_ = s.client.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}
}

// runServerGroup runs all assertions in a group against a shared server process.
// One fixture copy and one server process are created for the entire group.
// Returns results and a cleanup function. The caller must invoke cleanup after
// all results have been processed. This deferred cleanup avoids panics from
// mcp-go's background reader goroutine crashing the process before results
// are collected.
func runServerGroup(group serverGroup, opts suiteRunOptions, current *int, total int) ([]assertion.Result, func()) {
	trials := opts.trials
	if trials <= 0 {
		trials = 1
	}

	// Create one shared fixture for the group.
	isoFixture, fixtureCleanup := isolateFixture(opts.fixture, opts.docker)

	// Create the shared server.
	first := group.assertions[0]
	initTimeout := resolveTimeout(first, opts.timeout)
	ctx, cancel, mcpClient, err := initializedLongLivedClient(first.Server, isoFixture, initTimeout, opts.docker)

	ss := &sharedServer{
		ctx:     ctx,
		cancel:  cancel,
		client:  mcpClient,
		config:  first.Server,
		fixture: isoFixture,
		docker:  opts.docker,
		timeout: initTimeout,
		alive:   err == nil,
	}

	cleanup := func() {
		ss.close()
		fixtureCleanup()
	}

	var results []assertion.Result

	if err != nil {
		for _, a := range group.assertions {
			for trial := 1; trial <= trials; trial++ {
				*current++
				if opts.progress {
					report.ProgressLine(*current, total, a.Name)
				}
				r := failResult(a.Name, time.Now(), fmt.Sprintf("shared server failed to start: %s", err))
				r.Trial = trial
				results = append(results, r)
			}
		}
		return results, cleanup
	}

	for _, a := range group.assertions {
		for trial := 1; trial <= trials; trial++ {
			*current++
			if opts.progress {
				report.ProgressLine(*current, total, a.Name)
			}

			if rerr := ss.ensureAlive(); rerr != nil {
				r := failResult(a.Name, time.Now(), rerr.Error())
				r.Trial = trial
				results = append(results, r)
				continue
			}

			r := runAssertionWithSharedClientSafe(a, isoFixture, opts.timeout, opts.docker, ss)
			r.Trial = trial
			results = append(results, r)
		}
	}

	return results, cleanup
}

// runAssertionWithSharedClientSafe wraps runAssertionWithSharedClient with
// panic recovery. Destructive tests (restart_lsp, etc.) are already excluded
// from shared groups, but this provides defense-in-depth if the server dies
// unexpectedly. The server is marked dead so ensureAlive re-creates it.
func runAssertionWithSharedClientSafe(
	a assertion.Assertion,
	fixture string,
	timeout time.Duration,
	docker string,
	ss *sharedServer,
) (result assertion.Result) {
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			ss.alive = false
			result = failResult(a.Name, start, fmt.Sprintf("server process crashed: %v", r))
		}
	}()

	result = runAssertionWithSharedClient(a, fixture, timeout, docker, ss.ctx, ss.client)
	return result
}

// runAssertionWithFixture wraps runAssertion with fixture isolation and
// deferred cleanup. Using defer ensures cleanup runs even if runAssertion
// panics, which matters for Docker containers that need teardown.
func runAssertionWithFixture(a assertion.Assertion, fixture, docker string, timeout time.Duration) assertion.Result {
	isoFixture, cleanup := isolateFixture(fixture, docker)
	defer cleanup()
	return runAssertion(a, isoFixture, timeout, docker)
}

// collectFixSuggestions scans for position-sensitive assertion failures and
// probes nearby positions to find where the symbol actually is now. Returns
// suggestions with corrected coordinates.
func collectFixSuggestions(
	suite *assertion.Suite,
	results []assertion.Result,
	fixture string,
	timeout time.Duration,
	docker string,
) []FixSuggestion {
	var suggestions []FixSuggestion
	for _, r := range results {
		if r.Status != assertion.StatusFail || !IsPositionError(r.Detail) {
			continue
		}
		for _, a := range suite.Assertions {
			if a.Name != r.Name {
				continue
			}
			if s, err := ScanNearbyPositions(a, fixture, timeout, docker, 3, 5); err == nil && s != nil {
				suggestions = append(suggestions, *s)
			}
			break
		}
	}
	return suggestions
}

// isClientAlive checks if the MCP client is still responsive by issuing a
// lightweight tools/list request. Returns false if the request fails or panics,
// indicating the underlying process has died.
func isClientAlive(ctx context.Context, c client.MCPClient) (alive bool) {
	defer func() {
		if r := recover(); r != nil {
			alive = false
		}
	}()

	probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := c.ListTools(probeCtx, mcp.ListToolsRequest{})
	return err == nil
}

// Run executes assertions from a suite directory. This is the primary command
// for local development: load the suite, run each assertion (optionally
// multiple trials), print results, and exit non-zero on any failure.
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
	fix := fs.Bool("fix", false, "Scan nearby positions when position-sensitive assertions fail")
	reuseServer := fs.Bool("reuse-server", false, "Share server process across assertions with the same config (faster, less isolation)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	if *suiteDir == "" {
		return fmt.Errorf("--suite is required")
	}

	suite, err := assertion.LoadSuite(*suiteDir)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	allResults := runSuite(suite, suiteRunOptions{
		fixture:     *fixture,
		server:      *server,
		docker:      *docker,
		timeout:     *timeout,
		trials:      *trials,
		progress:    true,
		reuseServer: *reuseServer,
	})

	var fixSuggestions []FixSuggestion
	if *fix {
		fixSuggestions = collectFixSuggestions(suite, allResults, *fixture, *timeout, *docker)
	}

	if *jsonOut {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))
	} else {
		report.PrintResults(allResults)
		report.PrintBadgeSnippet(allResults)
		if *trials > 1 {
			report.PrintReliability(allResults)
		}
		PrintFixSuggestions(fixSuggestions)
	}

	writeReports(allResults, *junitPath, *markdownPath, *badgePath)

	if *saveBaseline != "" {
		if err := report.WriteBaseline(allResults, *saveBaseline); err != nil {
			fmt.Fprintf(os.Stderr, "warning: save-baseline: %v\n", err)
		}
	}

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

// Matrix runs the same assertion suite against multiple language servers.
// Each --languages entry is a "lang:server" pair (e.g., "go:gopls,ts:typescript-language-server").
// Results include a Language field so the report can show a cross-language comparison matrix.
func Matrix(args []string) error {
	fs := flag.NewFlagSet("matrix", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	languages := fs.String("languages", "", "Comma-separated lang:server pairs")
	fixture := fs.String("fixture", "", "Fixture directory")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-assertion timeout")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("matrix: %w", err)
	}

	if *suiteDir == "" || *languages == "" {
		return fmt.Errorf("--suite and --languages are required")
	}

	suite, err := assertion.LoadSuite(*suiteDir)
	if err != nil {
		return fmt.Errorf("matrix: %w", err)
	}

	var allResults []assertion.Result
	for _, langSpec := range strings.Split(*languages, ",") {
		parts := strings.SplitN(langSpec, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid language spec %q (expected lang:server)", langSpec)
		}
		lang, server := parts[0], parts[1]

		for _, a := range suite.Assertions {
			a.Server.Command = "agent-lsp"
			a.Server.Args = []string{lang + ":" + server}
			r := runAssertionWithFixture(a, *fixture, "", *timeout)
			r.Language = lang
			allResults = append(allResults, r)
		}
	}

	report.PrintMatrix(allResults)
	return nil
}

// CI runs assertions with CI-specific behavior: threshold-based pass/fail,
// automatic GitHub Step Summary output, baseline regression detection, and
// the --fail-on-regression flag for blocking regressions in pull requests.
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
	fix := fs.Bool("fix", false, "Scan nearby positions when position-sensitive assertions fail")
	reuseServer := fs.Bool("reuse-server", false, "Share server process across assertions with the same config (faster, less isolation)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("ci: %w", err)
	}

	if *suiteDir == "" {
		return fmt.Errorf("--suite is required")
	}
	if *failOnRegression && *baselinePath == "" {
		return fmt.Errorf("--fail-on-regression requires --baseline")
	}

	suite, err := assertion.LoadSuite(*suiteDir)
	if err != nil {
		return fmt.Errorf("ci: %w", err)
	}

	allResults := runSuite(suite, suiteRunOptions{
		fixture:     *fixture,
		server:      *server,
		docker:      *docker,
		timeout:     *timeout,
		trials:      1,
		progress:    true,
		reuseServer: *reuseServer,
	})

	var fixSuggestions []FixSuggestion
	if *fix {
		fixSuggestions = collectFixSuggestions(suite, allResults, *fixture, *timeout, *docker)
	}

	report.PrintResults(allResults)
	report.PrintBadgeSnippet(allResults)
	PrintFixSuggestions(fixSuggestions)

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
	total := 0
	for _, r := range allResults {
		if r.Status != assertion.StatusSkip {
			total++
		}
	}
	pct := 0
	if total > 0 {
		pct = (passed * 100) / total
	}

	if pct < *threshold {
		return fmt.Errorf("pass rate %d%% is below threshold %d%%", pct, *threshold)
	}
	return nil
}
