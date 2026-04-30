package report

import (
	"strings"
	"testing"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// TestFormatDiff_IdenticalStrings verifies that identical expected/actual returns empty string.
func TestFormatDiff_IdenticalStrings(t *testing.T) {
	result := FormatDiff("my-assertion", "hello\nworld", "hello\nworld")
	if result != "" {
		t.Errorf("expected empty string for identical inputs, got %q", result)
	}
}

// TestFormatDiff_AddedLines verifies that lines present in actual but not in expected appear with '+'.
func TestFormatDiff_AddedLines(t *testing.T) {
	expected := "line one\nline two"
	actual := "line one\nline two\nline three"

	result := FormatDiff("add-test", expected, actual)

	if result == "" {
		t.Fatal("expected non-empty diff for different inputs")
	}
	if !strings.Contains(result, "--- expected (assertion: add-test)") {
		t.Error("expected diff header with label")
	}
	if !strings.Contains(result, "+++ actual") {
		t.Error("expected +++ actual header")
	}
	if !strings.Contains(result, "+line three") {
		t.Errorf("expected '+line three' in diff, got:\n%s", result)
	}
	// Unchanged lines should appear as context with space prefix.
	if !strings.Contains(result, " line one") {
		t.Errorf("expected ' line one' context line in diff, got:\n%s", result)
	}
}

// TestFormatDiff_RemovedLines verifies that lines present in expected but not actual appear with '-'.
func TestFormatDiff_RemovedLines(t *testing.T) {
	expected := "line one\nline two\nline three"
	actual := "line one\nline three"

	result := FormatDiff("remove-test", expected, actual)

	if result == "" {
		t.Fatal("expected non-empty diff")
	}
	if !strings.Contains(result, "-line two") {
		t.Errorf("expected '-line two' in diff, got:\n%s", result)
	}
	// Unchanged lines appear as context.
	if !strings.Contains(result, " line one") {
		t.Errorf("expected ' line one' context line, got:\n%s", result)
	}
	if !strings.Contains(result, " line three") {
		t.Errorf("expected ' line three' context line, got:\n%s", result)
	}
}

// TestFormatDiff_MixedChanges verifies combination of adds, removes, and context lines.
func TestFormatDiff_MixedChanges(t *testing.T) {
	expected := "alpha\nbeta\ngamma"
	actual := "alpha\ndelta\ngamma"

	result := FormatDiff("mixed-test", expected, actual)

	if result == "" {
		t.Fatal("expected non-empty diff")
	}
	// beta was removed.
	if !strings.Contains(result, "-beta") {
		t.Errorf("expected '-beta' in diff, got:\n%s", result)
	}
	// delta was added.
	if !strings.Contains(result, "+delta") {
		t.Errorf("expected '+delta' in diff, got:\n%s", result)
	}
	// alpha and gamma are context.
	if !strings.Contains(result, " alpha") {
		t.Errorf("expected ' alpha' context line, got:\n%s", result)
	}
	if !strings.Contains(result, " gamma") {
		t.Errorf("expected ' gamma' context line, got:\n%s", result)
	}
}

// TestFormatStatusChange_PassToFail verifies that PASS->FAIL status change is formatted correctly.
func TestFormatStatusChange_PassToFail(t *testing.T) {
	result := FormatStatusChange("my-test", assertion.StatusPass, assertion.StatusFail, "expected foo, got bar")

	if result == "" {
		t.Fatal("expected non-empty status change string")
	}
	// Should contain PASS and FAIL.
	if !strings.Contains(result, "PASS") {
		t.Errorf("expected 'PASS' in status change output, got %q", result)
	}
	if !strings.Contains(result, "FAIL") {
		t.Errorf("expected 'FAIL' in status change output, got %q", result)
	}
	// Should contain detail.
	if !strings.Contains(result, "expected foo, got bar") {
		t.Errorf("expected detail in status change output, got %q", result)
	}
	// Should contain the arrow.
	if !strings.Contains(result, "->") {
		t.Errorf("expected '->' in status change output, got %q", result)
	}
	// Should contain the assertion name.
	if !strings.Contains(result, "my-test:") {
		t.Errorf("expected assertion name 'my-test:' in output, got %q", result)
	}
}

// TestFormatStatusChange_FailToPass verifies that FAIL->PASS status change is formatted correctly.
func TestFormatStatusChange_FailToPass(t *testing.T) {
	result := FormatStatusChange("my-test", assertion.StatusFail, assertion.StatusPass, "")

	if result == "" {
		t.Fatal("expected non-empty status change string")
	}
	if !strings.Contains(result, "FAIL") {
		t.Errorf("expected 'FAIL' in status change output, got %q", result)
	}
	if !strings.Contains(result, "PASS") {
		t.Errorf("expected 'PASS' in status change output, got %q", result)
	}
	if !strings.Contains(result, "->") {
		t.Errorf("expected '->' in status change output, got %q", result)
	}
	// No detail suffix when detail is empty.
	if strings.HasSuffix(result, ": ") {
		t.Errorf("should not have trailing ': ' when detail is empty, got %q", result)
	}
}
