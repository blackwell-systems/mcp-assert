package report

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// BaselineEntry records the last-known status of an assertion.
type BaselineEntry struct {
	Name     string           `json:"name"`
	Status   assertion.Status `json:"status"`
	Language string           `json:"language,omitempty"`
}

// Baseline is a collection of last-known assertion results.
type Baseline struct {
	Entries []BaselineEntry `json:"entries"`
}

// WriteBaseline writes current results as a baseline file for regression detection.
func WriteBaseline(results []assertion.Result, path string) error {
	// Deduplicate by name (take last trial result per assertion).
	seen := make(map[string]assertion.Result)
	for _, r := range results {
		seen[r.Name] = r
	}

	// Sort by name for deterministic output across runs.
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]BaselineEntry, 0, len(names))
	for _, name := range names {
		r := seen[name]
		entries = append(entries, BaselineEntry{
			Name:     r.Name,
			Status:   r.Status,
			Language: r.Language,
		})
	}

	b := Baseline{Entries: entries}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling baseline: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// LoadBaseline reads a baseline file.
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading baseline %s: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parsing baseline %s: %w", path, err)
	}
	return &b, nil
}

// Regression represents a test that previously passed but now fails.
type Regression struct {
	Name       string
	WasStatus  assertion.Status
	NowStatus  assertion.Status
}

// DetectRegressions compares current results against a baseline.
// Returns regressions: assertions that were PASS in baseline but are not PASS now.
func DetectRegressions(baseline *Baseline, results []assertion.Result) []Regression {
	// Build current status map (last trial per assertion).
	current := make(map[string]assertion.Status)
	for _, r := range results {
		current[r.Name] = r.Status
	}

	var regressions []Regression
	for _, entry := range baseline.Entries {
		if entry.Status != assertion.StatusPass {
			continue // only flag regressions from previously-passing
		}
		nowStatus, exists := current[entry.Name]
		if !exists {
			// Assertion was removed — not a regression, just a gap.
			continue
		}
		if nowStatus != assertion.StatusPass {
			regressions = append(regressions, Regression{
				Name:      entry.Name,
				WasStatus: entry.Status,
				NowStatus: nowStatus,
			})
		}
	}
	return regressions
}

// PrintRegressions prints detected regressions to stdout.
func PrintRegressions(regressions []Regression) {
	if len(regressions) == 0 {
		return
	}
	fmt.Printf("\nRegressions detected (%d):\n", len(regressions))
	for _, r := range regressions {
		fmt.Printf("  %s: was %s, now %s\n", r.Name, r.WasStatus, r.NowStatus)
	}
}
