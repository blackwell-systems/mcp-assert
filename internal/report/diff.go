package report

import (
	"fmt"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// FormatDiff produces a unified diff string between two text values.
// It returns an empty string when expected and actual are identical.
// The diff uses a simple line-by-line LCS algorithm with no context lines.
func FormatDiff(label, expected, actual string) string {
	if expected == actual {
		return ""
	}

	expLines := splitLines(expected)
	actLines := splitLines(actual)

	ops := lcs(expLines, actLines)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- expected (assertion: %s)\n", label))
	sb.WriteString("+++ actual\n")
	sb.WriteString("@@ @@\n")
	for _, op := range ops {
		switch op.kind {
		case opEqual:
			sb.WriteString(" " + op.line + "\n")
		case opRemove:
			sb.WriteString(colorize(red, "-"+op.line) + "\n")
		case opAdd:
			sb.WriteString(colorize(green, "+"+op.line) + "\n")
		}
	}
	return sb.String()
}

// FormatStatusChange returns a formatted status change indicator line.
// It includes color if the terminal supports it.
func FormatStatusChange(name string, prevStatus, currStatus assertion.Status, detail string) string {
	prev := colorize(yellow, string(prevStatus))
	curr := string(currStatus)
	switch currStatus {
	case assertion.StatusPass:
		curr = colorize(green, curr)
	case assertion.StatusFail:
		curr = colorize(red, curr)
	}

	suffix := ""
	if detail != "" {
		suffix = ": " + detail
	}
	return fmt.Sprintf("  status changed: %s -> %s%s", prev, curr, suffix)
}

// splitLines splits a string into lines. It handles trailing newlines gracefully
// by discarding a single trailing empty string.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// Drop trailing empty element produced by a terminal newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// diffOp kinds.
const (
	opEqual  = 0
	opRemove = 1
	opAdd    = 2
)

type diffOp struct {
	kind int
	line string
}

// lcs computes the diff operations between two line slices using longest
// common subsequence. Returns a flat list of equal/remove/add operations.
func lcs(a, b []string) []diffOp {
	m, n := len(a), len(b)

	// dp[i][j] = length of LCS of a[:i] and b[:j].
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to reconstruct diff operations.
	ops := make([]diffOp, 0, m+n)
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && a[i-1] == b[j-1]:
			ops = append(ops, diffOp{opEqual, a[i-1]})
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			ops = append(ops, diffOp{opAdd, b[j-1]})
			j--
		default:
			ops = append(ops, diffOp{opRemove, a[i-1]})
			i--
		}
	}

	// Reverse (backtracking produces reverse order).
	for lo, hi := 0, len(ops)-1; lo < hi; lo, hi = lo+1, hi-1 {
		ops[lo], ops[hi] = ops[hi], ops[lo]
	}
	return ops
}
