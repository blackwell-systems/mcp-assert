package report

import (
	"fmt"
	"strings"
)

// FuzzFailure holds display data for a single failed fuzz input.
type FuzzFailure struct {
	Label  string
	Status string // "crash", "timeout", "protocol"
	Detail string
}

// PrintFuzzHeader prints the fuzz run header.
func PrintFuzzHeader(serverName, transport string, seed int64) {
	fmt.Println()
	fmt.Printf("  %s %s\n", colorize(bold, "Server:"), serverName)
	fmt.Printf("  %s %s\n", colorize(bold, "Transport:"), transport)
	fmt.Printf("  %s %d\n", colorize(bold, "Seed:"), seed)
	fmt.Println()
}

// PrintFuzzToolResult prints the result summary for a single tool.
func PrintFuzzToolResult(tool string, runs, passed int, failures []FuzzFailure) {
	name := tool
	if len(name) > 35 {
		name = name[:34] + "…"
	}

	if len(failures) == 0 {
		fmt.Printf("  %s %s (%d/%d passed)\n", passIcon(), colorize(bold, name), passed, runs)
		return
	}

	fmt.Printf("  %s %s (%d/%d passed, %s)\n",
		failIcon(),
		colorize(bold, name),
		passed, runs,
		colorize(red, fmt.Sprintf("%d failures", len(failures))),
	)

	for _, f := range failures {
		var icon string
		switch f.Status {
		case "crash":
			icon = colorize(red, "CRASH")
		case "timeout":
			icon = colorize(yellow, "HANG")
		case "protocol":
			icon = colorize(red, "PROTO")
		default:
			icon = colorize(red, "FAIL")
		}

		detail := f.Detail
		if len(detail) > 80 {
			detail = detail[:79] + "…"
		}

		fmt.Printf("    %s %s %s %s\n",
			icon,
			colorize(gray, f.Label),
			colorize(gray, "→"),
			detail,
		)
	}
}

// PrintFuzzSummary prints the aggregate fuzz run summary.
func PrintFuzzSummary(toolsTested, totalRuns, totalPassed, totalFailures, crashes, timeouts, protocolErrors int) {
	fmt.Println()

	parts := []string{
		fmt.Sprintf("%d tools fuzzed", toolsTested),
		fmt.Sprintf("%d runs", totalRuns),
	}

	if totalPassed > 0 {
		parts = append(parts, colorize(green, fmt.Sprintf("%d passed", totalPassed)))
	}
	if totalFailures > 0 {
		parts = append(parts, colorize(red, fmt.Sprintf("%d failed", totalFailures)))
	}

	fmt.Println("  " + strings.Join(parts, ", "))

	if totalFailures > 0 {
		var breakdown []string
		if crashes > 0 {
			breakdown = append(breakdown, colorize(red, fmt.Sprintf("%d crashes", crashes)))
		}
		if timeouts > 0 {
			breakdown = append(breakdown, colorize(yellow, fmt.Sprintf("%d timeouts", timeouts)))
		}
		if protocolErrors > 0 {
			breakdown = append(breakdown, colorize(red, fmt.Sprintf("%d protocol errors", protocolErrors)))
		}
		fmt.Println("  Breakdown: " + strings.Join(breakdown, ", "))
	}
}
