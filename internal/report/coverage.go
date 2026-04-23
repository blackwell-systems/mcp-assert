package report

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// CoverageData is the JSON structure for --coverage-json output.
type CoverageData struct {
	Total          int             `json:"total"`
	Covered        int             `json:"covered"`
	Percentage     int             `json:"percentage"`
	CoveredTools   []CoveredTool   `json:"covered_tools"`
	UncoveredTools []string        `json:"uncovered_tools"`
}

// CoveredTool represents a tool that has at least one assertion.
type CoveredTool struct {
	Name       string `json:"name"`
	Assertions int    `json:"assertions"`
}

// WriteToolCoverageJSON writes coverage data as JSON for CI consumption.
func WriteToolCoverageJSON(serverTools []string, testedTools map[string]int, path string) error {
	sorted := make([]string, len(serverTools))
	copy(sorted, serverTools)
	sort.Strings(sorted)

	var coveredTools []CoveredTool
	var uncoveredTools []string

	for _, tool := range sorted {
		if count := testedTools[tool]; count > 0 {
			coveredTools = append(coveredTools, CoveredTool{Name: tool, Assertions: count})
		} else {
			uncoveredTools = append(uncoveredTools, tool)
		}
	}

	total := len(serverTools)
	covCount := len(coveredTools)
	pct := 0
	if total > 0 {
		pct = (covCount * 100) / total
	}

	data := CoverageData{
		Total:          total,
		Covered:        covCount,
		Percentage:     pct,
		CoveredTools:   coveredTools,
		UncoveredTools: uncoveredTools,
	}

	// Ensure non-nil slices in JSON output.
	if data.CoveredTools == nil {
		data.CoveredTools = []CoveredTool{}
	}
	if data.UncoveredTools == nil {
		data.UncoveredTools = []string{}
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling coverage JSON: %w", err)
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0644)
}
