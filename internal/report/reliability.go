package report

import (
	"fmt"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// ReliabilityStats holds pass@k and pass^k metrics for an assertion.
type ReliabilityStats struct {
	Name   string
	Trials int
	Passed int
	PassAt bool    // pass@k: passed at least once (capability)
	PassUp bool    // pass^k: passed every time (reliability)
	Rate   float64 // pass rate as 0.0-1.0
}

// ComputeReliability groups results by assertion name and computes metrics.
func ComputeReliability(results []assertion.Result) []ReliabilityStats {
	groups := make(map[string][]assertion.Result)
	order := []string{}

	for _, r := range results {
		if _, seen := groups[r.Name]; !seen {
			order = append(order, r.Name)
		}
		groups[r.Name] = append(groups[r.Name], r)
	}

	var stats []ReliabilityStats
	for _, name := range order {
		trials := groups[name]
		passed := 0
		for _, t := range trials {
			if t.Status == assertion.StatusPass {
				passed++
			}
		}
		stats = append(stats, ReliabilityStats{
			Name:   name,
			Trials: len(trials),
			Passed: passed,
			PassAt: passed > 0,
			PassUp: passed == len(trials),
			Rate:   float64(passed) / float64(len(trials)),
		})
	}
	return stats
}

// PrintReliability prints a reliability table when trials > 1.
func PrintReliability(results []assertion.Result) {
	stats := ComputeReliability(results)

	// Only print if there are actually multiple trials.
	if len(stats) == 0 || stats[0].Trials <= 1 {
		return
	}

	fmt.Println()
	fmt.Println("Reliability:")

	// Find max name length for alignment.
	maxName := 0
	for _, s := range stats {
		if len(s.Name) > maxName {
			maxName = len(s.Name)
		}
	}
	if maxName > 50 {
		maxName = 50
	}

	fmt.Printf("  %-*s  %6s  %6s  %8s  %6s\n", maxName, "Assertion", "Trials", "Passed", "pass@k", "pass^k")
	fmt.Printf("  %s  %s  %s  %s  %s\n",
		strings.Repeat("-", maxName), strings.Repeat("-", 6), strings.Repeat("-", 6), strings.Repeat("-", 8), strings.Repeat("-", 6))

	for _, s := range stats {
		name := s.Name
		if len(name) > maxName {
			name = name[:maxName-1] + "…"
		}
		passAt := "YES"
		if !s.PassAt {
			passAt = "NO"
		}
		passUp := "YES"
		if !s.PassUp {
			passUp = "NO"
		}
		fmt.Printf("  %-*s  %6d  %6d  %8s  %6s\n", maxName, name, s.Trials, s.Passed, passAt, passUp)
	}

	// Summary line.
	total := len(stats)
	capable := 0
	reliable := 0
	for _, s := range stats {
		if s.PassAt {
			capable++
		}
		if s.PassUp {
			reliable++
		}
	}
	fmt.Printf("\n  pass@k: %d/%d capable, pass^k: %d/%d reliable\n", capable, total, reliable, total)
}

