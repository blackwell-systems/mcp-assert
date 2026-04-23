package runner

import (
	"fmt"
	"os"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

func extractText(result *mcp.CallToolResult) string {
	var parts []string
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func countFails(results []assertion.Result) int {
	n := 0
	for _, r := range results {
		if r.Status == assertion.StatusFail {
			n++
		}
	}
	return n
}

func countPasses(results []assertion.Result) int {
	n := 0
	for _, r := range results {
		if r.Status == assertion.StatusPass {
			n++
		}
	}
	return n
}

// writeReports writes optional structured report files. Errors are printed
// to stderr but do not fail the run — reporting is best-effort.
func writeReports(results []assertion.Result, junitPath, markdownPath, badgePath string) {
	if junitPath != "" {
		if err := report.WriteJUnit(results, junitPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: junit: %v\n", err)
		}
	}
	if markdownPath != "" {
		if err := report.WriteMarkdownSummary(results, markdownPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: markdown: %v\n", err)
		}
	}
	if badgePath != "" {
		if err := report.WriteBadge(results, badgePath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: badge: %v\n", err)
		}
	}
}

// applyServerOverride parses a "--server" string like "agent-lsp go:gopls"
// and replaces the assertion's server config.
func applyServerOverride(a *assertion.Assertion, serverSpec string) {
	parts := strings.Fields(serverSpec)
	if len(parts) == 0 {
		return
	}
	a.Server.Command = parts[0]
	a.Server.Args = parts[1:]
}
