package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// FixSuggestion holds a suggested correction for a position-sensitive assertion.
type FixSuggestion struct {
	AssertionName string
	OriginalArgs  map[string]any
	FixedArgs     map[string]any
	YAMLPatch     string // unified diff showing the YAML change
}

// IsPositionError returns true if the error detail indicates a position-sensitive failure.
// It checks (case-insensitively) for "no identifier found" or "column is beyond end of line".
func IsPositionError(detail string) bool {
	lower := strings.ToLower(detail)
	return strings.Contains(lower, "no identifier found") ||
		strings.Contains(lower, "column is beyond end of line")
}

// toInt converts a value that may be int or float64 (from YAML decode) to int.
// Returns the value and true on success, or 0 and false if the value is absent/wrong type.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	case int64:
		return int(n), true
	}
	return 0, false
}

// cloneArgs makes a shallow copy of an args map.
func cloneArgs(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// buildYAMLPatch produces a unified-diff-style YAML patch showing the change
// from (origLine, origCol) to (newLine, newCol).
func buildYAMLPatch(origLine, origCol, newLine, newCol int) string {
	var sb strings.Builder
	sb.WriteString("--- original\n")
	sb.WriteString("+++ suggested\n")
	sb.WriteString("@@ fix @@\n")
	fmt.Fprintf(&sb, "-    line: %d\n", origLine)
	fmt.Fprintf(&sb, "-    column: %d\n", origCol)
	fmt.Fprintf(&sb, "+    line: %d\n", newLine)
	fmt.Fprintf(&sb, "+    column: %d\n", newCol)
	return sb.String()
}

// ScanNearbyPositions re-runs a tool call with nearby line/column values and
// returns the first position that succeeds. It searches in a spiral pattern:
// first the same line with columns ±1..colRange, then adjacent lines ±1..lineRange
// with columns ±0..colRange.
//
// Returns nil, nil when args lack line/column fields or no nearby position passes.
func ScanNearbyPositions(
	a assertion.Assertion,
	fixture string,
	timeout time.Duration,
	dockerImage string,
	lineRange int,
	colRange int,
) (*FixSuggestion, error) {
	origArgs := a.Assert.Args

	origLine, hasLine := toInt(origArgs["line"])
	origCol, hasCol := toInt(origArgs["column"])
	if !hasLine || !hasCol {
		// Nothing to scan without position fields.
		return nil, nil
	}

	// Build candidate list: (line, col) pairs to try.
	// Strategy: same line first (col ±1..colRange), then adjacent lines (±1..lineRange)
	// with col ±0..colRange.
	type candidate struct{ line, col int }
	var candidates []candidate

	// Same line, columns ±1..colRange.
	for delta := 1; delta <= colRange; delta++ {
		if origCol+delta > 0 {
			candidates = append(candidates, candidate{origLine, origCol + delta})
		}
		if origCol-delta > 0 {
			candidates = append(candidates, candidate{origLine, origCol - delta})
		}
	}

	// Adjacent lines, columns ±0..colRange.
	for lineDelta := 1; lineDelta <= lineRange; lineDelta++ {
		for _, lineOff := range []int{lineDelta, -lineDelta} {
			newLine := origLine + lineOff
			if newLine < 1 {
				continue
			}
			// Try original column first, then spread outward.
			candidates = append(candidates, candidate{newLine, origCol})
			for colDelta := 1; colDelta <= colRange; colDelta++ {
				if origCol+colDelta > 0 {
					candidates = append(candidates, candidate{newLine, origCol + colDelta})
				}
				if origCol-colDelta > 0 {
					candidates = append(candidates, candidate{newLine, origCol - colDelta})
				}
			}
		}
	}

	for _, c := range candidates {
		newArgs := cloneArgs(origArgs)
		newArgs["line"] = c.line
		newArgs["column"] = c.col

		candidate := a
		candidate.Assert.Args = newArgs

		r := runAssertion(candidate, fixture, timeout, dockerImage)
		if r.Status == assertion.StatusPass {
			patch := buildYAMLPatch(origLine, origCol, c.line, c.col)
			return &FixSuggestion{
				AssertionName: a.Name,
				OriginalArgs:  cloneArgs(origArgs),
				FixedArgs:     newArgs,
				YAMLPatch:     patch,
			}, nil
		}
	}

	return nil, nil
}

// PrintFixSuggestions prints each suggestion to stdout with the assertion name
// and the YAML patch. Does nothing when suggestions is empty.
func PrintFixSuggestions(suggestions []FixSuggestion) {
	if len(suggestions) == 0 {
		return
	}
	fmt.Println("\n--fix suggestions:")
	for _, s := range suggestions {
		fmt.Printf("  assertion: %s\n", s.AssertionName)
		fmt.Println(s.YAMLPatch)
	}
}
