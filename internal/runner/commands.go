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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
)

// suiteRunOptions configures the shared suite execution lifecycle.
type suiteRunOptions struct {
	fixture  string
	server   string
	docker   string
	timeout  time.Duration
	trials   int
	progress bool
}

// runSuite executes every assertion in the suite, applying server overrides,
// fixture isolation, and optional multi-trial repetition. Returns the collected
// results. This is the shared core of Run and CI.
func runSuite(suite *assertion.Suite, opts suiteRunOptions) []assertion.Result {
	trials := opts.trials
	if trials <= 0 {
		trials = 1
	}

	total := len(suite.Assertions) * trials
	current := 0

	var results []assertion.Result
	for _, a := range suite.Assertions {
		if opts.server != "" {
			applyServerOverride(&a, opts.server)
		}

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

	if opts.progress {
		report.ClearProgress()
	}

	return results
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

	allResults := runSuite(suite, suiteRunOptions{
		fixture:  *fixture,
		server:   *server,
		docker:   *docker,
		timeout:  *timeout,
		trials:   *trials,
		progress: true,
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

	allResults := runSuite(suite, suiteRunOptions{
		fixture:  *fixture,
		server:   *server,
		docker:   *docker,
		timeout:  *timeout,
		trials:   1,
		progress: true,
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
