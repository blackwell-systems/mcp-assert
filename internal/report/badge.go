// badge.go generates a shields.io endpoint JSON file for embedding pass/fail
// badges in READMEs and dashboards. The JSON format follows the shields.io
// endpoint schema (https://shields.io/endpoint): schemaVersion, label, message,
// and color. The badge URL is: https://img.shields.io/endpoint?url=<hosted-json>
package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// ShieldsEndpoint is the shields.io endpoint JSON schema. Consumers host this
// file and point a shields.io badge URL at it to render a dynamic badge.
type ShieldsEndpoint struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

// WriteBadge writes a shields.io endpoint JSON file showing "passed/total"
// with a color reflecting the overall status: green (all pass), red (any fail),
// or yellow (all skipped, no passes).
func WriteBadge(results []assertion.Result, path string) error {
	passed, failed, _ := countByStatus(results)
	total := len(results)

	badge := ShieldsEndpoint{
		SchemaVersion: 1,
		Label:         "mcp-assert",
		Message:       fmt.Sprintf("%d/%d", passed, total),
		Color:         badgeColor(passed, failed, total),
	}

	data, err := json.MarshalIndent(badge, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling badge JSON: %w", err)
	}

	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// badgeColor picks the badge color: red if any failures, bright green if all
// pass, yellow otherwise (e.g., all skipped).
func badgeColor(passed, failed, total int) string {
	if failed > 0 {
		return "red"
	}
	if passed == total && total > 0 {
		return "brightgreen"
	}
	return "yellow"
}
