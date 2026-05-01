# CLI Reference

## Commands

```
mcp-assert audit    --server <cmd> [--output <dir>] [--docker <image>] [--include-writes]
mcp-assert init     [dir] [--server <cmd>] [--fixture <dir>] [--timeout <duration>]
mcp-assert run      --suite <path> [--fix] [flags]
mcp-assert ci       --suite <path> [--fix] [flags]
mcp-assert matrix   --suite <dir> --languages <list> [flags]
mcp-assert coverage --suite <dir> --server <cmd> [flags]
mcp-assert generate --server <cmd> --output <dir> [flags]
mcp-assert snapshot --suite <dir> --server <cmd> [flags]
mcp-assert watch    --suite <dir> [flags]
mcp-assert intercept --server <cmd> --trajectory <path> [--timeout <duration>]
mcp-assert version
```

### `mcp-assert audit`

Zero-config quality audit. Connects to a server, discovers all tools, calls each one with schema-generated inputs, and reports which tools are healthy vs. which crash. No YAML required.

```bash
mcp-assert audit --server "npx my-mcp-server" [--output evals/]
```

| Flag | Description |
|------|-------------|
| `--server <cmd>` | Server command (stdio) or URL (http/sse) (required) |
| `--transport <type>` | Transport type: `stdio` (default), `http`, `sse` |
| `--headers <pairs>` | Custom headers as `key=value` pairs, comma-separated |
| `--docker <image>` | Run destructive tools in fresh Docker containers (stdio only) |
| `--timeout <duration>` | Per-tool call timeout (default: `15s`) |
| `--output <dir>` | Generate starter assertion YAML files in this directory |
| `--include-writes` | Also call destructive/write tools without Docker isolation (skipped by default) |
| `--json` | Output results as JSON |

**What it tests:** Crash resistance and error handling. The audit calls every tool with valid-shaped inputs generated from JSON Schema. Tools that respond (whether with data or a proper `isError: true`) are scored as healthy. Tools that crash with internal errors (`-32603`), stack traces, or panics are scored as crashed.

**What it doesn't test:** Business logic, expected output content, multi-step workflows, or state verification. For those, use the YAML assertion workflow (see `run`, `ci`).

**Destructive tool handling:** By default, tools annotated as destructive are skipped. Two ways to test them:

- `--docker <image>`: Spins up a fresh Docker container per destructive tool. Each tool gets an isolated environment; the container is destroyed afterward. Safe for write/delete tools.
- `--include-writes`: Calls destructive tools directly on the host, without isolation. Use only when you understand the side effects.

**Generating YAML for CI:** Pass `--output <dir>` to generate one assertion YAML per tool. These stubs use `not_error: true` as the default expectation. Edit them to add expected content checks, setup steps, and multi-step flows, then run them in CI with `mcp-assert ci --suite <dir>`.

### `mcp-assert init`

Scaffold an assertion template, or generate a complete working suite from a live server.

```bash
# Template mode (no server required)
mcp-assert init [dir]

# One-step suite generation (queries the server, creates stubs, captures snapshots)
mcp-assert init [dir] --server <cmd> [--fixture <dir>]
```

| Flag | Description |
|------|-------------|
| `--server <cmd>` | Server command to query for `tools/list`. When provided, runs generate + snapshot in one step |
| `--fixture <dir>` | Fixture directory for `{{fixture}}` substitution in generated assertions |
| `--timeout <duration>` | Timeout for `tools/list` call (default: `15s`) |

**Without `--server`:** Creates `<dir>/read_file.yaml` (a commented assertion template) and `<dir>/fixtures/hello.txt` (a fixture file). Default directory is `evals`.

**With `--server`:** Connects to the server, queries `tools/list`, generates one stub YAML per tool, then runs snapshot capture with `--update` to record baseline responses. The result is a complete working suite with 100% tool coverage and zero manual assertion writing. Destructive tools are generated with `skip: true` by default.

### `mcp-assert run`

Execute assertions against an MCP server.

```bash
mcp-assert run --suite <path> [flags]
```

| Flag | Description |
|------|-------------|
| `--suite <path>` | Directory or single YAML file containing assertions (required) |
| `--fixture <dir>` | Fixture directory for `{{fixture}}` substitution |
| `--server <cmd>` | Override server command from CLI instead of per-YAML |
| `--trials <n>` | Run each assertion N times for reliability metrics |
| `--docker <image>` | Run each assertion in a fresh Docker container |
| `--json` | Output full result array as JSON to stdout |
| `--junit <path>` | Write JUnit XML results |
| `--markdown <path>` | Write GitHub Step Summary markdown |
| `--badge <path>` | Write shields.io endpoint JSON |
| `--baseline <path>` | Compare against saved baseline |
| `--save-baseline <path>` | Save current results as baseline JSON |
| `--fix` | Scan nearby positions when position-sensitive assertions fail and suggest corrections |
| `--timeout <duration>` | Per-assertion timeout (default: `30s`) |

**Exit codes:** 0 = all passed, 1 = one or more failures.

### `mcp-assert ci`

Run with CI-specific exit codes and reporting. Supports all `run` flags plus CI-specific flags:

```bash
mcp-assert ci --suite <path> [flags]
```

| Flag | Description |
|------|-------------|
| `--threshold <n>` | Minimum pass percentage (e.g., `95`) |
| `--fail-on-regression` | Exit 1 if a previously-passing assertion now fails (requires `--baseline`) |
| `--fix` | Scan nearby positions when position-sensitive assertions fail and suggest corrections |

Auto-detects `$GITHUB_STEP_SUMMARY` for markdown output.

### `mcp-assert matrix`

Run assertions across multiple language servers.

```bash
mcp-assert matrix --suite <dir> --languages <list> [--fixture <dir>]
```

```bash
mcp-assert matrix \
  --suite evals/ \
  --languages go:gopls,typescript:typescript-language-server,python:pyright-langserver
```

Output:

```
                     hover           definition        references     completions
Go (gopls)           PASS            PASS              PASS           PASS
TypeScript (tsserver) PASS            PASS              PASS           PASS
Python (pyright)     PASS            PASS              SKIP           PASS
```

### `mcp-assert coverage`

Report which server tools have assertions and which don't.

```bash
mcp-assert coverage --suite <dir> --server <cmd> [--coverage-json <path>]
```

Starts the server, calls `tools/list`, compares against assertion tool names, and reports coverage percentage with covered/uncovered tool lists.

```
Server exposes 50 tools, 50 have assertions (100% coverage)

Covered (50):
  + add_workspace_folder (1 assertion)
  + call_hierarchy (1 assertion)
  ...

Not covered (0):
  (none)
```

### `mcp-assert generate`

Auto-generate stub assertions from a server's `tools/list` response.

```bash
mcp-assert generate --server <cmd> --output <dir> [--fixture <dir>] [--include-writes]
```

| Flag | Description |
|------|-------------|
| `--server <cmd>` | Server command to query for `tools/list` (required) |
| `--output <dir>` | Directory to write generated YAML files (required) |
| `--fixture <dir>` | Fixture directory for `{{fixture}}` substitution in generated stubs |
| `--include-writes` | Include destructive/write tools that are skipped by default |

Queries `tools/list`, reads input schemas, and creates one YAML per tool with sensible defaults. Edit the generated YAMLs to replace `TODO` placeholders with real values.

**Destructive tool handling:** Tools annotated as destructive (`destructiveHint: true`) or not explicitly read-only (`readOnlyHint: false`) are skipped by default. This prevents accidentally running tools that modify state during testing. The skipped tools are generated with `skip: true` in their YAML. Pass `--include-writes` to include all tools without the skip marker.

**Auth detection:** If a tool's input schema includes properties with names like `token`, `api_key`, or `password`, the generated YAML includes a comment hinting that authentication may be required. Review these stubs and configure credentials via environment variables before running.

### `mcp-assert snapshot`

Capture or compare tool response snapshots.

```bash
mcp-assert snapshot --suite <dir> --server <cmd> [--fixture <dir>] [--update] [--docker <image>]
```

| Flag | Description |
|------|-------------|
| `--update` | Capture actual outputs and save as `.snapshots.json` |
| (no `--update`) | Assert current outputs match saved snapshots |

### `mcp-assert watch`

Rerun assertions automatically when YAML files change.

```bash
mcp-assert watch --suite <dir> [--server <cmd>] [--fixture <dir>] [--interval <duration>]
```

| Flag | Description |
|------|-------------|
| `--interval <duration>` | Polling interval (default: `2s`) |

Polls for changes, clears terminal between runs. The assertion development loop: edit YAML, save, see result.

When an assertion's status changes between iterations (e.g., PASS to FAIL), watch mode displays a unified diff of the expected vs actual response to help diagnose the change.

### `mcp-assert intercept`

Proxy stdio between an agent and an MCP server, capturing every tool call in real time.

```bash
mcp-assert intercept --server <cmd> --trajectory <path> [--timeout <duration>]
```

| Flag | Description |
|------|-------------|
| `--server <cmd>` | MCP server command to proxy traffic to (required) |
| `--trajectory <path>` | YAML file containing trajectory assertions to validate on disconnect (required) |
| `--timeout <duration>` | Timeout for the proxy session (default: no timeout) |

Sits between your agent (on stdin/stdout) and the MCP server, forwarding all JSON-RPC messages transparently while recording every `tools/call` invocation. When the agent disconnects, intercept validates the captured call sequence against any trajectory assertions in the `--trajectory` file and reports the results. Use this as an alternative to `trace:` or `audit_log:` when you want to validate a real agent session without modifying the agent itself.

### `mcp-assert version`

Print the installed version.

```bash
mcp-assert version
```

```
mcp-assert v0.7.3
```

## Server Override

Override the server config from CLI instead of repeating it in every YAML file:

```bash
mcp-assert run --suite evals/ --server "agent-lsp go:gopls" --fixture test/fixtures/go
```

## Skipping Assertions

Add `skip: true` to any assertion YAML to exclude it from `run` and `ci` execution:

```yaml
name: dangerous tool that modifies state
skip: true
server:
  command: my-server
assert:
  tool: delete_everything
  args: {}
  expect:
    not_error: true
```

Skipped assertions appear as `SKIP` in output and do not count toward pass or fail totals. This is useful for temporarily disabling flaky tests, for assertions that require external services, or for destructive tools that should not run in CI.

The `generate` command automatically sets `skip: true` on tools detected as destructive. Use `--include-writes` to generate stubs without the skip marker.

## Docker Isolation

Run each assertion in a fresh Docker container for reproducibility:

```bash
mcp-assert run --suite evals/ --docker ghcr.io/blackwell-systems/agent-lsp:go --fixture /workspace
```

The fixture directory is mounted into the container. Each assertion gets a clean environment: no cross-test contamination, no "works on my machine."

Docker isolation is only supported with stdio transport (the default). HTTP/SSE transports connect to an already-running server and do not use Docker wrapping.

## Client Capabilities

Client capabilities are configured per-assertion in YAML, not via CLI flags. Set `client_capabilities` in the server block to make mcp-assert respond to server-initiated requests (roots, sampling, elicitation):

```yaml
server:
  command: /path/to/server
  client_capabilities:
    roots:
      - "{{fixture}}"
    sampling:
      text: "mock response"
    elicitation:
      content:
        confirmed: true
```

See [Writing Assertions](writing-assertions.md#client-capabilities-bidirectional-mcp) for full examples of each capability type.

## Resource Assertions

`assert_resources:` is a YAML-level feature with no CLI flag equivalent. It replaces `assert:` to test `resources/list` or `resources/read` instead of tools/call:

```yaml
assert_resources:
  list: {}          # or: read: "uri://resource"
  expect:
    not_empty: true
    contains: ["expected-resource"]
```

See [Writing Assertions](writing-assertions.md#resource-assertions) for full examples.

## Prompt Assertions

`assert_prompts:` is a YAML-level feature with no CLI flag equivalent. It replaces `assert:` to test `prompts/list` or `prompts/get` instead of tools/call:

```yaml
assert_prompts:
  list: {}          # or: get: {name: "my_prompt", arguments: {key: val}}
  expect:
    not_empty: true
    contains: ["expected_prompt"]
```

See [Writing Assertions](writing-assertions.md#prompt-assertions) for full examples including pagination.

## Trajectory Assertions

`trace:` and `audit_log:` are YAML-level features that replace `server:` for trajectory-based assertions. No CLI flag equivalent. No server is started.

```yaml
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
```

Replace `trace:` with `audit_log: path/to/agent.jsonl` to validate real agent behavior from a recorded JSONL log.

See [Writing Assertions](writing-assertions.md#trajectory-assertions) for the full format and all four assertion types.

## Notification Assertions

`assert_notifications:` is a YAML-level feature that captures all server notifications during a tool call and asserts on them:

```yaml
assert_notifications:
  tool: long_running_task
  args:
    input: "test"
  expect:
    min_count: 3
    methods: ["notifications/progress"]
    contains_data: ["processing"]
```

Six expectation fields: `min_count`, `max_count`, `methods`, `not_methods`, `contains_data`, `not_contains_data`. See [Writing Assertions](writing-assertions.md) for full details.

## Conditional Skipping

Skip assertions when an environment variable is not set:

```yaml
name: search with API key
skip_unless_env: SEARCH_API_KEY
server:
  command: my-server
  env:
    API_KEY: "${SEARCH_API_KEY}"
assert:
  tool: search
  args:
    query: "test"
  expect:
    not_error: true
```

When `SEARCH_API_KEY` is not set, the assertion is skipped with a clear message. When set, it runs normally. This enables auth-gated assertions to coexist with no-auth assertions in the same suite.

## Progress Capture

`capture_progress` and `min_progress` are YAML-level features on the `assert:` block, not CLI flags:

```yaml
assert:
  tool: long_operation
  args: {}
  capture_progress: true
  expect:
    min_progress: 3
```

See [Writing Assertions](writing-assertions.md#progress-notifications) for details.

## HTTP/SSE Transport

Transport is configured per-assertion in YAML, not via CLI flags. Set `transport: sse` or `transport: http` with a `url` field to connect to HTTP-based MCP servers instead of launching a subprocess:

```yaml
server:
  transport: sse
  url: "http://localhost:8080/sse"
```

See [Writing Assertions](writing-assertions.md#httpsse-transport) for full examples.

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

## Terminal Output

mcp-assert uses color in interactive terminals: green for pass, red for fail, yellow for skip. A progress counter (`[1/21]`, `[2/21]`, ...) prints to stderr while assertions run. The summary line only shows non-zero counts.

Color and progress are automatically disabled in pipes and CI environments. Set `NO_COLOR=1` to force plain `PASS`/`FAIL`/`SKIP` output explicitly.

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
