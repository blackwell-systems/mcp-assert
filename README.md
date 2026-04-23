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
# evals/get_references.yaml
name: get_references returns cross-file callers
server:
  command: agent-lsp
  args: ["go:gopls"]
setup:
  - tool: start_lsp
    args: { root_dir: "{{fixture}}" }
  - tool: open_document
    args: { file_path: "{{fixture}}/greeter.go", language_id: go }
assert:
  tool: get_references
  args:
    file_path: "{{fixture}}/main.go"
    line: 6
    column: 6
  expect:
    contains: ["greeter.go"]
    min_results: 2
```

Run it:

```bash
mcp-assert run --suite evals/ --fixture test/fixtures/go
```

Output:

```
PASS  get_references returns cross-file callers (go)        142ms
PASS  go_to_definition resolves to correct file (go)         38ms
FAIL  rename_symbol updates all call sites (go)             412ms
      expected 7 changed files, got 5

3 assertions, 2 passed, 1 failed
```

## Cross-Language Matrix

Run the same assertions across multiple language servers:

```bash
mcp-assert matrix \
  --suite evals/ \
  --languages go:gopls,typescript:typescript-language-server,python:pyright-langserver
```

```
                     get_references  go_to_definition  rename_symbol  completions
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
```

GitHub Action:

```yaml
- name: Assert MCP server correctness
  run: |
    go install github.com/blackwell-systems/mcp-assert@latest
    mcp-assert ci --suite evals/ --threshold 95
```

## How It Differs

| Dimension | Existing MCP evals | mcp-assert |
|---|---|---|
| Grading | LLM-as-judge (subjective, costly) | Deterministic assertions (exact, free) |
| Speed | Seconds per test (LLM round-trip) | Milliseconds per test (no LLM) |
| CI cost | API calls on every run | Zero external dependencies |
| Reliability | Varies by model temperature | Deterministic pass/fail |
| Multi-language | Not supported | Same assertion across N language servers |
| Docker isolation | Not supported | Per-trial container execution |

## Assertion Types

| Assertion | What it checks |
|---|---|
| `contains` | Response text contains all specified strings |
| `equals` | Response exactly matches expected value |
| `json_path` | JSON field at path matches expected value |
| `min_results` | Array field has at least N items |
| `max_results` | Array field has at most N items |
| `not_empty` | Response is non-empty and not `null`/`[]`/`{}` |
| `not_error` | Tool response has `isError: false` |
| `file_contains` | After tool execution, file on disk contains text |
| `file_unchanged` | File on disk was not modified |
| `net_delta` | Speculative execution diagnostic delta equals N |
| `in_order` | Tool calls in transcript appear in specified order |

## Assertion File Format

```yaml
name: Human-readable description
server:
  command: path/to/mcp-server
  args: ["arg1", "arg2"]
  env:
    KEY: value
setup:
  - tool: tool_name
    args: { key: value }
assert:
  tool: tool_name
  args: { key: value }
  expect:
    contains: ["expected", "strings"]
    json_path:
      "$.locations[0].file": "main.go"
    min_results: 3
    not_error: true
timeout: 30s
```

## Docker Isolation

Run assertions in isolated containers for reproducibility:

```bash
mcp-assert run --suite evals/ --docker ghcr.io/blackwell-systems/agent-lsp:go
```

Each assertion runs in a fresh container. No cross-test contamination. Same environment as CI.

## Reliability Metrics

Run multiple trials to measure consistency:

```bash
mcp-assert run --suite evals/ --trials 5
```

Reports `pass@k` (capability: passed at least once) and `pass^k` (reliability: passed every time).

## License

MIT
