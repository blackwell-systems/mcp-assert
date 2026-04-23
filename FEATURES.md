# mcp-assert Features

Machine-readable feature inventory. Dense structured lists for AI analysis and capability discovery.

---

## CLI Commands (8)

| Command | Description | Key flags |
|---------|-------------|-----------|
| `init` | Scaffold an assertion template and fixture directory | `[dir]` |
| `run` | Execute assertions against an MCP server | `--suite`, `--server`, `--fixture`, `--trials`, `--docker`, `--json`, `--junit`, `--markdown`, `--badge`, `--baseline`, `--save-baseline` |
| `ci` | Run with CI-specific exit codes and reporting | All `run` flags + `--threshold`, `--fail-on-regression` |
| `matrix` | Run assertions across multiple language servers | `--suite`, `--languages`, `--fixture` |
| `coverage` | Report which server tools have assertions | `--suite`, `--server`, `--coverage-json` |
| `generate` | Auto-generate stub assertions from a server's tools/list | `--server`, `--output`, `--fixture` |
| `snapshot` | Capture/compare tool response snapshots | `--suite`, `--server`, `--fixture`, `--update`, `--docker` |
| `watch` | Rerun assertions on YAML file change | Same as `run` + polling interval |

---

## Assertion Types (14)

| Type | Category | What it checks |
|------|----------|----------------|
| `contains` | Text | Response contains all specified substrings |
| `not_contains` | Text | Response does not contain any specified substrings |
| `equals` | Text | Response exactly matches expected value (whitespace-trimmed) |
| `matches_regex` | Text | Response matches all specified regex patterns |
| `json_path` | Structure | JSON field at `$.dot.path[N]` matches expected value |
| `min_results` | Structure | Array result has at least N items |
| `max_results` | Structure | Array result has at most N items |
| `not_empty` | Presence | Response is non-empty and not `null`/`[]`/`{}` |
| `not_error` | Status | Tool response has `isError: false` |
| `is_error` | Status | Tool response has `isError: true` (negative testing) |
| `file_contains` | Side effect | File on disk contains expected text after tool execution |
| `file_unchanged` | Side effect | File on disk was not modified (snapshot comparison) |
| `net_delta` | Speculative | Diagnostic delta equals expected value |
| `in_order` | Sequence | Substrings appear in specified order within response |

---

## Output Formats (7)

| Format | Flag | Description |
|--------|------|-------------|
| Terminal | (default) | Color pass/fail/skip with duration, progress counter on stderr |
| JSON | `--json` | Full result array as JSON to stdout |
| JUnit XML | `--junit <path>` | Standard JUnit format for CI test result tabs. Includes pass@k/pass^k properties when `--trials > 1` |
| Markdown | `--markdown <path>` | GitHub Step Summary table (auto-detects `$GITHUB_STEP_SUMMARY`). Includes reliability section when `--trials > 1` |
| Badge | `--badge <path>` | shields.io endpoint JSON (`schemaVersion`, `label`, `message`, `color`) |
| Coverage JSON | `--coverage-json <path>` | Machine-readable coverage data: total, covered, percentage, covered/uncovered tool lists |
| Snapshots | `.snapshots.json` | Captured tool responses for regression comparison via `snapshot` command |

---

## Reliability Metrics

When `--trials N` is used (N > 1):

| Metric | Definition |
|--------|------------|
| pass@k | Passed at least once in k trials (capability) |
| pass^k | Passed every time in k trials (reliability) |
| Rate | Pass count / trial count per assertion |

---

## Regression Detection

| Flag | Description |
|------|-------------|
| `--save-baseline <path>` | Persist current results as baseline JSON |
| `--baseline <path>` | Compare current run against saved baseline |
| `--fail-on-regression` | Exit 1 if a previously-passing assertion now fails (requires `--baseline`) |

Only PASS → non-PASS transitions are flagged. Previously-failing tests that still fail are not regressions. New tests not in baseline are not regressions.

---

## Transport Support

| Transport | Field | Description |
|-----------|-------|-------------|
| `stdio` (default) | `command`, `args`, `env` | Launch MCP server as a subprocess, communicate over stdin/stdout |
| `sse` | `url` | Connect to an SSE-based MCP server (legacy transport) |
| `http` | `url` | Connect to a streamable HTTP MCP server (modern transport) |

Transport is configured per-assertion in YAML via the `transport` and `url` fields. When omitted, defaults to stdio. Case-insensitive. Docker isolation is only supported with stdio.

---

## Docker Isolation

`--docker <image>` wraps the MCP server command in `docker run --rm -i` (stdio transport only):
- Fixture directory volume-mounted into the container
- Environment variables forwarded via `-e` flags
- Each assertion gets a fresh container (no cross-test contamination)
- stdio piping for MCP transport — no port mapping needed

---

## Coverage Analysis

`mcp-assert coverage --suite <dir> --server <cmd>`:
- Starts the MCP server and calls `tools/list`
- Compares server tool names against assertion tool names in the suite
- Reports: total tools, covered count, coverage percentage
- Lists each tool as covered (with assertion count) or uncovered

---

## Terminal Behavior

| Feature | TTY | Pipe/CI |
|---------|-----|---------|
| Pass icon | green `✓` | `PASS` |
| Fail icon | red `✗` | `FAIL` |
| Skip icon | yellow `○` | `SKIP` |
| Progress | `[1/21] assertion name` on stderr | disabled |
| Summary | colored counts, non-zero only | plain counts |
| Override | `NO_COLOR=1` forces plain output | n/a |

---

## Example Suites (4 servers, 3 languages, 85 assertions)

| Suite | Server | Language | Assertions | Key patterns |
|-------|--------|----------|------------|--------------|
| `examples/filesystem/` | `@modelcontextprotocol/server-filesystem` | TypeScript | 14 | Read, list, search, info, write, edit, create dir, move, directory tree, path traversal rejection (92% tool coverage) |
| `examples/memory/` | `@modelcontextprotocol/server-memory` | TypeScript | 5 | Stateful setup (create → query), relations, observations |
| `examples/sqlite/` | `mcp-server-sqlite` | Python | 6 | SQL queries, joins, counts, schema introspection, error handling |
| `examples/agent-lsp-go/` | agent-lsp + gopls | Go | 60 | All 50 tools: navigation, refactoring, analysis, session lifecycle, workspace, build (100% tool coverage) |

---

## CI Pipeline (5 jobs)

| Job | What | Depends on |
|-----|------|------------|
| `build-and-test` | Build, vet, 100 unit tests with `-race` | — |
| `e2e-filesystem` | 14 assertions against filesystem server | build-and-test |
| `e2e-memory` | 5 assertions against memory server | build-and-test |
| `e2e-sqlite` | 6 assertions against SQLite server (Python/uv) | build-and-test |
| `e2e-agent-lsp` | 60 assertions against agent-lsp + gopls | build-and-test |

All e2e jobs upload JUnit XML artifacts.

---

## Unit Test Coverage

| Package | Tests | What |
|---------|-------|------|
| `internal/assertion` | 22 | All 14 assertion types, loader (YAML parsing, subdirs, errors), snapshot comparison |
| `internal/report` | 36 | PrintResults, PrintMatrix, JUnit XML (with pass@k), markdown (with reliability), badge JSON, reliability metrics, baseline write/load, regression detection, coverage JSON, snapshot save/load/compare |
| `internal/runner` | 53 | Recursive fixture substitution, capture/extractJSONPath, server override, bad binary, timeout, Docker flag, transport selection (stdio/SSE/HTTP), URL validation, generate schema parsing, stub generation, filename sanitization, CLI error paths |
| Total | 111 | Race-detector clean |

---

## YAML Assertion Format

```yaml
name: Human-readable description
server:
  command: path/to/mcp-server        # stdio transport
  args: ["arg1", "arg2"]
  env:
    KEY: value
  transport: stdio                   # "stdio" (default), "sse", or "http"
  url: "http://localhost:8080/sse"   # required for sse/http transport
setup:
  - tool: setup_tool
    args: { key: value }
    capture:
      variable_name: "$.json.path"    # extract from response
assert:
  tool: tool_under_test
  args: { key: value }
  expect:
    not_error: true
    contains: ["expected"]
    json_path:
      "$.field": "value"
    min_results: 3
timeout: 30s
```

`{{fixture}}` in args is replaced with `--fixture` directory at runtime.

---

## Architecture

```
cmd/mcp-assert/main.go     CLI entry, command dispatch
internal/assertion/
  types.go                  Suite, Assertion, Expect, Result types
  loader.go                 YAML file loading, subdirectory recursion
  checker.go                14 assertion type implementations
internal/runner/
  runner.go                 Run, Matrix, CI commands, MCP client lifecycle
  runner_test.go            31 tests: substitution, overrides, error paths, timeout, Docker, generate
  coverage.go               Coverage command, tools/list query, --coverage-json
  generate.go               Auto-generate stub assertions from tools/list
  init.go                   Scaffold assertion template and fixture directory
  snapshot.go               Snapshot capture/compare command
  watch.go                  File-watching rerun loop
internal/report/
  report.go                 Terminal output (color-aware)
  color.go                  ANSI color, TTY detection, progress
  junit.go                  JUnit XML generation (with pass@k properties)
  markdown.go               GitHub Step Summary (with reliability table)
  badge.go                  shields.io endpoint JSON
  reliability.go            pass@k / pass^k computation
  baseline.go               Baseline write/load, regression detection
  coverage.go               Coverage JSON serialization
  snapshot.go               Snapshot file read/write/compare
```
