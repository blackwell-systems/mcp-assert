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
		icon := "PASS"
		switch r.Status {
		case assertion.StatusPass:
			icon = "PASS"
			passed++
		case assertion.StatusFail:
			icon = "FAIL"
			failed++
		case assertion.StatusSkip:
			icon = "SKIP"
			skipped++
		}

		suffix := ""
		if r.Language != "" {
			suffix = fmt.Sprintf(" (%s)", r.Language)
		}

		line := fmt.Sprintf("%-4s  %-60s %6dms", icon, r.Name+suffix, r.Duration.Milliseconds())
		fmt.Println(line)

		if r.Status == assertion.StatusFail && r.Detail != "" {
			fmt.Printf("      %s\n", r.Detail)
		}
	}

	fmt.Println()
	fmt.Printf("%d assertions, %d passed, %d failed, %d skipped\n", len(results), passed, failed, skipped)
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
			row += fmt.Sprintf("  %-18s", string(status))
		}
		fmt.Println(row)
	}
}
