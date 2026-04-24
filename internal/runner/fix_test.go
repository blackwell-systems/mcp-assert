package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// TestIsPositionError verifies detection of position-sensitive error messages.
func TestIsPositionError(t *testing.T) {
	tests := []struct {
		name   string
		detail string
		want   bool
	}{
		{
			name:   "no identifier found exact",
			detail: "no identifier found",
			want:   true,
		},
		{
			name:   "no identifier found with context",
			detail: "no identifier found at position (10, 5)",
			want:   true,
		},
		{
			name:   "no identifier found uppercase",
			detail: "No Identifier Found at line 5",
			want:   true,
		},
		{
			name:   "column is beyond end of line exact",
			detail: "column is beyond end of line",
			want:   true,
		},
		{
			name:   "column is beyond end of line with context",
			detail: "error: column is beyond end of line (col 200, line length 80)",
			want:   true,
		},
		{
			name:   "column is beyond end of line uppercase",
			detail: "Column Is Beyond End Of Line",
			want:   true,
		},
		{
			name:   "unrelated error",
			detail: "connection refused",
			want:   false,
		},
		{
			name:   "empty string",
			detail: "",
			want:   false,
		},
		{
			name:   "partial match - column only",
			detail: "column",
			want:   false,
		},
		{
			name:   "partial match - no identifier without found",
			detail: "no identifier",
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsPositionError(tc.detail)
			if got != tc.want {
				t.Errorf("IsPositionError(%q) = %v, want %v", tc.detail, got, tc.want)
			}
		})
	}
}

// TestScanNearbyPositions_NoLineColumn verifies graceful handling when args
// lack line/column fields — returns nil, nil.
func TestScanNearbyPositions_NoLineColumn(t *testing.T) {
	a := assertion.Assertion{
		Name: "test-no-position",
		Assert: assertion.AssertBlock{
			Tool: "some_tool",
			Args: map[string]any{
				"uri": "file:///example.go",
			},
		},
	}

	result, err := ScanNearbyPositions(a, "", 5*time.Second, "", 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result when args lack line/column, got: %+v", result)
	}
}

// TestScanNearbyPositions_NoLineOnly verifies graceful handling when only
// line is present but column is absent.
func TestScanNearbyPositions_NoLineOnly(t *testing.T) {
	a := assertion.Assertion{
		Name: "test-line-no-col",
		Assert: assertion.AssertBlock{
			Tool: "some_tool",
			Args: map[string]any{
				"line": 10,
			},
		},
	}

	result, err := ScanNearbyPositions(a, "", 5*time.Second, "", 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result when column is missing, got: %+v", result)
	}
}

// TestScanNearbyPositions_NoColOnly verifies graceful handling when only
// column is present but line is absent.
func TestScanNearbyPositions_NoColOnly(t *testing.T) {
	a := assertion.Assertion{
		Name: "test-col-no-line",
		Assert: assertion.AssertBlock{
			Tool: "some_tool",
			Args: map[string]any{
				"column": 5,
			},
		},
	}

	result, err := ScanNearbyPositions(a, "", 5*time.Second, "", 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result when line is missing, got: %+v", result)
	}
}

// TestFixSuggestion_YAMLPatch verifies the YAML patch output is in unified
// diff format with the correct before/after values.
func TestFixSuggestion_YAMLPatch(t *testing.T) {
	tests := []struct {
		name                   string
		origLine, origCol      int
		newLine, newCol        int
		wantMinusLine          string
		wantMinusCol           string
		wantPlusLine           string
		wantPlusCol            string
	}{
		{
			name:          "column shift right",
			origLine:      10,
			origCol:       15,
			newLine:       10,
			newCol:        16,
			wantMinusLine: "-    line: 10",
			wantMinusCol:  "-    column: 15",
			wantPlusLine:  "+    line: 10",
			wantPlusCol:   "+    column: 16",
		},
		{
			name:          "line shift down with column change",
			origLine:      10,
			origCol:       15,
			newLine:       11,
			newCol:        8,
			wantMinusLine: "-    line: 10",
			wantMinusCol:  "-    column: 15",
			wantPlusLine:  "+    line: 11",
			wantPlusCol:   "+    column: 8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			patch := buildYAMLPatch(tc.origLine, tc.origCol, tc.newLine, tc.newCol)

			if !strings.Contains(patch, "--- original") {
				t.Errorf("patch missing '--- original' header:\n%s", patch)
			}
			if !strings.Contains(patch, "+++ suggested") {
				t.Errorf("patch missing '+++ suggested' header:\n%s", patch)
			}
			if !strings.Contains(patch, "@@ fix @@") {
				t.Errorf("patch missing '@@ fix @@' hunk header:\n%s", patch)
			}
			if !strings.Contains(patch, tc.wantMinusLine) {
				t.Errorf("patch missing %q:\n%s", tc.wantMinusLine, patch)
			}
			if !strings.Contains(patch, tc.wantMinusCol) {
				t.Errorf("patch missing %q:\n%s", tc.wantMinusCol, patch)
			}
			if !strings.Contains(patch, tc.wantPlusLine) {
				t.Errorf("patch missing %q:\n%s", tc.wantPlusLine, patch)
			}
			if !strings.Contains(patch, tc.wantPlusCol) {
				t.Errorf("patch missing %q:\n%s", tc.wantPlusCol, patch)
			}
		})
	}
}

// TestFixSuggestion_YAMLPatch_FullFormat checks the exact line order and
// format of the patch produced by buildYAMLPatch.
func TestFixSuggestion_YAMLPatch_FullFormat(t *testing.T) {
	patch := buildYAMLPatch(10, 15, 11, 8)
	want := "--- original\n+++ suggested\n@@ fix @@\n-    line: 10\n-    column: 15\n+    line: 11\n+    column: 8\n"
	if patch != want {
		t.Errorf("patch mismatch:\ngot:  %q\nwant: %q", patch, want)
	}
}

// captureStdout redirects os.Stdout, runs f, then returns the captured output.
func captureStdout(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("os.Pipe: %v", err))
	}
	old := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	return buf.String()
}

// TestPrintFixSuggestions_Empty verifies no output is produced when there
// are no suggestions.
func TestPrintFixSuggestions_Empty(t *testing.T) {
	out := captureStdout(func() {
		PrintFixSuggestions(nil)
	})
	if out != "" {
		t.Errorf("expected no output for empty suggestions, got: %q", out)
	}

	out = captureStdout(func() {
		PrintFixSuggestions([]FixSuggestion{})
	})
	if out != "" {
		t.Errorf("expected no output for empty slice, got: %q", out)
	}
}

// TestPrintFixSuggestions_OneSuggestion verifies that a single suggestion
// is printed with the assertion name and YAML patch.
func TestPrintFixSuggestions_OneSuggestion(t *testing.T) {
	s := FixSuggestion{
		AssertionName: "my-assertion",
		OriginalArgs:  map[string]any{"line": 10, "column": 15},
		FixedArgs:     map[string]any{"line": 11, "column": 8},
		YAMLPatch:     buildYAMLPatch(10, 15, 11, 8),
	}

	out := captureStdout(func() {
		PrintFixSuggestions([]FixSuggestion{s})
	})

	if !strings.Contains(out, "--fix suggestions:") {
		t.Errorf("output missing '--fix suggestions:' header:\n%s", out)
	}
	if !strings.Contains(out, "my-assertion") {
		t.Errorf("output missing assertion name:\n%s", out)
	}
	if !strings.Contains(out, "-    line: 10") {
		t.Errorf("output missing original line:\n%s", out)
	}
	if !strings.Contains(out, "+    line: 11") {
		t.Errorf("output missing fixed line:\n%s", out)
	}
}

// TestPrintFixSuggestions_MultipleSuggestions verifies that multiple
// suggestions all appear in the output.
func TestPrintFixSuggestions_MultipleSuggestions(t *testing.T) {
	suggestions := []FixSuggestion{
		{
			AssertionName: "assertion-one",
			YAMLPatch:     buildYAMLPatch(5, 10, 5, 11),
		},
		{
			AssertionName: "assertion-two",
			YAMLPatch:     buildYAMLPatch(20, 3, 21, 1),
		},
	}

	out := captureStdout(func() {
		PrintFixSuggestions(suggestions)
	})

	if !strings.Contains(out, "assertion-one") {
		t.Errorf("output missing first assertion name:\n%s", out)
	}
	if !strings.Contains(out, "assertion-two") {
		t.Errorf("output missing second assertion name:\n%s", out)
	}
}
