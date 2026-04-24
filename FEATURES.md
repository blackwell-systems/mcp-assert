# mcp-assert Features

Machine-readable feature inventory. Dense structured lists for AI analysis and capability discovery.

---

## CLI Commands (9)

| Command | Description | Key flags |
|---------|-------------|-----------|
| `init` | Scaffold template or one-step suite generation from a live server | `[dir]`, `--server`, `--fixture`, `--timeout` |
| `run` | Execute assertions against an MCP server | `--suite` (dir or file), `--server`, `--fixture`, `--trials`, `--docker`, `--json`, `--junit`, `--markdown`, `--badge`, `--baseline`, `--save-baseline`, `--fix` |
| `ci` | Run with CI-specific exit codes and reporting | All `run` flags + `--threshold`, `--fail-on-regression`, `--fix` |
| `matrix` | Run assertions across multiple language servers | `--suite`, `--languages`, `--fixture` |
| `coverage` | Report which server tools have assertions | `--suite`, `--server`, `--coverage-json` |
| `generate` | Auto-generate stub assertions from a server's tools/list (destructive tools skipped by default) | `--server`, `--output`, `--fixture`, `--include-writes` |
| `snapshot` | Capture/compare tool response snapshots | `--suite`, `--server`, `--fixture`, `--update`, `--docker` |
| `watch` | Rerun assertions on YAML file change; shows unified diff when assertion status flips | Same as `run` + polling interval |
| `intercept` | Proxy stdio between agent and MCP server, capturing tool calls for live trajectory validation | `--server`, `--trajectory`, `--timeout` |

---

## Assertion Types (15 + 4 trajectory)

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
| `min_progress` | Progress | At least N `notifications/progress` received during tool execution (requires `capture_progress: true` on the assert block) |

## Trajectory Assertion Types (4)

Used with `trajectory:` field to validate tool call sequences. Source is either `trace:` (inline YAML) or `audit_log:` (JSONL file). Runs without a server.

| Type | What it checks |
|------|----------------|
| `order` | Listed tools appear in this sequence (not necessarily adjacent) |
| `presence` | All listed tools appear at least once |
| `absence` | None of the listed tools appear |
| `args_contain` | A tool was called with specific argument values (partial match) |

---

## Assertion Block Types (7)

| Block | What it tests | Key fields |
|-------|---------------|------------|
| `assert:` | Tool call via `tools/call` | `tool`, `args`, `expect`, `capture_progress` |
| `assert_prompts:` | `prompts/list` or `prompts/get` | `list: {}`, `get: {name, arguments}`, `expect` |
| `assert_resources:` | `resources/list`, `resources/read`, subscribe/unsubscribe | `list: {}`, `read: "uri"`, `subscribe`, `unsubscribe`, `expect` |
| `assert_completion:` | `completion/complete` for argument autocompletion | `ref: {type, name}`, `argument: {name, value}`, `expect` |
| `assert_sampling:` | Tool call that triggers `sampling/createMessage` | `tool`, `args`, `mock_text`, `mock_model`, `expect` |
| `assert_logging:` | `logging/setLevel` + `notifications/message` capture | `set_level`, `tool`, `args`, `expect: {min_messages, contains_level, contains_data}` |
| `trajectory:` | Tool call sequence validation (no server) | `trace:` or `audit_log:` source, `trajectory:` checks |

Each YAML file uses exactly one block type. The `assert:` block is the default for testing tool calls. The other block types test specific MCP protocol areas.

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
| `stdio` (default) | `command`, `args`, `env` | Launch MCP server as a subprocess, communicate over stdin/stdout. `env` values support `${VAR}` and `$VAR` expansion from the parent shell environment. |
| `sse` | `url` | Connect to an SSE-based MCP server (legacy transport) |
| `http` | `url` | Connect to a streamable HTTP MCP server (modern transport) |

Transport is configured per-assertion in YAML via the `transport` and `url` fields. When omitted, defaults to stdio. Case-insensitive. Docker isolation is only supported with stdio.

---

## Client Capabilities (Bidirectional MCP)

MCP is bidirectional: servers can request things from the client (roots, sampling, elicitation). Set `client_capabilities` in the server block to make mcp-assert respond to these requests:

| Field | Type | Description |
|-------|------|-------------|
| `roots` | `[]string` | File/directory paths returned for `roots/list` requests. `{{fixture}}` is substituted. |
| `sampling` | object | Mock LLM response for `sampling/createMessage` requests. |
| `sampling.text` | `string` | Response text content. |
| `sampling.model` | `string` | Model name to report (default: `"mock"`). |
| `sampling.stop_reason` | `string` | Stop reason (default: `"end_turn"`). |
| `elicitation` | object | Preset values for `elicitation/create` requests. Set `content: {...}` for the response fields, or set fields directly (used as the whole content). |

`client_capabilities` is a YAML-level feature; there is no CLI flag equivalent. Applies to stdio transport only.

---

## Fixture Isolation

Each assertion automatically receives its own copy of the fixture directory (via a temporary directory). The original fixture is never modified, so assertions that write files, apply edits, or commit changes cannot affect subsequent assertions. The temp copy is cleaned up after each assertion finishes, regardless of pass or fail.

Fixture isolation is automatic for stdio transport. Docker mode already isolates via fresh containers, so the copy is skipped when `--docker` is used.

---

## Docker Isolation

`--docker <image>` wraps the MCP server command in `docker run --rm -i` (stdio transport only):
- Fixture directory volume-mounted into the container
- Environment variables forwarded via `-e` flags
- Each assertion gets a fresh container (no cross-test contamination)
- stdio piping for MCP transport, no port mapping needed

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

## Example Suites (18 suites, 3 languages, 174 assertions)

| Suite | Server | Language | Transport | Assertions | Key patterns |
|-------|--------|----------|-----------|------------|--------------|
| `examples/filesystem/` | `@modelcontextprotocol/server-filesystem` | TypeScript | stdio | 14 | Read, list, search, info, write, edit, create dir, move, directory tree, path traversal rejection (92% tool coverage) |
| `examples/memory/` | `@modelcontextprotocol/server-memory` | TypeScript | stdio | 5 | Stateful setup (create → query), relations, observations |
| `examples/sqlite/` | `mcp-server-sqlite` | Python | stdio | 6 | SQL queries, joins, counts, schema introspection, error handling |
| `examples/agent-lsp-go/` | agent-lsp + gopls | Go | stdio | 63 | All 50 tools: navigation, refactoring, analysis, session lifecycle, workspace, build (100% tool coverage) |
| `examples/mcp-go-everything/` | mark3labs/mcp-go everything | Go | stdio | 9 | echo, add, image, resource link, notification, long-running operation (100% tool coverage) |
| `examples/mcp-go-typed-tools/` | mark3labs/mcp-go typed_tools | Go | stdio | 3 | Typed greeting with required/optional params, error case |
| `examples/mcp-go-structured/` | mark3labs/mcp-go structured | Go | stdio | 6 | Weather, user profile, assets, manual structured result |
| `examples/mcp-go-everything-http/` | mark3labs/mcp-go everything | Go | HTTP | 5 | Same tools as stdio suite, transport conformance test |
| `examples/mcp-go-everything-prompts/` | mark3labs/mcp-go everything | Go | stdio | 4 | `prompts/list`, `prompts/get` (static + with arguments), pagination pattern documentation |
| `examples/mcp-go-everything-resources/` | mark3labs/mcp-go everything | Go | stdio | 4 | `resources/list`, `resources/read`, `resources/subscribe`, `resources/unsubscribe` |
| `examples/mcp-go-roots/` | mark3labs/mcp-go roots_server | Go | stdio | 1 | `roots` tool calls back to client; mcp-assert responds via `client_capabilities.roots` |
| `examples/mcp-go-sampling/` | mark3labs/mcp-go sampling_server | Go | stdio | 3 | `ask_llm` (with/without system prompt), `greet`; mock LLM response via `client_capabilities.sampling` (100% tool coverage) |
| `examples/mcp-go-elicitation/` | mark3labs/mcp-go elicitation | Go | stdio | 4 | `create_project`, `cancel_flow`, `decline_flow`, `validation_constraints`; form-based elicitation via `client_capabilities.elicitation` |
| `examples/mcp-go-everything-completion/` | mark3labs/mcp-go everything | Go | stdio | 3 | `completion/complete` for prompt argument, resource URI, and empty prefix |
| `examples/mcp-go-everything-logging/` | mark3labs/mcp-go everything | Go | stdio | 2 | `logging/setLevel` with level setting and log message capture |
| `examples/fastmcp-testing-demo/` | PrefectHQ/fastmcp testing_demo | Python | stdio | 16 | add, greet, async_multiply: edge cases, defaults, negative tests, missing-arg error (100% tool coverage); resources (list, read static, read parameterized), prompts (list, get with arguments), all three MCP feature categories |
| `examples/github-mcp/` | github/github-mcp-server | Go | stdio | 6 | get_me, search_repositories, get_file_contents, list_issues, search_code, list_branches (read-only subset) |
| `examples/trajectory/` | Inline trace (no server) | N/A | N/A | 20 | All 20 agent-lsp skill protocols: required tool call sequences, safety gates (e.g. get_references before apply_edit), absence checks (e.g. no apply_edit in simulate), order constraints |

---

## Install Methods (6)

| Method | Command |
|--------|---------|
| npm | `npx @blackwell-systems/mcp-assert` |
| pip | `pip install mcp-assert` |
| Go | `go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest` |
| Homebrew | `brew install blackwell-systems/tap/mcp-assert` |
| Scoop | `scoop install mcp-assert` (via `blackwell-systems/scoop-bucket`) |
| curl\|sh | `curl -fsSL https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/install.sh \| sh` |

---

## CI Pipeline (5 jobs)

| Job | What | Depends on |
|-----|------|------------|
| `build-and-test` | Build, vet, 218 unit tests with `-race`, 20 trajectory assertions | - |
| `e2e-filesystem` | 14 assertions against filesystem server | build-and-test |
| `e2e-memory` | 5 assertions against memory server | build-and-test |
| `e2e-sqlite` | 6 assertions against SQLite server (Python/uv) | build-and-test |
| `e2e-agent-lsp` | 63 assertions against agent-lsp + gopls | build-and-test |

All e2e jobs upload JUnit XML artifacts.

---

## Unit Test Coverage

| Package | Tests | What |
|---------|-------|------|
| `internal/assertion` | 53 | All 15 assertion types (including min_progress), loader (YAML parsing, subdirs, errors), snapshot comparison, CheckProgress, completion JSON, logging checker, trajectory checker |
| `internal/report` | 42 | PrintResults, PrintMatrix, JUnit XML (with pass@k), markdown (with reliability), badge JSON, reliability metrics, baseline write/load, regression detection, coverage JSON, snapshot save/load/compare, diff formatting |
| `internal/runner` | 123 | Recursive fixture substitution, capture/extractJSONPath, server override, bad binary, timeout, Docker flag, transport selection (stdio/SSE/HTTP), URL validation, generate schema parsing, stub generation, filename sanitization, CLI error paths, client capabilities (handler unit tests, fixture substitution, capability path selection, bad-server error paths), prompt assertions (list/get/validation/fixture), progress capture, fix mode, fixture isolation, intercept, logging, sampling, completion |
| Total | 218 | Race-detector clean |

---

## YAML Assertion Format

```yaml
name: Human-readable description
server:
  command: path/to/mcp-server        # stdio transport
  args: ["arg1", "arg2"]
  env:
    KEY: value                         # supports ${VAR} expansion from shell
  transport: stdio                   # "stdio" (default), "sse", or "http"
  url: "http://localhost:8080/sse"   # required for sse/http transport
  client_capabilities:               # optional: respond to server-initiated requests
    roots:                           # respond to roots/list
      - "{{fixture}}"
    sampling:                        # respond to sampling/createMessage
      text: "mock LLM response"
      model: mock                    # default: "mock"
      stop_reason: end_turn          # default: "end_turn"
    elicitation:                     # respond to elicitation/create
      content:
        projectName: "myapp"
        framework: "react"
setup:
  - tool: setup_tool
    args: { key: value }
    capture:
      variable_name: "$.json.path"    # extract from response
assert:
  tool: tool_under_test
  args: { key: value }
  capture_progress: true             # optional: collect notifications/progress
  expect:
    not_error: true
    contains: ["expected"]
    json_path:
      "$.field": "value"
    min_results: 3
    min_progress: 2                  # requires capture_progress: true
skip: false                            # when true, assertion is skipped (set automatically by generate for destructive tools)
timeout: 30s

# OR: test MCP prompts instead of a tool
assert_prompts:
  list: {}                           # call prompts/list
  expect:
    not_empty: true
    contains: ["my_prompt"]

# OR: get a specific prompt with arguments
assert_prompts:
  get:
    name: "code_review"
    arguments:
      language: "go"
  expect:
    contains: ["review"]

# OR: test MCP resources
assert_resources:
  list: {}                           # call resources/list
  expect:
    not_empty: true
    contains: ["test://static/resource"]

# OR: read a specific resource by URI
assert_resources:
  read: "test://static/resource"
  expect:
    not_empty: true

# OR: test MCP completion (autocompletion)
assert_completion:
  ref:
    type: "ref/prompt"               # "ref/prompt" or "ref/resource"
    name: "complex_prompt"
  argument:
    name: "style"
    value: ""                        # partial value for completion
  expect:
    not_empty: true
    contains: ["formal"]

# OR: test sampling-triggered tool (mock LLM in one block)
assert_sampling:
  tool: ask_llm
  args:
    question: "What is the capital of France?"
  mock_text: "The capital of France is Paris."
  mock_model: mock-gpt               # optional
  expect:
    not_error: true
    contains: ["Paris"]

# OR: test logging (setLevel + message capture)
assert_logging:
  set_level: debug
  tool: echo
  args:
    message: "test"
  expect:
    min_messages: 1
    contains_level: ["debug"]
    contains_data: ["test"]

# OR: trajectory assertion (no server; uses trace: or audit_log:)
trace:
  - tool: prepare_rename
    args: { file_path: "main.go", line: 6, column: 6 }
  - tool: rename_symbol
    args: { file_path: "main.go", new_name: "Entity" }
trajectory:
  - type: order
    tools: ["prepare_rename", "rename_symbol"]
  - type: absence
    tools: ["apply_edit"]
  - type: args_contain
    tool: rename_symbol
    args:
      new_name: "Entity"
```

`{{fixture}}` in args is replaced with `--fixture` directory at runtime.

---

## Architecture

```
cmd/mcp-assert/main.go     CLI entry, command dispatch
internal/assertion/
  types.go                  Suite, Assertion, Expect, Result types + all block types
  loader.go                 YAML file loading, subdirectory recursion
  checker.go                15 assertion type implementations
  trajectory.go             4 trajectory assertion types (order, presence, absence, args_contain)
  sampling_types.go         SamplingAssertBlock type
  logging_types.go          LoggingAssertBlock, LoggingExpect, LogMessage types
  logging_checker.go        Logging assertion checker
internal/runner/
  runner.go                 Run, Matrix, CI commands, MCP client lifecycle
  client.go                 MCP client creation, transport selection, client capabilities
  commands.go               CLI command dispatch
  execute.go                Assertion routing (assert, resources, prompts, completion, sampling, logging, trajectory)
  coverage.go               Coverage command, tools/list query, --coverage-json
  fix.go                    --fix mode: ScanNearbyPositions, FixSuggestion, YAML patch generation
  fixture.go                Per-assertion fixture isolation (temp directory copy)
  generate.go               Auto-generate stub assertions from tools/list
  init.go                   Scaffold assertion template and fixture directory; init --server one-step generation
  intercept.go              intercept command: stdio proxy, live tool call capture, trajectory validation
  logging.go                runLoggingAssertion: logging/setLevel + notifications/message
  sampling.go               runSamplingAssertion: sampling-triggered tool calls
  snapshot.go               Snapshot capture/compare command
  substitute.go             {{fixture}} and ${VAR} substitution
  util.go                   Shared utilities
  watch.go                  File-watching rerun loop with unified diff on status flips
internal/report/
  report.go                 Terminal output (color-aware)
  color.go                  ANSI color, TTY detection, progress
  diff.go                   FormatDiff, FormatStatusChange: unified diff for watch mode status flips
  junit.go                  JUnit XML generation (with pass@k properties)
  markdown.go               GitHub Step Summary (with reliability table)
  badge.go                  shields.io endpoint JSON
  reliability.go            pass@k / pass^k computation
  baseline.go               Baseline write/load, regression detection
  coverage.go               Coverage JSON serialization
  snapshot.go               Snapshot file read/write/compare
```
