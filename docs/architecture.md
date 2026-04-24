# Architecture

## Overview

mcp-assert is a single Go binary that tests MCP servers by starting them as subprocesses, communicating over the MCP stdio transport, and asserting tool responses against YAML-defined expectations.

```
                              stdio (default)
┌──────────────────┐     ──────────────────>  ┌──────────────────┐
│   mcp-assert     │     <──────────────────  │   MCP Server     │
│   (Go binary)    │        JSON-RPC          │   (any language)  │
│                  │                          │                  │
│  Load YAML       │     SSE / HTTP           │  stdio: launched │
│  Initialize MCP  │     ──────────────────>  │   as subprocess  │
│  Call tools      │     <──────────────────  │  http/sse: remote│
│  Assert results  │        JSON-RPC          │   via URL        │
└──────────────────┘                          └──────────────────┘
```

## Data Flow

```
YAML files ──> Loader ──> Assertions ──> Runner ──> MCP Client ──> Server
                                           │
                                           ├──> Checker (15 assertion types + 4 trajectory types)
                                           │
                                           └──> Reporter (terminal, JUnit, markdown, badge)
```

### 1. Load Phase (`internal/assertion/loader.go`)

The loader reads a directory of YAML files (or a single YAML file, for `--suite path/to/file.yaml`), parses each into an `Assertion` struct, and returns a `Suite`. When given a directory, it recurses one level into subdirectories. Files must have `.yaml` or `.yml` extension. The `name` field defaults to the filename if omitted.

### 2. Run Phase (`internal/runner/runner.go`)

For each assertion:

1. **Start server**: create the MCP client via `createMCPClient`, which selects the transport based on the `transport` field: stdio (default, launches subprocess via `client.NewStdioMCPClient`), SSE (`client.NewSSEMCPClient`), or streamable HTTP (`client.NewStreamableHttpClient`). If `--docker` is set with stdio, the command is wrapped in `docker run --rm -i` with volume mounts.

   1.5. **Client capabilities**: if `client_capabilities` is set in the server config (roots, sampling, or elicitation), `createStdioClientWithCapabilities` is used instead of `NewStdioMCPClient`. It constructs a raw `clienttransport.StdioTransport`, registers bidirectional request handlers (`staticRootsHandler`, `staticSamplingHandler`, `staticElicitationHandler`), and calls `c.Start(ctx)` to activate them before `Initialize` is called.

2. **Initialize**: send `initialize` request with MCP protocol version, receive server capabilities.
3. **Route**: dispatch to the appropriate handler based on which block is set: `runResourceAssertion` (assert_resources), `runPromptAssertion` (assert_prompts), `runCompletionAssertion` (assert_completion), `runSamplingAssertion` (assert_sampling), `runLoggingAssertion` (assert_logging), or the default tools/call path.
4. **Progress registration**: if `capture_progress: true` is set on the `assert:` block, register an `OnNotification` handler that counts `notifications/progress` messages before running setup.
5. **Setup**: execute setup tool calls sequentially (e.g., `start_lsp`, `open_document`). These establish the state needed for the assertion. `{{fixture}}` substitution happens here.
6. **Snapshot**: if `file_unchanged` assertions exist, read the files before the tool call.
7. **Assert**: call the tool under test, capture the response text and `isError` flag.
8. **Check**: run all expectations against the response (`internal/assertion/checker.go`). Then, if `capture_progress` was set, call `CheckProgress` to verify the notification count.
9. **Close**: shut down the MCP client (kills the server subprocess).

Each assertion gets its own server process. No state leaks between assertions.

### 3. Check Phase (`internal/assertion/checker.go`)

The checker evaluates expectations in a fixed order:

1. `not_error` / `is_error`: check `isError` flag
2. `not_empty`: reject empty, null, [], {}
3. `equals`: exact match (whitespace-trimmed)
4. `contains` / `not_contains`: substring checks
5. `matches_regex`: compiled regex matching
6. `json_path`: dot-notation lookup on parsed JSON
7. `min_results` / `max_results`: array length bounds
8. `net_delta`: numeric field comparison
9. `file_contains`: read file from disk, check content
10. `in_order`: ordered substring search

`file_unchanged` is handled separately via `CheckWithSnapshots` which compares post-execution file content against pre-execution snapshots.

`min_progress` is evaluated by `CheckProgress` after the tool call returns. It receives the notification count collected by the `OnNotification` handler registered before the tool call (when `capture_progress: true` is set).

First failure short-circuits: only the first failing expectation is reported.

### 4. Report Phase (`internal/report/`)

Results are dispatched to multiple output sinks:

| File | Responsibility |
|------|---------------|
| `report.go` | Terminal table with color (TTY) or plain (pipe) |
| `color.go` | ANSI codes, TTY detection, `NO_COLOR` support, progress indicator |
| `diff.go` | Unified diff formatting for watch mode status flips |
| `junit.go` | JUnit XML serialization via `encoding/xml` |
| `markdown.go` | GitHub Step Summary markdown table |
| `badge.go` | shields.io endpoint JSON |
| `reliability.go` | pass@k / pass^k computation from multi-trial results |
| `baseline.go` | Baseline JSON write/load, regression detection |
| `coverage.go` | Coverage JSON serialization |
| `snapshot.go` | Snapshot file read/write/compare |

All report outputs are best-effort: write errors go to stderr but don't fail the run.

### 5. Coverage Phase (`internal/runner/coverage.go`)

The `coverage` command takes a different path:

1. Load the assertion suite (same as run)
2. Start the MCP server
3. Call `tools/list` to discover all server tools
4. Compare server tool names against assertion tool names
5. Report coverage percentage and per-tool status

This does not execute any assertions: it only queries the tool catalog.

## Key Design Decisions

**One server per assertion.** Each assertion starts a fresh MCP server subprocess. This prevents state leakage between tests but means server startup cost is paid per assertion. For fast servers (filesystem, memory) this is negligible. For slow servers (gopls, jdtls) it dominates test duration. The `setup` block amortizes some of this by allowing warmup calls within a single assertion's server lifetime.

**Checker is pure.** `Check()` takes a string and returns an error. No I/O, no state, no side effects. `CheckWithSnapshots()` adds file comparison but the snapshots are passed in, not read internally. This makes the checker trivially testable.

**Color degrades gracefully.** TTY detection via `os.ModeCharDevice`. `NO_COLOR` env var. `TERM=dumb`. In CI (pipes), output is plain `PASS`/`FAIL`/`SKIP`: no escape codes in JUnit XML or log files.

**Docker is a command wrapper (stdio only).** `--docker <image>` doesn't use the Docker SDK. It prepends `docker run --rm -i -v fixture:fixture` to the server command. Since MCP uses stdio, Docker's `-i` flag gives bidirectional pipe transport for free. The server process runs inside the container; the assertions run outside. Docker is only supported with stdio transport; HTTP/SSE transports connect to an already-running server.

**Transport is pluggable.** `createMCPClient` is a single helper (in `runner.go`) used by `runAssertion`, `runAndCapture`, and `Coverage`. It reads the `transport` field from `ServerConfig` and creates the appropriate mcp-go client. All three transports (stdio, SSE, streamable HTTP) return the same `MCPClient` interface, so the rest of the runner is transport-agnostic.

**Client capabilities use a separate construction path.** When `client_capabilities` has roots, sampling, or elicitation set, `createMCPClient` delegates to `createStdioClientWithCapabilities` instead of `client.NewStdioMCPClient`. The key difference: `NewStdioMCPClient` is a convenience wrapper that starts the transport and returns a connected client, whereas `createStdioClientWithCapabilities` uses the lower-level `client.NewClient` with explicit handler options so bidirectional handlers are registered before the client starts. `c.Start(ctx)` must be called after `NewClient` to register the handlers; without it, the server's `roots/list` or `sampling/createMessage` requests arrive before handlers are wired up.

**Fixture isolation.** Each stdio assertion automatically receives its own copy of the fixture directory via a temporary directory. The original fixture is never modified, so assertions that write files cannot affect subsequent assertions. Docker mode already isolates via fresh containers, so the copy is skipped when `--docker` is used.

**Setup tools are not counted as "tested" by coverage.** `start_lsp` and `open_document` appear in every assertion's setup but aren't the tools under test. The coverage command only counts the `assert.tool` field.

## Package Dependency Graph

```
cmd/mcp-assert/main.go
  └── internal/runner
        ├── internal/assertion (types, loader, checker)
        ├── internal/report (all output formats)
        ├── mark3labs/mcp-go/client (MCP transport: stdio, SSE, streamable HTTP)
        └── mark3labs/mcp-go/mcp (MCP protocol types)
```

No circular dependencies. The `assertion` and `report` packages do not import each other (report depends on assertion types, not the reverse).
