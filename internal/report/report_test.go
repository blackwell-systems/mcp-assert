package report

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintResults_AllPass(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover test", Status: assertion.StatusPass, Duration: 150 * time.Millisecond},
		{Name: "definition test", Status: assertion.StatusPass, Duration: 200 * time.Millisecond},
	}

	out := captureStdout(func() { PrintResults(results) })

	if !strings.Contains(out, "hover test") {
		t.Error("expected 'hover test' in output")
	}
	if !strings.Contains(out, "2 passed") {
		t.Error("expected '2 passed' in output")
	}
	// When all pass, summary should not mention failures.
	if strings.Contains(out, "failed") {
		t.Error("should not mention 'failed' when all pass")
	}
}

func TestPrintResults_WithFailure(t *testing.T) {
	results := []assertion.Result{
		{Name: "good", Status: assertion.StatusPass, Duration: 100 * time.Millisecond},
		{Name: "bad", Status: assertion.StatusFail, Detail: "expected foo", Duration: 50 * time.Millisecond},
	}

	out := captureStdout(func() { PrintResults(results) })

	if !strings.Contains(out, "FAIL") {
		t.Error("expected FAIL in output")
	}
	if !strings.Contains(out, "expected foo") {
		t.Error("expected failure detail in output")
	}
	if !strings.Contains(out, "1 passed") {
		t.Error("expected '1 passed' in output")
	}
	if !strings.Contains(out, "1 failed") {
		t.Error("expected '1 failed' in output")
	}
}

func TestPrintResults_WithLanguage(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass, Duration: 100 * time.Millisecond, Language: "go"},
	}

	out := captureStdout(func() { PrintResults(results) })

	if !strings.Contains(out, "(go)") {
		t.Error("expected '(go)' language suffix in output")
	}
}

func TestPrintMatrix_Basic(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass, Language: "go"},
		{Name: "hover", Status: assertion.StatusFail, Language: "python"},
		{Name: "definition", Status: assertion.StatusPass, Language: "go"},
		{Name: "definition", Status: assertion.StatusPass, Language: "python"},
	}

	out := captureStdout(func() { PrintMatrix(results) })

	if !strings.Contains(out, "go") {
		t.Error("expected 'go' row in matrix")
	}
	if !strings.Contains(out, "python") {
		t.Error("expected 'python' row in matrix")
	}
	if !strings.Contains(out, "PASS") {
		t.Error("expected PASS in matrix")
	}
	if !strings.Contains(out, "FAIL") {
		t.Error("expected FAIL in matrix")
	}
}

func TestPrintResults_Skip(t *testing.T) {
	results := []assertion.Result{
		{Name: "skipped", Status: assertion.StatusSkip, Duration: 0},
	}

	out := captureStdout(func() { PrintResults(results) })

	if !strings.Contains(out, "SKIP") {
		t.Error("expected SKIP in output")
	}
	if !strings.Contains(out, "1 skipped") {
		t.Error("expected '1 skipped' in output")
	}
}

// --- JUnit XML tests ---

func TestWriteJUnit_AllPass(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass, Duration: 150 * time.Millisecond},
		{Name: "definition", Status: assertion.StatusPass, Duration: 200 * time.Millisecond},
	}
	path := filepath.Join(t.TempDir(), "results.xml")

	if err := WriteJUnit(results, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, `<?xml`) {
		t.Error("expected XML header")
	}
	if !strings.Contains(content, `tests="2"`) {
		t.Error("expected tests=2")
	}
	if !strings.Contains(content, `failures="0"`) {
		t.Error("expected failures=0")
	}
	if !strings.Contains(content, `name="hover"`) {
		t.Error("expected testcase name=hover")
	}

	// Verify it's valid XML.
	var suites junitTestSuites
	if err := xml.Unmarshal(data, &suites); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
	if len(suites.Suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(suites.Suites))
	}
	if len(suites.Suites[0].Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(suites.Suites[0].Cases))
	}
}

func TestWriteJUnit_WithFailure(t *testing.T) {
	results := []assertion.Result{
		{Name: "good", Status: assertion.StatusPass, Duration: 100 * time.Millisecond},
		{Name: "bad", Status: assertion.StatusFail, Detail: "expected foo", Duration: 50 * time.Millisecond},
	}
	path := filepath.Join(t.TempDir(), "results.xml")

	WriteJUnit(results, path)
	data, _ := os.ReadFile(path)

	var suites junitTestSuites
	xml.Unmarshal(data, &suites)

	if suites.Suites[0].Failures != 1 {
		t.Errorf("expected 1 failure, got %d", suites.Suites[0].Failures)
	}
	if suites.Suites[0].Cases[1].Failure == nil {
		t.Error("expected failure element on second case")
	}
	if suites.Suites[0].Cases[1].Failure.Message != "expected foo" {
		t.Errorf("expected failure message 'expected foo', got %q", suites.Suites[0].Cases[1].Failure.Message)
	}
}

func TestWriteJUnit_WithLanguage(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass, Duration: 100 * time.Millisecond, Language: "go"},
	}
	path := filepath.Join(t.TempDir(), "results.xml")

	WriteJUnit(results, path)
	data, _ := os.ReadFile(path)

	if !strings.Contains(string(data), `classname="go"`) {
		t.Error("expected classname=go for language-tagged result")
	}
}

// --- Markdown tests ---

func TestWriteMarkdownSummary_AllPass(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass, Duration: 150 * time.Millisecond},
		{Name: "definition", Status: assertion.StatusPass, Duration: 200 * time.Millisecond},
	}
	path := filepath.Join(t.TempDir(), "summary.md")

	if err := WriteMarkdownSummary(results, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "2/2 passed") {
		t.Error("expected '2/2 passed' in summary")
	}
	if !strings.Contains(content, "| PASS | hover") {
		t.Error("expected PASS hover row")
	}
	if strings.Contains(content, "Failure details") {
		t.Error("should not contain failure details when all pass")
	}
}

func TestWriteMarkdownSummary_WithFailure(t *testing.T) {
	results := []assertion.Result{
		{Name: "good", Status: assertion.StatusPass, Duration: 100 * time.Millisecond},
		{Name: "bad", Status: assertion.StatusFail, Detail: "expected foo", Duration: 50 * time.Millisecond},
	}
	path := filepath.Join(t.TempDir(), "summary.md")

	WriteMarkdownSummary(results, path)
	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "1 passed, 1 failed") {
		t.Error("expected '1 passed, 1 failed' in header")
	}
	if !strings.Contains(content, "Failure details") {
		t.Error("expected failure details section")
	}
	if !strings.Contains(content, "expected foo") {
		t.Error("expected failure detail text")
	}
}

func TestWriteMarkdownSummary_NoPath(t *testing.T) {
	// No path and no $GITHUB_STEP_SUMMARY should error.
	os.Unsetenv("GITHUB_STEP_SUMMARY")
	err := WriteMarkdownSummary(nil, "")
	if err == nil {
		t.Error("expected error when no path given")
	}
}

// --- Badge tests ---

func TestWriteBadge_AllPass(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass},
		{Name: "definition", Status: assertion.StatusPass},
	}
	path := filepath.Join(t.TempDir(), "badge.json")

	if err := WriteBadge(results, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var badge ShieldsEndpoint
	json.Unmarshal(data, &badge)

	if badge.SchemaVersion != 1 {
		t.Errorf("expected schemaVersion 1, got %d", badge.SchemaVersion)
	}
	if badge.Label != "mcp-assert" {
		t.Errorf("expected label 'mcp-assert', got %q", badge.Label)
	}
	if badge.Message != "2/2" {
		t.Errorf("expected message '2/2', got %q", badge.Message)
	}
	if badge.Color != "brightgreen" {
		t.Errorf("expected color 'brightgreen', got %q", badge.Color)
	}
}

func TestWriteBadge_WithFailure(t *testing.T) {
	results := []assertion.Result{
		{Name: "good", Status: assertion.StatusPass},
		{Name: "bad", Status: assertion.StatusFail},
	}
	path := filepath.Join(t.TempDir(), "badge.json")

	WriteBadge(results, path)
	data, _ := os.ReadFile(path)

	var badge ShieldsEndpoint
	json.Unmarshal(data, &badge)

	if badge.Message != "1/2" {
		t.Errorf("expected message '1/2', got %q", badge.Message)
	}
	if badge.Color != "red" {
		t.Errorf("expected color 'red', got %q", badge.Color)
	}
}

// --- Reliability tests ---

func TestComputeReliability_AllPass(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass, Trial: 1},
		{Name: "hover", Status: assertion.StatusPass, Trial: 2},
		{Name: "hover", Status: assertion.StatusPass, Trial: 3},
	}

	stats := ComputeReliability(results)
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if !stats[0].PassAt {
		t.Error("expected pass@k = true")
	}
	if !stats[0].PassUp {
		t.Error("expected pass^k = true")
	}
	if stats[0].Rate != 1.0 {
		t.Errorf("expected rate 1.0, got %f", stats[0].Rate)
	}
}

func TestComputeReliability_Flaky(t *testing.T) {
	results := []assertion.Result{
		{Name: "refs", Status: assertion.StatusPass, Trial: 1},
		{Name: "refs", Status: assertion.StatusFail, Trial: 2},
		{Name: "refs", Status: assertion.StatusPass, Trial: 3},
	}

	stats := ComputeReliability(results)
	if !stats[0].PassAt {
		t.Error("expected pass@k = true (passed at least once)")
	}
	if stats[0].PassUp {
		t.Error("expected pass^k = false (didn't pass every time)")
	}
	if stats[0].Passed != 2 {
		t.Errorf("expected 2 passed, got %d", stats[0].Passed)
	}
}

func TestComputeReliability_NeverPassed(t *testing.T) {
	results := []assertion.Result{
		{Name: "broken", Status: assertion.StatusFail, Trial: 1},
		{Name: "broken", Status: assertion.StatusFail, Trial: 2},
	}

	stats := ComputeReliability(results)
	if stats[0].PassAt {
		t.Error("expected pass@k = false")
	}
	if stats[0].PassUp {
		t.Error("expected pass^k = false")
	}
}

func TestComputeReliability_MultipleAssertions(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass, Trial: 1},
		{Name: "hover", Status: assertion.StatusPass, Trial: 2},
		{Name: "refs", Status: assertion.StatusPass, Trial: 1},
		{Name: "refs", Status: assertion.StatusFail, Trial: 2},
	}

	stats := ComputeReliability(results)
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	// hover: reliable
	if !stats[0].PassUp {
		t.Error("hover should be pass^k = true")
	}
	// refs: flaky
	if stats[1].PassUp {
		t.Error("refs should be pass^k = false")
	}
}

// --- Baseline / Regression tests ---

func TestWriteAndLoadBaseline(t *testing.T) {
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass},
		{Name: "refs", Status: assertion.StatusFail},
	}
	path := filepath.Join(t.TempDir(), "baseline.json")

	if err := WriteBaseline(results, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	baseline, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(baseline.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(baseline.Entries))
	}
}

func TestDetectRegressions_NoRegression(t *testing.T) {
	baseline := &Baseline{Entries: []BaselineEntry{
		{Name: "hover", Status: assertion.StatusPass},
		{Name: "refs", Status: assertion.StatusPass},
	}}
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass},
		{Name: "refs", Status: assertion.StatusPass},
	}

	regressions := DetectRegressions(baseline, results)
	if len(regressions) != 0 {
		t.Errorf("expected 0 regressions, got %d", len(regressions))
	}
}

func TestDetectRegressions_WithRegression(t *testing.T) {
	baseline := &Baseline{Entries: []BaselineEntry{
		{Name: "hover", Status: assertion.StatusPass},
		{Name: "refs", Status: assertion.StatusPass},
	}}
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass},
		{Name: "refs", Status: assertion.StatusFail},
	}

	regressions := DetectRegressions(baseline, results)
	if len(regressions) != 1 {
		t.Fatalf("expected 1 regression, got %d", len(regressions))
	}
	if regressions[0].Name != "refs" {
		t.Errorf("expected regression on 'refs', got %q", regressions[0].Name)
	}
	if regressions[0].NowStatus != assertion.StatusFail {
		t.Errorf("expected now status FAIL, got %s", regressions[0].NowStatus)
	}
}

func TestDetectRegressions_PreviouslyFailing(t *testing.T) {
	// Previously failing -> still failing is NOT a regression.
	baseline := &Baseline{Entries: []BaselineEntry{
		{Name: "broken", Status: assertion.StatusFail},
	}}
	results := []assertion.Result{
		{Name: "broken", Status: assertion.StatusFail},
	}

	regressions := DetectRegressions(baseline, results)
	if len(regressions) != 0 {
		t.Errorf("expected 0 regressions for previously-failing test, got %d", len(regressions))
	}
}

func TestDetectRegressions_NewAssertion(t *testing.T) {
	// New assertion not in baseline is not a regression.
	baseline := &Baseline{Entries: []BaselineEntry{
		{Name: "hover", Status: assertion.StatusPass},
	}}
	results := []assertion.Result{
		{Name: "hover", Status: assertion.StatusPass},
		{Name: "new_test", Status: assertion.StatusFail},
	}

	regressions := DetectRegressions(baseline, results)
	if len(regressions) != 0 {
		t.Errorf("expected 0 regressions for new test, got %d", len(regressions))
	}
}

func TestLoadBaseline_Invalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := LoadBaseline(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadBaseline_Missing(t *testing.T) {
	_, err := LoadBaseline("/nonexistent/baseline.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
