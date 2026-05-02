package report

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// FuzzToolReport holds per-tool fuzz data for structured report output.
type FuzzToolReport struct {
	Tool       string
	Runs       int
	Passed     int
	Failures   []FuzzFailure
	DurationMS int64
}

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

// WriteFuzzJUnit writes JUnit XML for fuzz results. Each tool becomes a test
// suite; each fuzz input that fails becomes a test case failure.
func WriteFuzzJUnit(tools []FuzzToolReport, path string) error {
	var suites []junitTestSuite

	for _, t := range tools {
		failures := len(t.Failures)
		var cases []junitTestCase

		// One passing test case per tool if there are any passes.
		if t.Passed > 0 {
			cases = append(cases, junitTestCase{
				Name:      fmt.Sprintf("%s: %d/%d inputs handled", t.Tool, t.Passed, t.Runs),
				Classname: "mcp-assert-fuzz",
				Time:      float64(t.DurationMS) / 1000.0,
			})
		}

		// One failing test case per failure.
		for _, f := range t.Failures {
			cases = append(cases, junitTestCase{
				Name:      fmt.Sprintf("%s: %s", t.Tool, f.Label),
				Classname: "mcp-assert-fuzz",
				Failure: &junitFailure{
					Message: fmt.Sprintf("[%s] %s", f.Status, f.Label),
					Text:    f.Detail,
				},
			})
		}

		suites = append(suites, junitTestSuite{
			Name:     t.Tool,
			Tests:    t.Passed + failures,
			Failures: failures,
			Time:     float64(t.DurationMS) / 1000.0,
			Cases:    cases,
		})
	}

	out := junitTestSuites{Suites: suites}
	data, err := xml.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling fuzz JUnit XML: %w", err)
	}

	output := []byte(xml.Header)
	output = append(output, data...)
	output = append(output, '\n')

	return os.WriteFile(path, output, 0644)
}

// WriteFuzzMarkdown writes a GitHub Step Summary compatible markdown table
// for fuzz results.
func WriteFuzzMarkdown(tools []FuzzToolReport, toolsTested, totalRuns, totalPassed, totalFailures int, seed int64, path string) error {
	var b strings.Builder

	if totalFailures > 0 {
		b.WriteString(fmt.Sprintf("### mcp-assert fuzz: %d passed, %d failed (%d tools, seed %d)\n\n", totalPassed, totalFailures, toolsTested, seed))
	} else {
		b.WriteString(fmt.Sprintf("### mcp-assert fuzz: %d/%d passed (%d tools, seed %d)\n\n", totalPassed, totalRuns, toolsTested, seed))
	}

	b.WriteString("| Tool | Runs | Passed | Failed | Status |\n")
	b.WriteString("|------|------|--------|--------|--------|\n")

	for _, t := range tools {
		failures := len(t.Failures)
		status := "PASS"
		if failures > 0 {
			status = "FAIL"
		}
		b.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %s |\n", t.Tool, t.Runs, t.Passed, failures, status))
	}

	if totalFailures > 0 {
		b.WriteString("\n<details><summary>Failure details</summary>\n\n")
		for _, t := range tools {
			for _, f := range t.Failures {
				detail := f.Detail
				if len(detail) > 200 {
					detail = detail[:200] + "..."
				}
				b.WriteString(fmt.Sprintf("**%s: %s** [%s]\n```\n%s\n```\n\n", t.Tool, f.Label, f.Status, detail))
			}
		}
		b.WriteString("</details>\n")
	}

	b.WriteString(fmt.Sprintf("\n_Reproduce: `mcp-assert fuzz --server \"...\" --seed %d`_\n", seed))

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("writing fuzz markdown: %w", err)
	}
	if _, err = f.WriteString(b.String()); err != nil {
		_ = f.Close()
		return fmt.Errorf("writing fuzz markdown: %w", err)
	}
	return f.Close()
}
