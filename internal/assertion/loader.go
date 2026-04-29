// loader.go discovers and parses assertion YAML files into a Suite.
//
// Two loading modes are supported:
//   - Single file: path points directly to a .yaml/.yml file.
//   - Directory: path points to a directory. All YAML files in the directory
//     are loaded, plus YAML files one subdirectory deep. Deeper nesting is
//     intentionally ignored to keep suite layout predictable.
//
// Each YAML file is parsed into exactly one Assertion struct. If the file
// does not include a "name" field, the filename is used as the assertion name.
package assertion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadSuite reads YAML assertion files from a directory or a single YAML file.
// When given a file path, it loads only that one assertion.
// When given a directory, it loads all YAML files (recursing one level into subdirectories).
func LoadSuite(path string) (*Suite, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("reading suite path %s: %w", path, err)
	}

	// Single file mode: load one YAML file directly.
	if !info.IsDir() {
		if !isYAML(info.Name()) {
			return nil, fmt.Errorf("%s is not a YAML file", path)
		}
		a, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		return &Suite{
			Assertions: []Assertion{*a},
			Dir:        filepath.Dir(path),
		}, nil
	}

	// Directory mode: existing behavior.
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading suite dir %s: %w", path, err)
	}

	suite := &Suite{Dir: path}
	for _, entry := range entries {
		if entry.IsDir() {
			// Recurse one level into subdirectories.
			subDir := filepath.Join(path, entry.Name())
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
			a, err := loadFile(filepath.Join(path, entry.Name()))
			if err != nil {
				return nil, err
			}
			suite.Assertions = append(suite.Assertions, *a)
		}
	}

	if len(suite.Assertions) == 0 {
		return nil, fmt.Errorf("no assertion files found in %s", path)
	}
	return suite, nil
}

// loadFile reads a single YAML file and unmarshals it into an Assertion.
// Falls back to using the filename as the assertion name when "name" is omitted.
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

// isYAML returns true for filenames ending in .yaml or .yml.
func isYAML(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
