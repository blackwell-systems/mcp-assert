# mcp-assert

Test any MCP server in any language. No SDK required. No LLM required.

A single Go binary that starts your MCP server over stdio, calls your tools, and asserts the results. Define assertions in YAML, run them in CI. Works with servers written in Go, TypeScript, Python, Rust, Java — anything that speaks MCP.

## Why

Most MCP tools are deterministic: `read_file` returns file contents, `read_query` returns rows, `get_references` returns locations. Given the same input, the correct output is knowable in advance. You don't need an LLM to grade it — you need `assert.Equal`.

Existing MCP eval frameworks use LLM-as-judge for everything: send a prompt, get a response, ask GPT "was this good?" on a 1-5 scale. This adds cost, latency, and false variance to tests that should be instant and exact.

mcp-assert tests MCP server tools the way you test code: given this input, assert this output.

### When to use what

| Your tool returns... | Use |
|---|---|
| Structured data (files, rows, locations, symbols) | **mcp-assert** — deterministic assertions |
| Predictable state changes (rename, create, delete) | **mcp-assert** — assert the state after |
| Error responses for bad input | **mcp-assert** — `is_error` and `contains` |
| Natural language (summaries, explanations, descriptions) | **LLM-as-judge** — quality is subjective |
| Creative content (commit messages, code suggestions) | **LLM-as-judge** — many correct answers |

Most MCP servers are heavy on the first three and light on the last two. If your server returns data, mcp-assert covers it. If your server generates prose, you need a different tool.

## Why not just write tests in Go/Python/etc?

You could. The assertion logic is straightforward. What you'd have to build yourself:

- **MCP protocol bootstrapping** — stdio transport, JSON-RPC framing, initialize/initialized handshake, tool call request/response lifecycle. This is ~200 lines of boilerplate per test suite, and easy to get wrong.
- **Server-agnostic test runner** — your Go tests are coupled to your Go server. mcp-assert tests any server from any language with the same YAML. Switch `server.command` from `npx my-ts-server` to `python -m my_server` and the assertions don't change.
- **Eval-framework features** — pass@k/pass^k reliability metrics, baseline regression detection, JUnit XML output, Docker isolation, cross-language matrix mode. These are eval concerns, not unit test concerns. Go's `testing` package doesn't have opinions about them.

The value isn't in the assertion logic. It's in not writing MCP client boilerplate, having one tool that works across every MCP server regardless of implementation language, and getting CI-grade reporting for free.

## Quick Start

```bash
go install github.com/blackwell-systems/mcp-assert@latest
```

Define an assertion for any MCP server — here's one for the TypeScript filesystem server:

```yaml
# evals/read_file.yaml
name: read_file returns file contents
server:
  command: npx
  args: ["@modelcontextprotocol/server-filesystem", "{{fixture}}"]
assert:
  tool: read_file
  args:
    path: "{{fixture}}/hello.txt"
  expect:
    not_error: true
    contains: ["Hello, world!"]
```

Run it:

```bash
mcp-assert run --suite evals/ --fixture ./fixtures
```

Works the same for a Go server, a Python server, or anything else that speaks MCP:

```yaml
# Same assertion format, different server
server:
  command: python
  args: ["-m", "my_mcp_server"]
```

```yaml
server:
  command: agent-lsp
  args: ["go:gopls"]
```

## Example Suites

mcp-assert ships with example assertions for four MCP servers in three languages:

### Filesystem server — TypeScript (`examples/filesystem/`)

Tests the official `@modelcontextprotocol/server-filesystem`. 5 assertions: read file, list directory, get file info, search files, and a **negative test** that verifies path traversal is rejected.

```bash
npm install -g @modelcontextprotocol/server-filesystem
mcp-assert run --suite examples/filesystem --fixture examples/filesystem/fixtures
```

### Memory server — TypeScript (`examples/memory/`)

Tests the official `@modelcontextprotocol/server-memory`. 5 assertions with **stateful setup**: create entities, add observations, create relations, search nodes, and verify empty search returns nothing.

```bash
npm install -g @modelcontextprotocol/server-memory
mcp-assert run --suite examples/memory
```

### SQLite server — Python (`examples/sqlite/`)

Tests the official `mcp-server-sqlite` (Python). 6 assertions: list tables, SELECT queries, COUNT, JOINs, describe table schema, and error handling for invalid SQL. Fixture is a pre-built `.db` file.

```bash
uvx mcp-server-sqlite  # or: pip install mcp-server-sqlite
mcp-assert run --suite examples/sqlite --fixture examples/sqlite/fixtures
```

### agent-lsp — Go (`examples/agent-lsp-go/`)

Tests [agent-lsp](https://github.com/blackwell-systems/agent-lsp) with gopls. 21 assertions covering navigation, refactoring, analysis, and build tools.

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
| Reliability | Not measured | pass@k / pass^k per assertion |
| Regression | Not supported | Baseline comparison, fail on backslide |
| Docker | Not supported | Per-assertion container isolation |
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

## Docker Isolation

Run each assertion in a fresh Docker container for reproducibility:

```bash
mcp-assert run --suite evals/ --docker ghcr.io/blackwell-systems/agent-lsp:go --fixture /workspace
```

The fixture directory is mounted into the container. Each assertion gets a clean environment — no cross-test contamination, no "works on my machine."

## Reliability Metrics

Run multiple trials to measure consistency:

```bash
mcp-assert run --suite evals/ --trials 5
```

```
PASS  hover returns type info                 690ms
PASS  hover returns type info                 650ms
PASS  hover returns type info                 710ms
FAIL  get_references finds cross-file callers 90001ms
      tool call get_references failed: context deadline exceeded
PASS  get_references finds cross-file callers 27305ms

Reliability:
  Assertion                                     Trials  Passed    pass@k  pass^k
  ------------------------------------------    ------  ------  --------  ------
  hover returns type info                            3       3       YES     YES
  get_references finds cross-file callers            2       1       YES      NO

  pass@k: 2/2 capable, pass^k: 1/2 reliable
```

- **pass@k** (capability): Did the assertion pass at least once? If NO, the tool is broken.
- **pass^k** (reliability): Did the assertion pass every time? If NO, the tool is flaky.

## Regression Detection

Save a baseline, then detect regressions on future runs:

```bash
# Save current results as baseline
mcp-assert run --suite evals/ --save-baseline baseline.json

# Later: compare against baseline
mcp-assert ci --suite evals/ --baseline baseline.json --fail-on-regression
```

```
Regressions detected (1):
  get_references finds cross-file callers: was PASS, now FAIL
error: 1 regression(s) detected
```

Only flags transitions from PASS to FAIL. Previously-failing tests that still fail are not regressions. New tests that fail are not regressions.

## Coverage

See which server tools have assertions and which don't:

```bash
mcp-assert coverage --suite evals/ --server "agent-lsp go:gopls"
```

```
Server exposes 50 tools, 21 have assertions (42% coverage)

Covered (21):
  ✓ call_hierarchy (1 assertion)
  ✓ format_document (1 assertion)
  ✓ get_references (1 assertion)
  ...

Not covered (29):
  ○ add_workspace_folder
  ○ apply_edit
  ○ close_document
  ...
```

The command queries the server's `tools/list` endpoint, compares against assertion tool names in the suite, and reports coverage percentage with covered/uncovered tool lists.

## Terminal Output

mcp-assert uses color in interactive terminals: green for pass, red for fail, yellow for skip. A progress counter (`[1/21]`, `[2/21]`, ...) prints to stderr while assertions run. The summary line only shows non-zero counts.

Color and progress are automatically disabled in pipes and CI environments. Set `NO_COLOR=1` to force plain `PASS`/`FAIL`/`SKIP` output explicitly.

## License

MIT
