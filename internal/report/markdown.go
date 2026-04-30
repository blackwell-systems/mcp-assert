package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// WriteMarkdownSummary writes a GitHub Step Summary compatible markdown table.
// If path is empty, writes to $GITHUB_STEP_SUMMARY if set.
func WriteMarkdownSummary(results []assertion.Result, path string) error {
	if path == "" {
		path = os.Getenv("GITHUB_STEP_SUMMARY")
	}
	if path == "" {
		return fmt.Errorf("no output path: set --markdown or $GITHUB_STEP_SUMMARY")
	}

	var b strings.Builder

	passed, failed, skipped := countByStatus(results)

	// Header with pass/fail counts.
	if failed > 0 {
		b.WriteString(fmt.Sprintf("### mcp-assert: %d passed, %d failed\n\n", passed, failed))
	} else {
		b.WriteString(fmt.Sprintf("### mcp-assert: %d/%d passed\n\n", passed, len(results)))
	}

	// Results table.
	b.WriteString("| Status | Assertion | Duration |\n")
	b.WriteString("|--------|-----------|----------|\n")

	for _, r := range results {
		icon := statusIcon(r.Status)
		suffix := ""
		if r.Language != "" {
			suffix = fmt.Sprintf(" (%s)", r.Language)
		}
		b.WriteString(fmt.Sprintf("| %s | %s%s | %dms |\n", icon, r.Name, suffix, r.Duration.Milliseconds()))
	}

	// Failure details.
	if failed > 0 {
		b.WriteString("\n<details><summary>Failure details</summary>\n\n")
		for _, r := range results {
			if r.Status == assertion.StatusFail {
				b.WriteString(fmt.Sprintf("**%s**\n```\n%s\n```\n\n", r.Name, r.Detail))
			}
		}
		b.WriteString("</details>\n")
	}

	if skipped > 0 {
		b.WriteString(fmt.Sprintf("\n_%d assertion(s) skipped_\n", skipped))
	}

	// Reliability section when trials > 1.
	if hasMultipleTrials(results) {
		stats := ComputeReliability(results)
		capable, reliable := 0, 0
		for _, s := range stats {
			if s.PassAt {
				capable++
			}
			if s.PassUp {
				reliable++
			}
		}
		total := len(stats)

		b.WriteString("\n### Reliability\n\n")
		b.WriteString("| Assertion | Trials | Passed | pass@k | pass^k |\n")
		b.WriteString("|-----------|--------|--------|--------|--------|\n")
		for _, s := range stats {
			passAt := "YES"
			if !s.PassAt {
				passAt = "NO"
			}
			passUp := "YES"
			if !s.PassUp {
				passUp = "NO"
			}
			b.WriteString(fmt.Sprintf("| %s | %d | %d | %s | %s |\n", s.Name, s.Trials, s.Passed, passAt, passUp))
		}
		b.WriteString(fmt.Sprintf("\n**pass@k: %d/%d capable, pass^k: %d/%d reliable**\n", capable, total, reliable, total))
	}

	// Append to file (GitHub Step Summary expects append).
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("writing markdown summary: %w", err)
	}
	if _, err = f.WriteString(b.String()); err != nil {
		_ = f.Close()
		return fmt.Errorf("writing markdown summary: %w", err)
	}
	return f.Close()
}

func statusIcon(s assertion.Status) string {
	switch s {
	case assertion.StatusPass:
		return "PASS"
	case assertion.StatusFail:
		return "FAIL"
	case assertion.StatusSkip:
		return "SKIP"
	default:
		return "?"
	}
}

func countByStatus(results []assertion.Result) (passed, failed, skipped int) {
	for _, r := range results {
		switch r.Status {
		case assertion.StatusPass:
			passed++
		case assertion.StatusFail:
			failed++
		case assertion.StatusSkip:
			skipped++
		}
	}
	return
}
