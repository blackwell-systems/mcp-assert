package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// ShieldsEndpoint is the shields.io JSON endpoint schema.
type ShieldsEndpoint struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

// WriteBadge writes a shields.io endpoint JSON file.
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

func badgeColor(passed, failed, total int) string {
	if failed > 0 {
		return "red"
	}
	if passed == total && total > 0 {
		return "brightgreen"
	}
	return "yellow"
}
