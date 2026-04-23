# mcp-assert

Deterministic correctness testing for MCP servers. Assert your tools return the right results, not just any results.

## Why

Every existing MCP eval framework uses LLM-as-judge: send a prompt, get a response, ask GPT "was this good?" on a 1-5 scale. This makes sense for subjective outputs. It's the wrong approach for deterministic tools.

When `get_references` is called on line 42 of a Go file, the correct answer is a specific set of locations. The tool either returns them or it doesn't. No LLM needed. No API costs. No false variance.

mcp-assert tests MCP server tools the way you test code: given this input, assert this output.

## Quick Start

```bash
go install github.com/blackwell-systems/mcp-assert@latest
```

Define an assertion:

```yaml
# evals/hover.yaml
name: hover returns type info for Person
server:
  command: agent-lsp
  args: ["go:gopls"]
setup:
  - tool: start_lsp
    args: { root_dir: "{{fixture}}" }
  - tool: open_document
    args: { file_path: "{{fixture}}/main.go", language_id: go }
assert:
  tool: get_info_on_location
  args:
    file_path: "{{fixture}}/main.go"
    line: 6
    column: 6
  expect:
    not_error: true
    not_empty: true
    contains: ["Person"]
```

Run it:

```bash
mcp-assert run --suite evals/ --fixture test/fixtures/go
```

Output:

```
PASS  hover returns type info for Person                              690ms
PASS  go_to_definition resolves Person to main.go                     652ms
PASS  get_diagnostics returns clean for valid file                  25097ms
PASS  get_references finds cross-file callers of Person             27358ms
PASS  speculative edit detects type error with net_delta > 0        28676ms
PASS  completions suggest methods after Person dot                    699ms
PASS  get_document_symbols lists Person type and methods              415ms

7 assertions, 7 passed, 0 failed, 0 skipped
```

## Example Suites

mcp-assert ships with example assertions for three MCP servers:

### Filesystem server (`examples/filesystem/`)

Tests the official `@modelcontextprotocol/server-filesystem`. 5 assertions: read file, list directory, get file info, search files, and a **negative test** that verifies path traversal is rejected.

```bash
npm install -g @modelcontextprotocol/server-filesystem
mcp-assert run --suite examples/filesystem --fixture examples/filesystem/fixtures
```

### Memory server (`examples/memory/`)

Tests the official `@modelcontextprotocol/server-memory`. 5 assertions with **stateful setup**: create entities, add observations, create relations, search nodes, and verify empty search returns nothing.

```bash
npm install -g @modelcontextprotocol/server-memory
mcp-assert run --suite examples/memory
```

### agent-lsp (`examples/agent-lsp-go/`)

Tests [agent-lsp](https://github.com/blackwell-systems/agent-lsp) with gopls. 7 assertions: hover, definition, references, diagnostics, symbols, completions, and speculative execution.

```bash
mcp-assert run --suite examples/agent-lsp-go --fixture /path/to/go/fixtures
```

## Server Override

Override the server config from CLI instead of repeating it in every YAML file:

```bash
mcp-assert run --suite evals/ --server "agent-lsp go:gopls" --fixture test/fixtures/go
```

## Cross-Language Matrix

Run the same assertions across multiple language servers:

```bash
mcp-assert matrix \
  --suite evals/ \
  --languages go:gopls,typescript:typescript-language-server,python:pyright-langserver
```

```
                     hover           definition        references     completions
Go (gopls)           PASS            PASS              PASS           PASS
TypeScript (tsserver) PASS            PASS              PASS           PASS
Python (pyright)     PASS            PASS              SKIP           PASS
```

## CI Integration

```bash
# Fail the build if any assertion regresses
mcp-assert ci --suite evals/ --fail-on-regression

# Set a minimum pass threshold
mcp-assert ci --suite evals/ --threshold 95

# Override server from CLI
mcp-assert ci --suite evals/ --server "my-mcp-server" --threshold 100
```

GitHub Action:

```yaml
- name: Assert MCP server correctness
  run: |
    go install github.com/blackwell-systems/mcp-assert@latest
    mcp-assert ci --suite evals/ --threshold 95 --junit results.xml

- name: Upload test results
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: mcp-assert-results
    path: results.xml
```

## Structured Reporting

```bash
# JUnit XML for CI test result tabs (GitHub Actions, Jenkins, CircleCI)
mcp-assert run --suite evals/ --junit results.xml

# GitHub Step Summary (auto-detects $GITHUB_STEP_SUMMARY in ci mode)
mcp-assert ci --suite evals/ --markdown summary.md

# shields.io badge endpoint
mcp-assert run --suite evals/ --badge badge.json
# Then use: ![mcp-assert](https://img.shields.io/endpoint?url=<badge-url>)
```

## How It Differs

| Dimension | Existing MCP evals | mcp-assert |
|---|---|---|
| Grading | LLM-as-judge (subjective, costly) | Deterministic assertions (exact, free) |
| Speed | Seconds per test (LLM round-trip) | Milliseconds per test (no LLM) |
| CI cost | API calls on every run | Zero external dependencies |
| Reliability | Varies by model temperature | Deterministic pass/fail |
| Multi-language | Not supported | Same assertion across N language servers |

## Assertion Types

| Assertion | What it checks |
|---|---|
| `contains` | Response text contains all specified strings |
| `not_contains` | Response text does not contain any of the specified strings |
| `equals` | Response exactly matches expected value (whitespace-trimmed) |
| `matches_regex` | Response matches all specified regex patterns |
| `json_path` | JSON field at `$.dot.path` matches expected value |
| `min_results` | Array result has at least N items |
| `max_results` | Array result has at most N items |
| `not_empty` | Response is non-empty and not `null`/`[]`/`{}` |
| `not_error` | Tool response has `isError: false` |
| `is_error` | Tool response has `isError: true` (for negative testing) |
| `file_contains` | After tool execution, file on disk contains expected text |
| `file_unchanged` | File on disk was not modified by the tool |
| `net_delta` | Speculative execution diagnostic delta equals N |
| `in_order` | Substrings appear in the specified order within the response |

## Assertion File Format

```yaml
name: Human-readable description
server:
  command: path/to/mcp-server
  args: ["arg1", "arg2"]
  env:
    KEY: value
setup:
  - tool: setup_tool
    args: { key: value }
  - tool: another_setup_tool
    args: { key: value }
assert:
  tool: tool_under_test
  args: { key: value }
  expect:
    not_error: true
    contains: ["expected", "strings"]
    not_contains: ["unexpected"]
    matches_regex: ["\\d+ items"]
    json_path:
      "$.locations[0].file": "main.go"
    min_results: 3
timeout: 30s
```

The `{{fixture}}` placeholder in args is replaced with the `--fixture` directory at runtime.

## Reliability Metrics

Run multiple trials to measure consistency:

```bash
mcp-assert run --suite evals/ --trials 5
```

## License

MIT
