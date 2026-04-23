package report

import (
	"bytes"
	"os"
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

	if !strings.Contains(out, "PASS") {
		t.Error("expected PASS in output")
	}
	if !strings.Contains(out, "hover test") {
		t.Error("expected 'hover test' in output")
	}
	if !strings.Contains(out, "2 passed") {
		t.Error("expected '2 passed' in output")
	}
	if !strings.Contains(out, "0 failed") {
		t.Error("expected '0 failed' in output")
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
