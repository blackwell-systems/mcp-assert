package report

import (
	"fmt"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// PrintResults prints a human-readable results table.
func PrintResults(results []assertion.Result) {
	passed, failed, skipped := 0, 0, 0

	for _, r := range results {
		var icon string
		switch r.Status {
		case assertion.StatusPass:
			icon = passIcon()
			passed++
		case assertion.StatusFail:
			icon = failIcon()
			failed++
		case assertion.StatusSkip:
			icon = skipIcon()
			skipped++
		}

		suffix := ""
		if r.Language != "" {
			suffix = fmt.Sprintf(" %s", colorize(gray, "("+r.Language+")"))
		}

		dur := colorize(gray, fmt.Sprintf("%dms", r.Duration.Milliseconds()))
		line := fmt.Sprintf("%-4s  %-60s %6s", icon, r.Name+suffix, dur)
		fmt.Println(line)

		if r.Status == assertion.StatusFail && r.Detail != "" {
			fmt.Printf("      %s\n", colorize(red, r.Detail))
		}
	}

	fmt.Println()

	// Summary line.
	total := len(results)
	parts := []string{fmt.Sprintf("%d assertions", total)}
	if passed > 0 {
		parts = append(parts, colorize(green, fmt.Sprintf("%d passed", passed)))
	}
	if failed > 0 {
		parts = append(parts, colorize(red, fmt.Sprintf("%d failed", failed)))
	}
	if skipped > 0 {
		parts = append(parts, colorize(yellow, fmt.Sprintf("%d skipped", skipped)))
	}
	fmt.Println(strings.Join(parts, ", "))
}

// PrintMatrix prints a cross-language matrix table.
func PrintMatrix(results []assertion.Result) {
	// Collect unique languages and assertion names.
	langSet := make(map[string]bool)
	nameSet := make(map[string]bool)
	resultMap := make(map[string]assertion.Status) // "lang:name" -> status

	for _, r := range results {
		langSet[r.Language] = true
		nameSet[r.Name] = true
		resultMap[r.Language+":"+r.Name] = r.Status
	}

	var langs []string
	for l := range langSet {
		langs = append(langs, l)
	}
	var names []string
	for n := range nameSet {
		names = append(names, n)
	}

	// Print header.
	header := fmt.Sprintf("%-20s", "")
	for _, name := range names {
		short := name
		if len(short) > 18 {
			short = short[:18]
		}
		header += fmt.Sprintf("  %-18s", short)
	}
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	// Print rows.
	for _, lang := range langs {
		row := fmt.Sprintf("%-20s", lang)
		for _, name := range names {
			status := resultMap[lang+":"+name]
			cell := string(status)
			switch status {
			case assertion.StatusPass:
				cell = colorize(green, "PASS")
			case assertion.StatusFail:
				cell = colorize(red, "FAIL")
			case assertion.StatusSkip:
				cell = colorize(yellow, "SKIP")
			}
			row += fmt.Sprintf("  %-18s", cell)
		}
		fmt.Println(row)
	}
}
