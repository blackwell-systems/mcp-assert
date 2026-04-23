package report

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Snapshot stores a captured tool response for comparison.
type Snapshot struct {
	Name     string `json:"name"`
	Tool     string `json:"tool"`
	Text     string `json:"text"`
	IsError  bool   `json:"is_error"`
	Checksum string `json:"checksum"`
}

// SnapshotFile holds all snapshots for a suite.
type SnapshotFile struct {
	Snapshots []Snapshot `json:"snapshots"`
}

// SnapshotPath returns the .snapshot file path for a suite directory.
func SnapshotPath(suiteDir string) string {
	return filepath.Join(suiteDir, ".snapshots.json")
}

// LoadSnapshots reads the snapshot file for a suite.
func LoadSnapshots(suiteDir string) (*SnapshotFile, error) {
	path := SnapshotPath(suiteDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SnapshotFile{}, nil
		}
		return nil, fmt.Errorf("reading snapshots: %w", err)
	}
	var sf SnapshotFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parsing snapshots: %w", err)
	}
	return &sf, nil
}

// SaveSnapshots writes the snapshot file for a suite.
func SaveSnapshots(suiteDir string, sf *SnapshotFile) error {
	path := SnapshotPath(suiteDir)
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling snapshots: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// Checksum returns a SHA-256 hash of the response text.
func Checksum(text string) string {
	h := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", h[:8])
}

// CompareSnapshot compares a tool response against a saved snapshot.
// Returns nil if they match, or an error describing the difference.
func CompareSnapshot(saved Snapshot, actualText string, actualIsError bool) error {
	if saved.IsError != actualIsError {
		return fmt.Errorf("isError changed: was %v, now %v", saved.IsError, actualIsError)
	}

	savedChecksum := saved.Checksum
	actualChecksum := Checksum(actualText)

	if savedChecksum != actualChecksum {
		// Show a useful diff summary.
		savedLines := strings.Split(saved.Text, "\n")
		actualLines := strings.Split(actualText, "\n")

		if len(savedLines) == 1 && len(actualLines) == 1 {
			// Single-line: show both.
			maxShow := 200
			savedShow := saved.Text
			actualShow := actualText
			if len(savedShow) > maxShow {
				savedShow = savedShow[:maxShow] + "..."
			}
			if len(actualShow) > maxShow {
				actualShow = actualShow[:maxShow] + "..."
			}
			return fmt.Errorf("response changed:\n  saved:  %s\n  actual: %s", savedShow, actualShow)
		}

		return fmt.Errorf("response changed (checksum %s → %s, %d lines → %d lines)",
			savedChecksum, actualChecksum, len(savedLines), len(actualLines))
	}

	return nil
}

// PrintSnapshotSummary prints what happened during a snapshot run.
func PrintSnapshotSummary(updated, matched, changed, newSnaps int) {
	if updated > 0 {
		fmt.Printf("\nSnapshots updated: %d new, %d changed\n", newSnaps, changed)
	} else {
		fmt.Printf("\nSnapshots: %d matched", matched)
		if changed > 0 {
			fmt.Printf(", %d changed (run with --update to accept)", changed)
		}
		fmt.Println()
	}
}
