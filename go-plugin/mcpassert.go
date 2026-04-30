// Package mcpassert provides Go test helpers for running mcp-assert YAML
// assertion files. It shells out to the mcp-assert binary with --json and
// maps results to Go test pass/fail/skip outcomes.
//
// Same YAML files work across Go test, Jest, Vitest, pytest, and the CLI.
//
// Usage:
//
//	func TestMCPServer(t *testing.T) {
//	    mcpassert.Run(t, "evals/echo.yaml")
//	}
//
//	func TestMCPSuite(t *testing.T) {
//	    mcpassert.Suite(t, "evals/")
//	}
package mcpassert

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Result represents a single mcp-assert assertion outcome.
type Result struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Detail     string `json:"detail,omitempty"`
	DurationMS int    `json:"duration_ms,omitempty"`
	Trial      int    `json:"trial,omitempty"`
}

// Options configures how mcp-assert runs.
type Options struct {
	// Binary is the path to the mcp-assert binary. Auto-detected if empty.
	Binary string
	// Timeout is the per-assertion timeout (e.g. "30s", "1m"). Default: "30s".
	Timeout string
	// Fixture is the directory for {{fixture}} substitution.
	Fixture string
	// Server overrides the server command (e.g. "npx my-server").
	Server string
}

// Run executes a single YAML assertion file and reports the result to t.
// On FAIL, the test fails with the assertion detail. On SKIP, the test is skipped.
func Run(t *testing.T, yamlPath string, opts ...Options) {
	t.Helper()

	opt := mergeOpts(opts)
	results := run(t, yamlPath, opt)

	for _, r := range results {
		switch r.Status {
		case "PASS":
			// success
		case "SKIP":
			t.Skipf("%s: %s", r.Name, r.Detail)
		case "FAIL":
			t.Errorf("%s: %s", r.Name, r.Detail)
		default:
			t.Errorf("%s: unexpected status %q", r.Name, r.Status)
		}
	}
}

// Suite discovers all YAML files in a directory and runs each as a subtest.
func Suite(t *testing.T, suiteDir string, opts ...Options) {
	t.Helper()

	entries, err := os.ReadDir(suiteDir)
	if err != nil {
		t.Fatalf("reading suite directory %s: %v", suiteDir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml")) {
			files = append(files, e.Name())
		}
	}

	if len(files) == 0 {
		t.Fatalf("no YAML files found in %s", suiteDir)
	}

	for _, file := range files {
		name := strings.TrimSuffix(strings.TrimSuffix(file, ".yaml"), ".yml")
		yamlPath := filepath.Join(suiteDir, file)

		t.Run(name, func(t *testing.T) {
			Run(t, yamlPath, opts...)
		})
	}
}

func run(t *testing.T, yamlPath string, opt Options) []Result {
	t.Helper()

	binary := findBinary(opt.Binary)
	if binary == "" {
		t.Fatal("mcp-assert binary not found. Install via: " +
			"brew install blackwell-systems/tap/mcp-assert, " +
			"go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest, or " +
			"npm install @blackwell-systems/mcp-assert")
	}

	args := []string{
		"run",
		"--suite", yamlPath,
		"--json",
		"--timeout", opt.Timeout,
	}

	if opt.Fixture != "" {
		args = append(args, "--fixture", opt.Fixture)
	}
	if opt.Server != "" {
		args = append(args, "--server", opt.Server)
	}

	cmd := exec.Command(binary, args...)
	out, err := cmd.Output()

	// mcp-assert exits non-zero on assertion failure but still writes JSON to stdout
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(out) > 0 {
			_ = exitErr // expected on failure
		} else if len(out) == 0 {
			stderr := ""
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr = string(exitErr.Stderr)
			}
			t.Fatalf("mcp-assert failed: %v\n%s", err, stderr)
		}
	}

	var results []Result
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatalf("could not parse mcp-assert output: %v\n%s", err, string(out[:min(len(out), 500)]))
	}

	if len(results) == 0 {
		t.Fatal("mcp-assert returned no results")
	}

	return results
}

func findBinary(explicit string) string {
	if explicit != "" {
		return explicit
	}

	// Check PATH
	if path, err := exec.LookPath("mcp-assert"); err == nil {
		return path
	}

	// Check common Go install location
	if home, err := os.UserHomeDir(); err == nil {
		gopath := filepath.Join(home, "go", "bin", "mcp-assert")
		if _, err := os.Stat(gopath); err == nil {
			return gopath
		}
	}

	return ""
}

func mergeOpts(opts []Options) Options {
	opt := Options{Timeout: "30s"}
	if len(opts) > 0 {
		o := opts[0]
		if o.Binary != "" {
			opt.Binary = o.Binary
		}
		if o.Timeout != "" {
			opt.Timeout = o.Timeout
		}
		if o.Fixture != "" {
			opt.Fixture = o.Fixture
		}
		if o.Server != "" {
			opt.Server = o.Server
		}
	}
	return opt
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
