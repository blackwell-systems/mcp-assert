# CLI Reference

## Commands

### `mcp-assert init`

Scaffold an assertion template and fixture directory.

```bash
mcp-assert init [dir]
```

Creates `<dir>/read_file.yaml` (a commented assertion template) and `<dir>/fixtures/hello.txt` (a fixture file). Default directory is `evals`.

### `mcp-assert run`

Execute assertions against an MCP server.

```bash
mcp-assert run --suite <dir> [flags]
```

| Flag | Description |
|------|-------------|
| `--suite <dir>` | Directory containing assertion YAML files (required) |
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
| `--timeout <duration>` | Per-assertion timeout (default: `30s`) |

**Exit codes:** 0 = all passed, 1 = one or more failures.

### `mcp-assert ci`

Run with CI-specific exit codes and reporting. Supports all `run` flags plus CI-specific flags:

```bash
mcp-assert ci --suite <dir> [flags]
```

| Flag | Description |
|------|-------------|
| `--threshold <n>` | Minimum pass percentage (e.g., `95`) |
| `--fail-on-regression` | Exit 1 if a previously-passing assertion now fails (requires `--baseline`) |

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
mcp-assert generate --server <cmd> --output <dir> [--fixture <dir>]
```

Queries `tools/list`, reads input schemas, and creates one YAML per tool with sensible defaults. Edit the generated YAMLs to replace `TODO` placeholders with real values.

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

### `mcp-assert version`

Print the installed version.

```bash
mcp-assert version
```

```
mcp-assert v0.1.1
```

## Server Override

Override the server config from CLI instead of repeating it in every YAML file:

```bash
mcp-assert run --suite evals/ --server "agent-lsp go:gopls" --fixture test/fixtures/go
```

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
