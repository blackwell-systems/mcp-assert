// commands.go contains the top-level CLI commands: Run, Matrix, and CI.
//
// Each command follows the same pattern:
//  1. Parse flags with a dedicated FlagSet
//  2. Load the assertion suite from YAML
//  3. Iterate assertions (with optional server override and fixture isolation)
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
			isoFixture, cleanup := isolateFixture(*fixture, *docker)
			r := runAssertion(a, isoFixture, *timeout, *docker)
			cleanup()
			r.Trial = trial
			allResults = append(allResults, r)
		}
	}
	report.ClearProgress()

	// --fix: when position-sensitive assertions fail (e.g., line/column in
	// LSP assertions drifted), scan nearby positions to find where the symbol
	// actually is now and suggest corrected coordinates.
	var fixSuggestions []FixSuggestion
	if *fix {
		for _, r := range allResults {
			if r.Status == assertion.StatusFail && IsPositionError(r.Detail) {
				for _, a := range suite.Assertions {
					if a.Name == r.Name {
						if s, err := ScanNearbyPositions(a, *fixture, *timeout, *docker, 3, 5); err == nil && s != nil {
							fixSuggestions = append(fixSuggestions, *s)
						}
						break
					}
				}
			}
		}
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
			// Override server config for matrix mode.
			a.Server.Command = "agent-lsp"
			a.Server.Args = []string{lang + ":" + server}
			isoFixture, cleanup := isolateFixture(*fixture, "")
			r := runAssertion(a, isoFixture, *timeout, "")
			cleanup()
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

	var allResults []assertion.Result
	for i, a := range suite.Assertions {
		if *server != "" {
			applyServerOverride(&a, *server)
		}
		report.ProgressLine(i+1, len(suite.Assertions), a.Name)
		isoFixture, cleanup := isolateFixture(*fixture, *docker)
		r := runAssertion(a, isoFixture, *timeout, *docker)
		cleanup()
		allResults = append(allResults, r)
	}
	report.ClearProgress()

	// --fix: scan nearby positions for position-sensitive failures.
	var fixSuggestions []FixSuggestion
	if *fix {
		for _, r := range allResults {
			if r.Status == assertion.StatusFail && IsPositionError(r.Detail) {
				for _, a := range suite.Assertions {
					if a.Name == r.Name {
						if s, err := ScanNearbyPositions(a, *fixture, *timeout, *docker, 3, 5); err == nil && s != nil {
							fixSuggestions = append(fixSuggestions, *s)
						}
						break
					}
				}
			}
		}
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
