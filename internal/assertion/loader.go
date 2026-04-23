package assertion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadSuite reads all YAML assertion files from a directory.
func LoadSuite(dir string) (*Suite, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading suite dir %s: %w", dir, err)
	}

	suite := &Suite{Dir: dir}
	for _, entry := range entries {
		if entry.IsDir() {
			// Recurse one level into subdirectories.
			subDir := filepath.Join(dir, entry.Name())
			subEntries, err := os.ReadDir(subDir)
			if err != nil {
				continue
			}
			for _, sub := range subEntries {
				if isYAML(sub.Name()) {
					a, err := loadFile(filepath.Join(subDir, sub.Name()))
					if err != nil {
						return nil, err
					}
					suite.Assertions = append(suite.Assertions, *a)
				}
			}
			continue
		}
		if isYAML(entry.Name()) {
			a, err := loadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			suite.Assertions = append(suite.Assertions, *a)
		}
	}

	if len(suite.Assertions) == 0 {
		return nil, fmt.Errorf("no assertion files found in %s", dir)
	}
	return suite, nil
}

func loadFile(path string) (*Assertion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var a Assertion
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if a.Name == "" {
		a.Name = filepath.Base(path)
	}
	return &a, nil
}

func isYAML(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
