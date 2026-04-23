# Architecture

## Overview

mcp-assert is a single Go binary that tests MCP servers by starting them as subprocesses, communicating over the MCP stdio transport, and asserting tool responses against YAML-defined expectations.

```
┌──────────────────┐     stdio      ┌──────────────────┐
│   mcp-assert     │ ──────────────>│   MCP Server     │
│   (Go binary)    │ <──────────────│   (any language)  │
│                  │   JSON-RPC     │                  │
│  Load YAML       │                │  Start via cmd   │
│  Initialize MCP  │                │  Respond to      │
│  Call tools      │                │  tool calls      │
│  Assert results  │                │                  │
└──────────────────┘                └──────────────────┘
```

## Data Flow

```
YAML files ──> Loader ──> Assertions ──> Runner ──> MCP Client ──> Server
                                           │
                                           ├──> Checker (14 assertion types)
                                           │
                                           └──> Reporter (terminal, JUnit, markdown, badge)
```

### 1. Load Phase (`internal/assertion/loader.go`)

The loader reads a directory of YAML files (recursing one level into subdirectories), parses each into an `Assertion` struct, and returns a `Suite`. Files must have `.yaml` or `.yml` extension. The `name` field defaults to the filename if omitted.

### 2. Run Phase (`internal/runner/runner.go`)

For each assertion:

1. **Start server** — launch the MCP server as a subprocess via `client.NewStdioMCPClient`. If `--docker` is set, the command is wrapped in `docker run --rm -i` with volume mounts.
2. **Initialize** — send `initialize` request with MCP protocol version, receive server capabilities.
3. **Setup** — execute setup tool calls sequentially (e.g., `start_lsp`, `open_document`). These establish the state needed for the assertion. `{{fixture}}` substitution happens here.
4. **Snapshot** — if `file_unchanged` assertions exist, read the files before the tool call.
5. **Assert** — call the tool under test, capture the response text and `isError` flag.
6. **Check** — run all expectations against the response (`internal/assertion/checker.go`).
7. **Close** — shut down the MCP client (kills the server subprocess).

Each assertion gets its own server process. No state leaks between assertions.

### 3. Check Phase (`internal/assertion/checker.go`)

The checker evaluates expectations in a fixed order:

1. `not_error` / `is_error` — check `isError` flag
2. `not_empty` — reject empty, null, [], {}
3. `equals` — exact match (whitespace-trimmed)
4. `contains` / `not_contains` — substring checks
5. `matches_regex` — compiled regex matching
6. `json_path` — dot-notation lookup on parsed JSON
7. `min_results` / `max_results` — array length bounds
8. `net_delta` — numeric field comparison
9. `file_contains` — read file from disk, check content
10. `in_order` — ordered substring search

`file_unchanged` is handled separately via `CheckWithSnapshots` which compares post-execution file content against pre-execution snapshots.

First failure short-circuits — only the first failing expectation is reported.

### 4. Report Phase (`internal/report/`)

Results are dispatched to multiple output sinks:

| File | Responsibility |
|------|---------------|
| `report.go` | Terminal table with color (TTY) or plain (pipe) |
| `color.go` | ANSI codes, TTY detection, `NO_COLOR` support, progress indicator |
| `junit.go` | JUnit XML serialization via `encoding/xml` |
| `markdown.go` | GitHub Step Summary markdown table |
| `badge.go` | shields.io endpoint JSON |
| `reliability.go` | pass@k / pass^k computation from multi-trial results |
| `baseline.go` | Baseline JSON write/load, regression detection |

All report outputs are best-effort — write errors go to stderr but don't fail the run.

### 5. Coverage Phase (`internal/runner/coverage.go`)

The `coverage` command takes a different path:

1. Load the assertion suite (same as run)
2. Start the MCP server
3. Call `tools/list` to discover all server tools
4. Compare server tool names against assertion tool names
5. Report coverage percentage and per-tool status

This does not execute any assertions — it only queries the tool catalog.

## Key Design Decisions

**One server per assertion.** Each assertion starts a fresh MCP server subprocess. This prevents state leakage between tests but means server startup cost is paid per assertion. For fast servers (filesystem, memory) this is negligible. For slow servers (gopls, jdtls) it dominates test duration. The `setup` block amortizes some of this by allowing warmup calls within a single assertion's server lifetime.

**Checker is pure.** `Check()` takes a string and returns an error. No I/O, no state, no side effects. `CheckWithSnapshots()` adds file comparison but the snapshots are passed in, not read internally. This makes the checker trivially testable.

**Color degrades gracefully.** TTY detection via `os.ModeCharDevice`. `NO_COLOR` env var. `TERM=dumb`. In CI (pipes), output is plain `PASS`/`FAIL`/`SKIP` — no escape codes in JUnit XML or log files.

**Docker is a command wrapper.** `--docker <image>` doesn't use the Docker SDK. It prepends `docker run --rm -i -v fixture:fixture` to the server command. Since MCP uses stdio, Docker's `-i` flag gives bidirectional pipe transport for free. The server process runs inside the container; the assertions run outside.

**Setup tools are not counted as "tested" by coverage.** `start_lsp` and `open_document` appear in every assertion's setup but aren't the tools under test. The coverage command only counts the `assert.tool` field.

## Package Dependency Graph

```
cmd/mcp-assert/main.go
  └── internal/runner
        ├── internal/assertion (types, loader, checker)
        ├── internal/report (all output formats)
        ├─�� mark3labs/mcp-go/client (MCP stdio transport)
        └── mark3labs/mcp-go/mcp (MCP protocol types)
```

No circular dependencies. The `assertion` and `report` packages do not import each other (report depends on assertion types, not the reverse).
