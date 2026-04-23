package runner

import (
	"fmt"
	"os"
	"path/filepath"
)

const assertionTemplate = `# Name shown in test output. Make it descriptive — it's the only context
# you get when a test fails in CI.
name: my_tool returns expected output

# Server config: how to start the MCP server under test.
# mcp-assert launches this process, connects over stdio, and runs the
# MCP initialize handshake automatically.
server:
  command: npx                                    # The binary to run
  args: ["@modelcontextprotocol/server-filesystem", "{{fixture}}"]  # CLI arguments
  # env:                                          # Optional environment variables
  #   API_KEY: test-key

# Setup steps run before the assertion, in order. Use them to create state
# the assertion depends on (e.g. create a file before reading it).
# Each step is a tool call. If any setup step fails, the assertion fails.
# setup:
#   - tool: write_file
#     args:
#       path: "{{fixture}}/test.txt"
#       content: "setup data"

# The tool call to test and its expected results.
assert:
  tool: read_file                                 # MCP tool name
  args:                                           # Arguments passed to the tool
    path: "{{fixture}}/hello.txt"
  expect:
    # --- Error checks ---
    not_error: true                               # Tool response has isError: false
    # is_error: true                              # Tool response has isError: true (negative tests)

    # --- Content checks ---
    not_empty: true                               # Response is non-empty
    contains: ["Hello, world!"]                   # Response contains all listed strings
    # not_contains: ["secret"]                    # Response does NOT contain any listed strings
    # equals: "exact match"                       # Response exactly matches (whitespace-trimmed)
    # matches_regex: ["\\d+ items"]               # Response matches all regex patterns
    # in_order: ["first", "second", "third"]      # Substrings appear in this order

    # --- JSON checks ---
    # json_path:                                  # Assert on JSON fields using dot-path
    #   "$.name": "Alice"
    #   "$.items[0].id": 1

    # --- Array checks ---
    # min_results: 3                              # Array result has at least N items
    # max_results: 10                             # Array result has at most N items

    # --- File system checks (after tool runs) ---
    # file_contains:                              # File on disk contains expected text
    #   "{{fixture}}/output.txt": "expected content"
    # file_unchanged: ["{{fixture}}/readonly.txt"]  # File was NOT modified

# Per-assertion timeout. The MCP server is killed if it doesn't respond in time.
# Default: 30s. Increase for slow servers or complex operations.
timeout: 15s
`

const fixtureContent = `Hello, world!
`

// Init scaffolds an assertion YAML template and fixture directory.
func Init(args []string) error {
	dir := "evals"
	if len(args) > 0 && args[0] != "" {
		dir = args[0]
	}

	// Create assertion directory.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Write template assertion.
	assertPath := filepath.Join(dir, "read_file.yaml")
	if _, err := os.Stat(assertPath); err == nil {
		return fmt.Errorf("%s already exists — not overwriting", assertPath)
	}
	if err := os.WriteFile(assertPath, []byte(assertionTemplate), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", assertPath, err)
	}

	// Create fixtures directory with sample file.
	fixtureDir := filepath.Join(dir, "fixtures")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		return fmt.Errorf("creating fixtures: %w", err)
	}
	helloPath := filepath.Join(fixtureDir, "hello.txt")
	if _, err := os.Stat(helloPath); err != nil {
		if err := os.WriteFile(helloPath, []byte(fixtureContent), 0o644); err != nil {
			return fmt.Errorf("writing fixture: %w", err)
		}
	}

	fmt.Printf("Created %s with:\n", dir)
	fmt.Printf("  %s          — assertion template (edit this)\n", assertPath)
	fmt.Printf("  %s  — fixture file for {{fixture}} substitution\n", helloPath)
	fmt.Println()
	fmt.Println("Run it:")
	fmt.Printf("  mcp-assert run --suite %s --fixture %s\n", dir, fixtureDir)
	return nil
}
