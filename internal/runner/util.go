package runner

import (
	"fmt"
	"os"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// extractText concatenates all TextContent entries in a tool result into a
// single string. Non-text content (images, embedded resources) is ignored.
func extractText(result *mcp.CallToolResult) string {
	var parts []string
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// countFails returns the number of results with a Fail status.
func countFails(results []assertion.Result) int {
	n := 0
	for _, r := range results {
		if r.Status == assertion.StatusFail {
			n++
		}
	}
	return n
}

// countPasses returns the number of results with a Pass status.
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

// parseServerSpec parses a "--server" string like "agent-lsp go:gopls" into a
// ServerConfig. This is the shared entry point for commands that accept a
// server override string (coverage, generate, audit, etc.).
func parseServerSpec(serverSpec string) (assertion.ServerConfig, error) {
	parts := strings.Fields(serverSpec)
	if len(parts) == 0 {
		return assertion.ServerConfig{}, fmt.Errorf("--server cannot be empty")
	}
	return assertion.ServerConfig{
		Command: parts[0],
		Args:    parts[1:],
	}, nil
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
