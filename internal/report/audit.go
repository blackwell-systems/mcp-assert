package report

import (
	"fmt"
	"strings"
	"time"
)

// AuditToolResult holds display data for a single audited tool.
type AuditToolResult struct {
	Tool     string
	Status   string // "healthy", "crash", "timeout", "skipped"
	Detail   string
	Duration time.Duration
}

// PrintAuditHeader prints server info and quality score.
func PrintAuditHeader(serverName, transport string, score int) {
	fmt.Println()
	fmt.Printf("  %s %s\n", colorize(bold, "Server:"), serverName)
	fmt.Printf("  %s %s\n", colorize(bold, "Transport:"), transport)

	scoreColor := green
	if score < 80 {
		scoreColor = yellow
	}
	if score < 50 {
		scoreColor = red
	}
	fmt.Printf("  %s %s\n", colorize(bold, "Score:"), colorize(scoreColor, fmt.Sprintf("%d%%", score)))
	fmt.Println()
}

// PrintAuditResults prints the per-tool results table.
func PrintAuditResults(results []AuditToolResult) {
	for _, r := range results {
		var icon string
		switch r.Status {
		case "healthy":
			icon = passIcon()
		case "crash":
			icon = failIcon()
		case "timeout":
			icon = failIcon()
		case "skipped":
			icon = skipIcon()
		}

		dur := colorize(gray, fmt.Sprintf("%dms", r.Duration.Milliseconds()))
		name := r.Tool
		if len(name) > 35 {
			name = name[:34] + "…"
		}

		line := fmt.Sprintf("  %-4s %-36s %6s  %s", icon, name, dur, colorize(gray, r.Detail))
		fmt.Println(line)
	}
}

// PrintAuditSummary prints the summary counts.
func PrintAuditSummary(healthy, crashes, timeouts, skipped, total int) {
	fmt.Println()

	parts := []string{fmt.Sprintf("%d tools tested", total-skipped)}
	if healthy > 0 {
		parts = append(parts, colorize(green, fmt.Sprintf("%d healthy", healthy)))
	}
	if crashes > 0 {
		parts = append(parts, colorize(red, fmt.Sprintf("%d crashed", crashes)))
	}
	if timeouts > 0 {
		parts = append(parts, colorize(red, fmt.Sprintf("%d timed out", timeouts)))
	}
	if skipped > 0 {
		parts = append(parts, colorize(yellow, fmt.Sprintf("%d skipped (destructive)", skipped)))
	}
	fmt.Println("  " + strings.Join(parts, ", "))
}

// PrintAuditNextSteps prints the YAML generation path and guidance.
func PrintAuditNextSteps(generatedAt string) {
	fmt.Println()
	fmt.Println(colorize(bold, "  Generated assertion files:"))
	fmt.Printf("    %s/\n", generatedAt)
	fmt.Println()
	fmt.Println(colorize(bold, "  Next steps:"))
	fmt.Println("    1. Review and customize the generated YAML assertions")
	fmt.Println("    2. Add expected content, setup steps, and multi-step workflows")
	fmt.Printf("    3. Run in CI: mcp-assert ci --suite %s\n", generatedAt)
	fmt.Println()
	fmt.Printf("  %s The audit tests crash resistance and error handling.\n", colorize(cyan, "Tip:"))
	fmt.Println("  For deeper coverage (expected outputs, multi-step flows, state")
	fmt.Println("  verification), edit the generated YAML files or write new ones.")
	fmt.Println("  See: https://github.com/blackwell-systems/mcp-assert#writing-assertions")
	fmt.Println()
}
