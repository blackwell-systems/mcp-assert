# Architecture

## Overview

mcp-assert is an MCP client that tests MCP servers. It connects to a server exactly like Claude, Cursor, or any other MCP host would: full initialize handshake, standard JSON-RPC tool calls, proper transport negotiation. The server cannot distinguish mcp-assert from a real agent. This is the core design principle: test against the real protocol, not a mock.

It reads test definitions from YAML files, starts each MCP server as a subprocess (or connects to a remote one), sends requests using the MCP protocol, and checks the responses against expectations you define. If the response matches, the test passes. If not, it fails with a clear error message.

The problem it solves: MCP servers expose tools, prompts, and resources to AI agents, but there is no built-in way to verify that those capabilities return correct results. mcp-assert fills that gap by providing deterministic, repeatable correctness tests that run in CI or locally, for any MCP server written in any language.

The tool is a single Go binary with no runtime dependencies. You write YAML files describing what to call and what to expect, point mcp-assert at them, and get pass/fail results in the terminal, JUnit XML, markdown, or JSON.

---

## How MCP Works (Brief Primer)

The **Model Context Protocol (MCP)** is a standard for AI agents to interact with external services. If you are already familiar with MCP, skip to the next section.

**Servers and clients.** An MCP server is a program that exposes capabilities (tools, prompts, resources) over a well-defined protocol. An MCP client connects to the server and makes requests. In mcp-assert's case, mcp-assert itself is the client.

**Tools** are functions the server offers. A client calls a tool by name with JSON arguments and receives a text response. For example, a filesystem server might expose a `read_file` tool that accepts `{"path": "/tmp/foo.txt"}` and returns the file contents.

**Prompts** are reusable prompt templates the server provides. A client can list available prompts and retrieve a specific one (optionally with arguments that fill template variables).

**Resources** are data the server exposes for reading. A client can list available resources and read a specific one by URI, similar to a REST GET endpoint.

**JSON-RPC** is the wire format. Every MCP message is a JSON-RPC 2.0 request or response. The client sends `{"jsonrpc":"2.0", "method":"tools/call", "params":{...}, "id":1}` and the server replies with `{"jsonrpc":"2.0", "result":{...}, "id":1}`.

**Transports** determine how JSON-RPC messages travel between client and server:

- **stdio** (default): The client launches the server as a subprocess. JSON-RPC messages flow over the subprocess's stdin and stdout. This is the most common transport for local development.
- **SSE (Server-Sent Events)**: The client connects to an HTTP endpoint. The server pushes responses as SSE events. This is a legacy transport.
- **Streamable HTTP**: The client sends HTTP POST requests and receives streamed responses. This is the modern remote transport.

**The handshake.** Before any tool calls, the client and server perform an initialization exchange:

```
Client (mcp-assert)                          Server (under test)
       │                                            │
       │──── initialize ───────────────────────────>│
       │     { protocolVersion: "2025-03-26",       │
       │       clientInfo: { name: "mcp-assert" } } │
       │                                            │
       │<─── initialize response ──────────────────│
       │     { protocolVersion: "2025-03-26",       │
       │       capabilities: {                      │
       │         tools: { listChanged: true },      │
       │         prompts: {},                       │
       │         resources: {}                      │
       │       },                                   │
       │       serverInfo: { name: "my-server" } }  │
       │                                            │
       │──── initialized (notification) ───────────>│
       │                                            │
       │     ═══ handshake complete ═══             │
       │                                            │
       │──── tools/call ───────────────────────────>│
       │     { name: "read_file",                   │
       │       arguments: { path: "/tmp/foo" } }    │
       │                                            │
       │<─── result ───────────────────────────────│
       │     { content: [{ type: "text",            │
       │       text: "file contents" }] }           │
       │                                            │
```

The client declares its name and the protocol version it supports. The server responds with its own capabilities (which tools, prompts, and resources it offers). The client sends an `initialized` notification to confirm, and then the session is open for requests.

**Bidirectional requests.** MCP is not strictly client-to-server. The server can also request things from the client during a tool call:

```
Client (mcp-assert)                          Server (under test)
       │                                            │
       │──── tools/call { name: "refactor" } ──────>│
       │                                            │
       │<─── roots/list (server asks client) ──────│
       │                                            │
       │──── roots response ───────────────────────>│
       │     [{ uri: "file:///workspace" }]         │
       │                                            │
       │<─── sampling/createMessage ───────────────│
       │     (server asks for LLM completion)       │
       │                                            │
       │──── mock LLM response ────────────────────>│
       │     { text: "mocked answer" }              │
       │                                            │
       │<─── tools/call result ────────────────────│
       │                                            │
```

The server can request filesystem roots (`roots/list`), an LLM completion (`sampling/createMessage`), or user input (`elicitation/create`). mcp-assert supports mocking all three of these via the `client_capabilities` YAML field.

---

## Lifecycle of an Assertion

This section walks through exactly what happens when you run:

```
mcp-assert run --suite evals/ --fixture ./test-data
```

### Step 1: CLI dispatch

`cmd/mcp-assert/main.go` reads the first argument (`run`) and calls `runner.Run()`, passing the remaining flags. The `Run` function in `internal/runner/commands.go` parses `--suite`, `--fixture`, `--trials`, output flags, and others.

### Step 2: Load the suite

`assertion.LoadSuite("evals/")` (in `internal/assertion/loader.go`) reads the directory. It collects every `.yaml` and `.yml` file, recursing one level into subdirectories. Each file is parsed into an `Assertion` struct via Go's `yaml.v3` library. If the `name` field is omitted, the filename becomes the assertion name. The result is a `Suite` containing a slice of `Assertion` values.

### Step 3: Iterate and isolate

The runner loops over each assertion. Before executing, it calls `isolateFixture()` (in `internal/runner/fixture.go`), which copies the entire fixture directory to a temporary location. This ensures each assertion gets a pristine copy of the test data. The original fixture is never modified.

### Step 4: Start the MCP server

`createMCPClient()` (in `internal/runner/client.go`) reads the assertion's `server` block and selects the transport:

- **stdio**: launches the server command as a subprocess, piping stdin/stdout for JSON-RPC.
- **sse**: connects to the server's URL via SSE.
- **http**: connects via streamable HTTP.

If `client_capabilities` is set (roots, sampling, or elicitation), the stdio path uses a lower-level construction (`createStdioClientWithCapabilities`) that registers bidirectional request handlers before the client starts. This ensures the handlers are active before the `initialize` handshake, since the server may immediately request roots or sampling.

### Step 5: Initialize

The runner sends an `initialize` JSON-RPC request with the protocol version and client identity (`mcp-assert v1.0`). The server responds with its capabilities. This is the standard MCP handshake.

Steps 4 and 5 are combined in `initializedClientFromConfig()` (execute.go), which is the shared entry point for all commands that need a connected client:

```go
// execute.go — shared client lifecycle
func initializedClientFromConfig(
    server assertion.ServerConfig,
    fixture string,
    timeout time.Duration,
    dockerImage string,
) (context.Context, context.CancelFunc, client.MCPClient, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)

    mcpClient, err := createMCPClient(server, fixture, dockerImage)
    if err != nil {
        cancel()
        return nil, nil, nil, fmt.Errorf("failed to start MCP server: %w", err)
    }

    initReq := mcp.InitializeRequest{}
    initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
    initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION

    if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
        _ = mcpClient.Close()
        cancel()
        return nil, nil, nil, fmt.Errorf("MCP initialize failed: %w", err)
    }

    return ctx, cancel, mcpClient, nil
}
```

On error at any step, all resources are cleaned up before returning. On success, the caller is responsible for calling `cancel()` and `mcpClient.Close()` (typically via `defer`).

### Step 6: Route to the correct handler

`runAssertion()` (in `internal/runner/execute.go`) inspects which block is present on the assertion and dispatches accordingly:

| Block present | Handler called | MCP method used |
|---------------|----------------|-----------------|
| `trajectory:` | `runTrajectoryAssertion` | None (no server) |
| `assert_resources:` | `runResourceAssertion` | `resources/list` or `resources/read` |
| `assert_prompts:` | `runPromptAssertion` | `prompts/list` or `prompts/get` |
| `assert_completion:` | `runCompletionAssertion` | `completion/complete` |
| `assert_sampling:` | `runSamplingAssertion` | `tools/call` (triggers server-side sampling) |
| `assert_logging:` | `runLoggingAssertion` | `logging/setLevel` + `tools/call` |
| `assert:` (default) | inline in `runAssertion` | `tools/call` |

### Step 7: Run setup steps

If the assertion has a `setup:` block, the runner executes each setup tool call sequentially. Setup calls establish the state the assertion needs (for example, starting a language server or opening a document). Template substitution replaces `{{fixture}}` with the isolated fixture path and `{{variable_name}}` with values captured from prior setup responses.

### Step 8: Snapshot (if needed)

If the assertion's `expect` block contains `file_unchanged` entries, the runner reads those files from disk before the tool call and stores their contents in memory. After the tool call, it compares the files to detect modifications.

### Step 9: Call the tool under test

The runner sends the `tools/call` JSON-RPC request with the tool name and arguments (after template substitution). It captures the response text and the `isError` flag.

```go
// execute.go — the core tool call + check sequence (simplified)
assertArgs := substituteAll(a.Assert.Args, fixture, captured)

req := mcp.CallToolRequest{}
req.Params.Name = a.Assert.Tool
req.Params.Arguments = assertArgs

result, err := mcpClient.CallTool(ctx, req)
if err != nil {
    return failResult(a.Name, start, fmt.Sprintf("tool call %s failed: %v", a.Assert.Tool, err))
}

resultText := extractText(result)  // joins all TextContent blocks into a single string
isError := result.IsError
```

### Step 10: Check expectations

`assertion.Check()` (in `internal/assertion/checker.go`) evaluates every expectation in the `expect` block against the response. Expectations are checked in a fixed order (see "Key Abstractions" below). The first failure short-circuits: only the first failing expectation is reported, to keep error messages actionable.

```go
// checker.go — the checker is pure: string in, error out
func Check(expect Expect, response string, isError bool) error {
    for _, entry := range checkRegistry {
        if err := entry.fn(expect, response, isError); err != nil {
            return err  // short-circuit on first failure
        }
    }
    return nil
}
```

The `checkRegistry` is an ordered slice of check functions, not a map, so evaluation order is deterministic and documented. Each check function receives the full `Expect` struct but only inspects its own field (e.g., `checkContains` only looks at `expect.Contains`). If the field is nil or empty, the check is a no-op.

If `capture_progress: true` was set, the runner also checks `min_progress` via `assertion.CheckProgress()`, verifying that enough `notifications/progress` messages arrived during the tool call.

### Step 11: Clean up

The MCP client is closed (which kills the subprocess for stdio transport). The temporary fixture directory is removed. The `Result` struct (pass, fail, or skip, plus timing) is appended to the results list.

### Step 12: Report

After all assertions finish, the runner dispatches results to output sinks. The terminal always gets a human-readable table. Optional flags produce JUnit XML (`--junit`), markdown (`--markdown`), shields.io badge JSON (`--badge`), or raw JSON (`--json`). If `--trials` was greater than 1, reliability metrics (pass@k, pass^k) are also printed.

```
┌──────────┐     ┌──────────┐     ┌──────────────┐     ┌────────────┐
│  YAML    │────>│  Loader  │────>│   Runner     │────>│  Reporter  │
│  files   │     │          │     │  (per assert)│     │            │
└──────────┘     └──────────┘     │              │     │  Terminal   │
                                  │  Isolate     │     │  JUnit XML  │
                                  │  Start server│     │  Markdown   │
                                  │  Initialize  │     │  Badge JSON │
                                  │  Setup calls │     │  Raw JSON   │
                                  │  Tool call   │     └────────────┘
                                  │  Check       │
                                  │  Close       │
                                  └──────────────┘
```

---

## Package Structure

The codebase is organized into three packages under `internal/`, plus the entry point.

### `cmd/mcp-assert/main.go`

The binary entry point. It reads the first CLI argument and dispatches to the appropriate function in the `runner` package via a command registry map (`map[string]func([]string) error`). Commands include `Audit`, `Run`, `Matrix`, `CI`, `Init`, `Coverage`, `Generate`, `Snapshot`, `Watch`, and `Intercept`. This file also defines `printUsage()` for help text and exposes a `Version` variable set at build time.

### `internal/assertion/` (types, loading, checking)

This package defines the data model and all validation logic. It has no I/O beyond reading YAML files and checking files on disk for `file_contains`/`file_unchanged`. It does not import the runner or report packages.

| File | Responsibility |
|------|----------------|
| `types.go` | All core types: `Suite`, `Assertion`, `ServerConfig`, `AssertBlock`, `Expect`, `Result`, `Status`, and the block types for resources, prompts, completion, sampling, logging, and trajectory. |
| `loader.go` | `LoadSuite()` reads a directory (or single file) of YAML, parses each into an `Assertion`, and returns a `Suite`. Recurses one level into subdirectories. Defaults the `name` field to the filename if omitted. |
| `checker.go` | `Check()` evaluates 15 registered check functions (covering 16 of the 18 expectation fields; `min_results` and `max_results` share one check). `CheckWithSnapshots()` adds `file_unchanged` comparison. `CheckProgress()` checks `min_progress` notification counts. Also contains `jsonPathLookup()` for simple `$.dot.path[N]` queries. |
| `trajectory.go` | `CheckTrajectory()` evaluates the 4 trajectory assertion types (order, presence, absence, args_contain) against a trace of tool calls. `LoadAuditLog()` parses JSONL files into trace entries. |
| `sampling_types.go` | `SamplingAssertBlock` type for assertions that test tools which trigger server-side LLM sampling. |
| `logging_types.go` | `LoggingAssertBlock`, `LoggingExpect`, and `LogMessage` types for assertions that test log level setting and message capture. |
| `logging_checker.go` | Logging-specific assertion checking logic. |

### `internal/runner/` (execution engine)

This package contains all the execution logic: CLI flag parsing, server lifecycle, assertion routing, fixture management, and every CLI command.

| File | Responsibility |
|------|----------------|
| `audit.go` | `Audit()`: zero-config quality audit. Connects to a server, discovers tools via `tools/list`, calls each with schema-generated inputs, classifies results (healthy/crash/timeout), reports a quality score, optionally generates starter YAML files. |
| `commands.go` | `Run()`, `Matrix()`, `CI()`: CLI entry points that parse flags, load suites, iterate assertions, collect results, and trigger reporting. |
| `runner.go` | Package doc comment only (the actual runner logic is in `execute.go` and `commands.go`). |
| `execute.go` | `runAssertion()`: the core execution function. Routes to the correct handler based on which block is present. Contains inline logic for the default `assert:` (tool call) path, plus `runResourceAssertion`, `runPromptAssertion`, `runCompletionAssertion`, and `runTrajectoryAssertion`. |
| `client.go` | `createMCPClient()`: transport selection (stdio/SSE/HTTP), Docker wrapping, and `createStdioClientWithCapabilities()` for bidirectional handlers. Also defines the static handler types for roots, sampling, and elicitation. |
| `substitute.go` | `substituteAll()`: recursive template replacement for `{{fixture}}` and captured variables in tool arguments. Also `extractJSONPath()` for pulling values from setup responses. |
| `fixture.go` | `isolateFixture()` and `copyDir()`: per-assertion fixture directory copying to a temp location. |
| `coverage.go` | `Coverage()` command: starts the server, calls `tools/list`, compares against assertion tool names, reports coverage percentage. |
| `generate.go` | `Generate()` command: connects to a server, queries `tools/list`, and writes stub YAML assertion files. |
| `init.go` | `Init()` command: scaffolds a template assertion directory, or generates a complete suite with `--server`. |
| `snapshot.go` | `Snapshot()` command: captures tool responses for regression comparison, similar to Jest snapshot testing. |
| `watch.go` | `Watch()` command: polls YAML files for changes and reruns assertions, showing unified diffs when assertion status flips. |
| `intercept.go` | `Intercept()` command: proxies stdio between an agent and MCP server, capturing tool calls for live trajectory validation. |
| `sampling.go` | `runSamplingAssertion()`: handles assertions that test tools triggering server-side `sampling/createMessage`. |
| `logging.go` | `runLoggingAssertion()`: handles `logging/setLevel` plus `notifications/message` capture. |
| `fix.go` | `--fix` mode: `ScanNearbyPositions()` tries nearby line/column values when position-sensitive assertions fail, and generates YAML patch suggestions. |
| `util.go` | Shared helpers: `writeReports()`, `applyServerOverride()`, `countFails()`, `countPasses()`, `extractText()`. |

### `internal/report/` (output formatting)

This package consumes `[]assertion.Result` and produces output in various formats. It depends on the `assertion` package for types but nothing else. All write errors go to stderr and do not fail the run.

| File | Responsibility |
|------|----------------|
| `audit.go` | `PrintAuditHeader()`, `PrintAuditResults()`, `PrintAuditSummary()`, `PrintAuditNextSteps()`: audit-specific report formatting with quality score and CI guidance. |
| `report.go` | `PrintResults()`: terminal table with color (TTY) or plain text (pipe/CI). `PrintMatrix()`: cross-language comparison table. |
| `color.go` | ANSI color codes, TTY detection (`os.ModeCharDevice`), `NO_COLOR` env var support, progress indicator on stderr. |
| `diff.go` | `FormatDiff()`, `FormatStatusChange()`: unified diff output for the `watch` command when an assertion's status changes. |
| `junit.go` | JUnit XML serialization via `encoding/xml`. Includes pass@k/pass^k properties when `--trials > 1`. |
| `markdown.go` | GitHub Step Summary table. Includes a reliability section when `--trials > 1`. Auto-detects `$GITHUB_STEP_SUMMARY` in CI mode. |
| `badge.go` | shields.io endpoint JSON (`schemaVersion`, `label`, `message`, `color`). |
| `reliability.go` | pass@k (passed at least once in k trials) and pass^k (passed every time in k trials) computation. |
| `baseline.go` | Baseline JSON write/load and regression detection. Only PASS-to-non-PASS transitions count as regressions. |
| `coverage.go` | Coverage JSON serialization for the `coverage` command. |
| `snapshot.go` | Snapshot file read/write/compare for the `snapshot` command. |

### Package dependency graph

```
cmd/mcp-assert/main.go
  └── internal/runner
        ├── internal/assertion   (types, loader, checker)
        ├── internal/report      (all output formats)
        ├── mark3labs/mcp-go/client  (MCP transport: stdio, SSE, streamable HTTP)
        └── mark3labs/mcp-go/mcp     (MCP protocol types)
```

No circular dependencies. The `assertion` and `report` packages do not import each other. `report` depends on `assertion` for the `Result` and `Status` types. Neither package imports `runner`.

---

## Key Abstractions

### Suite and Assertion

A `Suite` is a collection of `Assertion` values loaded from a directory. Each `Assertion` represents a single test: connect to a server, optionally run setup steps, make a request, check the response.

```go
type Suite struct {
    Assertions []Assertion
    Dir        string        // directory the suite was loaded from
}

type Assertion struct {
    Name             string
    Server           ServerConfig           // how to connect
    Setup            []ToolCall             // prerequisite tool calls
    Assert           AssertBlock            // the tool call to test (default block)
    AssertResources  *ResourceAssertBlock   // or test resources
    AssertPrompts    *PromptAssertBlock     // or test prompts
    AssertCompletion *CompletionAssertBlock // or test completion
    AssertSampling   *SamplingAssertBlock   // or test sampling
    AssertLogging    *LoggingAssertBlock    // or test logging
    Trace            []TraceEntry           // or validate a tool call trace
    AuditLog         string                 // path to JSONL audit log (alternative to Trace)
    Trajectory       []TrajectoryAssertion  // trajectory checks (no server)
    Timeout          string
    Skip             bool
    SkipUnlessEnv    string                 // skip if this env var is not set
}
```

Exactly one block type is active per assertion. The runner checks them in priority order (trajectory first, then resources, prompts, completion, sampling, logging, and finally the default `assert:` block).

### ServerConfig

Describes how to connect to the MCP server under test.

```go
type ServerConfig struct {
    Command            string             // executable to launch (stdio)
    Args               []string           // arguments to the command
    Env                map[string]string  // environment variables
    Transport          string             // "stdio", "sse", or "http"
    URL                string             // endpoint for sse/http
    Headers            map[string]string  // custom headers for sse/http (supports ${VAR} expansion)
    Docker             string             // Docker image for container isolation (stdio only)
    ClientCapabilities ClientCapabilities // mock bidirectional responses
}
```

### Expect

The `Expect` struct holds all possible expectations. You set only the fields you care about; unset fields are skipped during checking.

```go
type Expect struct {
    Contains        []string          // response must contain these substrings
    ContainsAny     []string          // response must contain at least one of these
    NotContains     []string          // response must not contain these
    Equals          *string           // exact match (whitespace-trimmed)
    JSONPath        map[string]any    // $.field.path must equal expected value
    MinResults      *int              // array must have at least N items
    MaxResults      *int              // array must have at most N items
    NotEmpty        *bool             // response must not be empty/null/[]/{}
    NotError        *bool             // isError flag must be false
    IsError         *bool             // isError flag must be true
    MatchesRegex    []string          // response must match all patterns
    FileContains    map[string]string // file on disk must contain text
    FileNotContains map[string]string // file on disk must NOT contain text
    FileNotExists   []string          // file must NOT exist on disk
    FileUnchanged   []string          // file on disk must not have changed
    NetDelta        *int              // net_delta field must equal N
    InOrder         []string          // substrings must appear in this order
    MinProgress     *int              // at least N progress notifications
}
```

The checker evaluates expectations in this fixed order:

1. `not_error` / `is_error` (check the isError flag)
2. `not_empty` (reject empty, null, `[]`, `{}`)
3. `equals` (exact match, whitespace-trimmed)
4. `contains` / `contains_any` / `not_contains` (substring checks)
5. `matches_regex` (compiled regex matching)
6. `json_path` (dot-notation lookup on parsed JSON)
7. `min_results` / `max_results` (array length bounds)
8. `net_delta` (numeric field comparison)
9. `file_contains` / `file_not_contains` (read file from disk, check content)
10. `file_not_exists` (verify file does not exist on disk)
11. `in_order` (ordered substring search)

`file_unchanged` is handled separately by `CheckWithSnapshots()`. `min_progress` is handled by `CheckProgress()` after the main check.

First failure short-circuits: only the first failing expectation is reported.

### Result

The outcome of running a single assertion.

```go
type Result struct {
    Name     string      // assertion name
    Status   Status      // "PASS", "FAIL", or "SKIP"
    Detail   string      // error message on failure
    Duration DurationMS  // wall-clock time (serializes as integer milliseconds in JSON)
    Language string      // set in matrix mode
    Trial    int         // trial number when --trials > 1
}
```

Results flow from the runner to the report package. Every output format (terminal, JUnit, markdown, badge, JSON) consumes the same `[]Result` slice.

---

## Transport Layer

### How mcp-assert connects to servers

The `createMCPClient()` function in `client.go` is the single point where transport selection happens. It reads the `transport` field from the assertion's `ServerConfig` and creates the appropriate mcp-go client:

**stdio (default).** The function calls `client.NewStdioMCPClient(command, env, args...)` from the mcp-go library. This launches the server command as a child process and wires stdin/stdout for JSON-RPC. If `--docker` is set, the command is rewritten to `docker run --rm -i -v fixture:fixture <image> <original-command>`, so Docker's `-i` flag provides the same bidirectional pipe transport.

**SSE.** The function calls `client.NewSSEMCPClient(url)`. The client connects to the server's SSE endpoint for receiving responses and sends requests over a standard HTTP POST channel.

**Streamable HTTP.** The function calls `client.NewStreamableHttpClient(url)`. This is the modern remote transport where both requests and responses use HTTP with streaming.

All three return the same `client.MCPClient` interface, so the rest of the runner is transport-agnostic. After creation, the runner calls `Initialize()` on the client to perform the MCP handshake.

### Client capabilities (bidirectional path)

When `client_capabilities` is configured in the YAML, the stdio transport takes a different construction path. Instead of the convenience `NewStdioMCPClient`, the runner uses the lower-level `client.NewClient` with explicit handler options:

1. Create a raw `StdioTransport` and start it.
2. Build handler options: `WithRootsHandler`, `WithSamplingHandler`, `WithElicitationHandler`.
3. Call `client.NewClient(transport, opts...)` to create the client with handlers registered.
4. Call `c.Start(ctx)` to activate the bidirectional channel.

This ordering is critical. If handlers are registered after `Start`, the server's `roots/list` or `sampling/createMessage` requests would arrive before handlers exist, causing errors.

The three static handlers are simple:

- **`staticRootsHandler`**: returns a fixed list of filesystem roots (with `{{fixture}}` substituted).
- **`staticSamplingHandler`**: returns a mock LLM response with configurable text, model name, and stop reason.
- **`staticElicitationHandler`**: returns preset form values, with support for accept/decline/cancel actions.

---

## Fixture Isolation

### The problem

MCP servers often modify files on disk. A filesystem server might create, edit, or delete files. A language server might apply refactoring edits. If two assertions share the same fixture directory, one assertion's side effects can break subsequent assertions or produce non-deterministic results.

### The solution

Before each assertion executes, `isolateFixture()` (in `fixture.go`) copies the entire fixture directory to a unique temporary directory. The assertion receives the path to this copy. After the assertion finishes (pass or fail), the temporary directory is deleted.

```
Original fixture:  ./test-data/
                        ↓ (copy)
Temp copy:         /tmp/mcp-assert-fixture-abc123/test-data/
                        ↓ (used by assertion)
                        ↓ (deleted after assertion)
```

The `{{fixture}}` template in YAML arguments is replaced with the temp copy path, not the original. This means any file paths the server sees point to the disposable copy.

### When isolation is skipped

- If no `--fixture` is provided, there is nothing to isolate.
- If `--docker` is used, Docker already provides isolation through fresh containers, so copying is redundant.

### Implementation

`copyDir()` recursively walks the source directory, preserving file permissions and directory structure. Symlinks are not followed. The copy target is placed inside the temp directory with the same base name as the original, so relative paths within the fixture remain valid.

---

## Block Types

Each YAML assertion file uses exactly one block type. The block type determines which MCP protocol method is called and how the response is processed.

### 1. `assert:` (tool calls)

The default and most common block. Calls `tools/call` with a named tool and JSON arguments.

```yaml
assert:
  tool: read_file
  args:
    path: "{{fixture}}/example.txt"
  capture_progress: true    # optional: collect notifications/progress
  expect:
    not_error: true
    contains: ["hello world"]
```

The optional `capture_progress: true` field registers a notification listener before the tool call that counts `notifications/progress` messages. This enables the `min_progress` expectation.

### 2. `assert_prompts:` (prompt listing and retrieval)

Tests the `prompts/list` and `prompts/get` MCP methods. Set exactly one of `list` or `get`.

```yaml
assert_prompts:
  get:
    name: "code_review"
    arguments:
      language: "go"
  expect:
    contains: ["review"]
```

For `prompts/get`, the response text is built by joining the prompt's description and message contents.

### 3. `assert_resources:` (resource listing and reading)

Tests `resources/list`, `resources/read`, and resource subscriptions.

```yaml
assert_resources:
  read: "test://static/resource"
  expect:
    not_empty: true
```

Supports `subscribe`/`unsubscribe` fields and `expect_notification` to verify resource update notifications arrive.

### 4. `assert_completion:` (autocompletion)

Tests `completion/complete` for argument autocompletion on prompts or resources.

```yaml
assert_completion:
  ref:
    type: "ref/prompt"
    name: "complex_prompt"
  argument:
    name: "style"
    value: ""
  expect:
    contains: ["formal"]
```

### 5. `assert_sampling:` (sampling-triggered tools)

Tests tools that cause the server to request an LLM completion via `sampling/createMessage`. The block configures both the tool call and the mock LLM response in one place.

```yaml
assert_sampling:
  tool: ask_llm
  args:
    question: "What is the capital of France?"
  mock_text: "The capital of France is Paris."
  mock_model: mock-gpt
  expect:
    not_error: true
    contains: ["Paris"]
```

This is a convenience wrapper. You can achieve the same result with `assert:` plus `client_capabilities.sampling`, but `assert_sampling` keeps the mock and assertion together.

### 6. `assert_logging:` (log level and message capture)

Tests `logging/setLevel` and captures `notifications/message` log events during a tool call.

```yaml
assert_logging:
  set_level: debug
  tool: echo
  args:
    message: "test"
  expect:
    min_messages: 1
    contains_level: ["debug"]
    contains_data: ["test"]
```

The runner first calls `logging/setLevel`, then executes the tool while listening for `notifications/message`. The logging-specific `expect` fields (`min_messages`, `contains_level`, `contains_data`) are checked by a dedicated logging checker.

### 7. `trajectory:` (tool call sequence validation)

Validates a sequence of tool calls without starting any server. The trace comes from inline YAML (`trace:` field) or an external JSONL audit log (`audit_log:` field).

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

Four trajectory assertion types are available:

| Type | What it checks |
|------|----------------|
| `order` | Listed tools appear in this sequence (not necessarily adjacent) |
| `presence` | All listed tools appear at least once |
| `absence` | None of the listed tools appear |
| `args_contain` | A specific tool was called with specific argument values (partial match) |

### Routing logic

In `execute.go`, `runAssertion()` checks block types in priority order using nil checks on the optional block pointers:

```
trajectory → assert_resources → assert_prompts → assert_completion
→ assert_sampling → assert_logging → assert (default)
```

Each handler follows the same pattern: validate inputs, create MCP client, initialize, run setup, execute the protocol-specific call, check expectations, return result.

---

## Reporting Pipeline

Results flow through a simple pipeline. The runner collects `[]assertion.Result` from all assertions, then passes the slice to each output sink.

### Terminal output (`report.go`, `color.go`)

Always produced. Each assertion prints as a single line: status icon, name, and duration. Failed assertions include the error detail on the next line. A summary line at the end shows total/passed/failed/skipped counts.

Color behavior adapts to the environment:

| Condition | Behavior |
|-----------|----------|
| stdout is a TTY | Green checkmarks, red Xs, ANSI color |
| stdout is a pipe | Plain `PASS`/`FAIL`/`SKIP`, no escape codes |
| `NO_COLOR=1` set | Plain output regardless of TTY |
| `TERM=dumb` | Plain output |

A progress indicator (`[3/21] assertion name`) prints to stderr during execution, so it does not interfere with piped stdout.

### JUnit XML (`junit.go`)

Standard JUnit format consumed by CI systems (GitHub Actions, Jenkins, GitLab). Each assertion becomes a `<testcase>`. Failed assertions include `<failure>` elements with the error detail. When `--trials > 1`, pass@k and pass^k metrics are included as `<property>` elements.

### Markdown (`markdown.go`)

A GitHub Step Summary table. The `ci` command auto-detects `$GITHUB_STEP_SUMMARY` and writes to it. When `--trials > 1`, a reliability section is appended. The table includes assertion name, status, and duration.

### Badge JSON (`badge.go`)

A shields.io endpoint JSON file (`schemaVersion`, `label`, `message`, `color`). Host the file via GitHub Pages to get a live pass-rate badge in your README.

### Raw JSON (`--json`)

The full `[]Result` array serialized as JSON to stdout. Useful for programmatic consumption.

### Reliability metrics (`reliability.go`)

When `--trials N` is used with N > 1, each assertion runs N times. Two metrics are computed:

- **pass@k**: passed at least once in k trials (measures capability).
- **pass^k**: passed every time in k trials (measures reliability).

These appear in terminal output, JUnit XML properties, and the markdown table.

### Baseline and regression detection (`baseline.go`)

`--save-baseline results.json` persists the current results. `--baseline results.json` on a subsequent run compares against the saved baseline. Only PASS-to-non-PASS transitions are flagged as regressions. Previously-failing tests that still fail are not regressions. New tests not in the baseline are not regressions.

### Best-effort writes

All report outputs are best-effort. If writing a JUnit file fails, the error goes to stderr but the run itself does not fail. This prevents flaky CI permissions from blocking test results.

---

## Extension Points

### Adding a new assertion type

To add a new expectation (like `max_length` to check response string length):

1. **Add the field to `Expect`** in `internal/assertion/types.go`:
   ```go
   MaxLength *int `yaml:"max_length"`
   ```
2. **Add the check to `Check()`** in `internal/assertion/checker.go`, placing it in the appropriate position in the evaluation order.
3. **Add unit tests** in `internal/assertion/checker_test.go`.

The checker is pure: it takes a string and returns an error. No I/O, no state. This makes new assertion types trivially testable.

### Adding a new block type

To add a new block type (like `assert_notifications:` for testing arbitrary server notifications):

1. **Define the block struct** in `internal/assertion/types.go`:
   ```go
   type NotificationAssertBlock struct { ... }
   ```
2. **Add the field to `Assertion`**:
   ```go
   AssertNotifications *NotificationAssertBlock `yaml:"assert_notifications,omitempty"`
   ```
3. **Add a handler function** in a new file `internal/runner/notifications.go`:
   ```go
   func runNotificationAssertion(a assertion.Assertion, ...) assertion.Result { ... }
   ```
4. **Add the routing check** in `runAssertion()` in `execute.go`, following the existing priority pattern:
   ```go
   if a.AssertNotifications != nil {
       return runNotificationAssertion(a, fixture, timeout, dockerImage, start)
   }
   ```

### Adding a new CLI command

1. **Add the function** in a new file in `internal/runner/` (e.g., `mycommand.go`):
   ```go
   func MyCommand(args []string) error { ... }
   ```
2. **Add the dispatch** in `cmd/mcp-assert/main.go`:
   ```go
   case "mycommand":
       if err := runner.MyCommand(os.Args[2:]); err != nil { ... }
   ```
3. **Update `printUsage()`** with the new command's documentation.

### Adding a new output format

1. **Add a file** in `internal/report/` (e.g., `csv.go`) with a function that consumes `[]assertion.Result`.
2. **Add a flag** in `commands.go` for the new format.
3. **Call the new function** from `writeReports()` in `util.go`.

---

## Concurrency Model

mcp-assert uses goroutines in several places, but assertions themselves run sequentially. Understanding where concurrency exists and where it doesn't prevents bugs when contributing.

### Suite execution: sequential

Assertions within a suite run one at a time, in file-name order. There is no parallelism at the assertion level. This is a deliberate choice:

- Fixture isolation copies a directory per assertion. Parallel copies of the same fixture directory would race on filesystem operations.
- Terminal output (progress indicator, results table) assumes sequential printing.
- Server subprocess management (start, initialize, call, close) is not designed for interleaving.

To parallelize, run separate `mcp-assert run` processes on different suite directories. The GitHub Action supports this via CI matrix strategies.

### Per-assertion: goroutines for I/O

Within a single assertion, the mcp-go library uses goroutines internally for transport I/O (reading from stdin, writing to stdout). These are managed by the library and cleaned up when the client is closed. mcp-assert does not spawn its own goroutines during normal assertion execution.

The exception is `capture_progress: true`, which registers a notification listener that increments an `atomic.Int32` counter from the mcp-go callback goroutine. The counter is read on the main goroutine after the tool call completes. This is safe because `atomic` operations don't require a mutex.

### Intercept command: explicit goroutines

The `intercept` command is the most concurrent part of the codebase. It spawns:

1. **Agent-to-server goroutine**: reads lines from os.Stdin via a buffered scanner, inspects each for `tools/call` requests, and forwards to the server's stdin pipe.
2. **Server-to-agent goroutine**: copies from the server's stdout pipe to os.Stdout.
3. **Main goroutine**: waits for both I/O goroutines to finish (or a timeout), then validates trajectory assertions.

Shared state:

```
trace []TraceEntry    ← appended by goroutine 1, read by main goroutine
                        protected by sync.Mutex
cmd   *exec.Cmd       ← written by goroutine (proxyStdio), read by main
                        on timeout for Process.Kill()
                        safe because written before done channel send
```

On timeout, the main goroutine kills the server process (`cmd.Process.Kill()`), which causes the I/O goroutines to unblock and exit. The main goroutine then waits for the proxy goroutine to finish before reading the trace.

### Watch command: polling loop

The `watch` command polls for file changes in a `for` loop with `time.Sleep`. It does not use goroutines for file watching (no `fsnotify`). When a change is detected, it re-runs the suite synchronously on the main goroutine. There is no concurrent execution of assertions.

### What is NOT safe to parallelize

- **Assertions within a suite.** Fixture isolation, terminal output, and server lifecycle all assume sequential execution.
- **Setup steps within an assertion.** Steps execute in order because later steps may depend on captured variables from earlier steps.
- **Checker evaluation.** The checker short-circuits on first failure. Parallel evaluation would produce non-deterministic error messages.

### What IS safe to parallelize

- **Separate suite directories.** Each `mcp-assert run` process is fully independent.
- **Unrelated tool calls in separate assertions.** No shared state between assertions (each gets its own server, fixture, and context).

---

## Key Design Decisions

**One server per assertion.** Each assertion starts a fresh MCP server subprocess. This prevents state leakage between tests but means server startup cost is paid per assertion. For fast servers (filesystem, memory) this is negligible. For slow servers (gopls, jdtls) it dominates test duration. The `setup` block amortizes some of this by allowing warmup calls within a single assertion's server lifetime.

**Checker is pure.** `Check()` takes a string and returns an error. No I/O, no state, no side effects. `CheckWithSnapshots()` adds file comparison but the snapshots are passed in, not read internally. This makes the checker trivially testable.

**Transport is pluggable.** `createMCPClient` is a single function that all execution paths use. All three transports return the same `MCPClient` interface, so the runner never needs to know which transport is active.

**Docker is a command wrapper.** `--docker <image>` does not use the Docker SDK. It rewrites the server command to `docker run --rm -i ...`. Since MCP uses stdio, Docker's `-i` flag provides bidirectional pipe transport. The server runs inside the container; assertions run outside. Docker is only supported with stdio transport.

**Color degrades gracefully.** TTY detection via `os.ModeCharDevice`. `NO_COLOR` env var. `TERM=dumb`. In CI (pipes), output is plain `PASS`/`FAIL`/`SKIP` with no escape codes.

**Setup tools are not counted as "tested" by coverage.** The `coverage` command only counts the `assert.tool` field, not tools that appear in `setup:` blocks. This correctly reflects that setup tools are prerequisites, not the subject of the test.

---

## Error Model

MCP defines two distinct error layers. Understanding the difference is critical for writing correct assertions and for interpreting audit results.

### Application errors (`isError: true`)

The tool ran, understood the request, and is reporting a failure. The JSON-RPC response is a normal `result` with the `isError` flag set to `true`:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{ "type": "text", "text": "File not found: /tmp/missing.txt" }],
    "isError": true
  }
}
```

The agent receives the error message and can reason about it: retry with a different path, ask the user for help, or try an alternative approach. This is the correct way for MCP tools to report failures.

mcp-assert checks this with `not_error: true` (expects `isError` to be false) or `is_error: true` (expects `isError` to be true for negative tests).

### Protocol errors (JSON-RPC error codes)

The server crashed, panicked, or threw an unhandled exception. Instead of a `result`, the response is a JSON-RPC `error` object:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "TypeError: Cannot read properties of undefined"
  }
}
```

The agent receives an opaque error code with no actionable message. It cannot distinguish "file not found" from "server crashed" from "out of memory." The `-32603` code (Internal Error) is the most common: it means the tool handler threw an exception instead of catching it and returning `isError: true`.

### How errors flow through mcp-assert

```
Server response
       │
       ├── JSON-RPC error (-32603, -32600, etc.)
       │     └── Assertion FAILS with: "server returned error: Internal error"
       │
       ├── Result with isError: true
       │     ├── If expect.not_error: true  → FAIL ("expected no error, got isError")
       │     └── If expect.is_error: true   → PASS (negative test)
       │
       └── Result with isError: false
             ├── If expect.is_error: true   → FAIL ("expected error, got success")
             └── Continue to content checks (contains, equals, json_path, etc.)
```

### Transport errors

A third category exists below the JSON-RPC layer: the connection itself fails. The server process exits before responding, the SSE connection drops, or the HTTP request times out. These produce failures like "transport error: transport closed" or "context deadline exceeded." mcp-assert reports these as assertion failures with the transport error as the detail message.

### The audit classification

The `audit` command uses the error model to classify each tool into three categories:

| Classification | Condition | Meaning |
|----------------|-----------|---------|
| **Healthy** | `isError: false` or `isError: true` | Tool handled the input correctly (success or reported failure) |
| **Crashed** | JSON-RPC error (-32603) | Tool threw an unhandled exception |
| **Timed out** | Context deadline exceeded | Tool did not respond within the timeout |

A tool returning `isError: true` with a helpful error message is **healthy**, not crashed. The audit quality score reflects this: servers that catch errors and return `isError: true` get 100% even if every tool "fails," because the failures are handled correctly.

---

## Template Engine

mcp-assert replaces template variables in tool arguments, setup step arguments, and certain configuration fields before sending requests to the server. Three substitution types are supported.

### `{{fixture}}` substitution

Replaced with the absolute path to the isolated fixture directory (or the original fixture path if isolation is disabled). This runs first, before any other substitution.

```yaml
assert:
  tool: read_file
  args:
    path: "{{fixture}}/hello.txt"
```

After substitution: `path: "/tmp/mcp-assert-fixture-abc123/test-data/hello.txt"`

### `{{variable}}` capture substitution

Setup steps can capture values from tool responses using the `capture` field. Captured values are substituted into subsequent setup steps and the main assertion's arguments.

```yaml
setup:
  - tool: create_item
    args:
      name: "test"
    capture:
      item_id: "$.id"
  - tool: get_item
    args:
      id: "{{item_id}}"
```

The `capture` field maps variable names to JSON path expressions. After the first setup step runs, the response is parsed as JSON, `$.id` is extracted, and `{{item_id}}` is available for the rest of the assertion.

### `${ENV}` expansion in environment variables

Server environment variables support shell-style variable expansion:

```yaml
server:
  command: my-server
  env:
    API_KEY: "${API_KEY}"
    DB_URL: "${DATABASE_URL:-sqlite:///tmp/test.db}"
```

This uses Go's `os.ExpandEnv`, which supports `$VAR`, `${VAR}`, and default values via `${VAR:-default}`. Expansion happens at runtime, not at YAML parse time.

### Substitution order

1. **Fixture substitution**: all `{{fixture}}` tokens are replaced in tool arguments.
2. **Variable substitution**: `{{captured_var}}` tokens are replaced with values from prior setup steps.
3. **Environment expansion**: `${VAR}` patterns in `server.env` values are expanded from the process environment.

Substitution is recursive for tool arguments (nested maps and arrays are walked), but not for the `server` block fields (command, args) except `server.env` values and `server.args` (which get fixture substitution only).

### Implementation

`substituteAll()` in `internal/runner/substitute.go` walks the argument map recursively. For each string value, it replaces `{{fixture}}` first, then replaces any `{{key}}` where `key` matches a captured variable name. The function also handles `substituteFixture()` for the simpler case where only fixture substitution is needed.

`extractJSONPath()` in the same file implements the `$.field.subfield[N]` path syntax used by `capture`. It walks parsed JSON using dot notation, with bracket syntax for array indices. It does not support full JSONPath (no wildcards, no filters, no recursive descent).

---

## Audit Command Flow

The `audit` command is a distinct execution path from `run`. It requires no YAML files: just point it at a server and it discovers and tests everything automatically.

### Lifecycle

```
mcp-assert audit --server "npx my-server" --timeout 15s
```

1. **Parse the server spec.** `strings.Fields("npx my-server")` splits into command and args.

2. **Connect and initialize.** Same `createMCPClient` + `Initialize` handshake as `run`.

3. **Discover tools.** Send `tools/list` request. The server responds with its full tool catalog, including JSON Schema for each tool's input parameters.

4. **Generate inputs.** For each tool, the audit command generates a plausible input from the tool's JSON Schema. For required string fields, it uses the field name as the value (e.g., `{"path": "path"}`). For numbers, it uses `0`. For booleans, `false`. For arrays, an empty array. If no schema is provided, it sends an empty object.

5. **Call each tool.** Send `tools/call` for every discovered tool, one at a time. Capture the response and the `isError` flag, and measure the duration.

6. **Classify results.** Each tool is classified as healthy (normal response or `isError: true`), crashed (JSON-RPC error), or timed out (deadline exceeded).

7. **Report.** Print a quality score (percentage of tools that are healthy), a per-tool result table, and next-steps guidance for fixing crashed tools.

8. **Optionally generate YAML stubs.** With `--output <dir>`, write a starter assertion YAML file for each discovered tool. The generated YAML includes the tool name, the schema-derived arguments, and basic expectations (`not_error: true`, `not_empty: true`).

### Why audit exists

Most MCP servers have never been tested by an external client. The most common failure mode is tools that throw exceptions instead of returning `isError: true`. The audit command finds these in seconds without any setup. Every bug filed by the mcp-assert project was discovered via `audit` first, then confirmed with a targeted assertion YAML.

---

## Snapshot Testing Flow

The `snapshot` command captures tool responses as golden files for regression detection. This is conceptually similar to Jest's snapshot testing.

### Update mode (`--update`)

```
mcp-assert snapshot --suite evals/ --server "npx my-server" --update
```

1. Load the suite and iterate assertions.
2. For each assertion, start the server, run setup, call the tool.
3. Write the response text to a snapshot file alongside the YAML: `evals/echo.yaml` produces `evals/.snapshots/echo.snap`.
4. Report how many snapshots were created, updated, or unchanged.

### Verify mode (default)

```
mcp-assert snapshot --suite evals/ --server "npx my-server"
```

1. Load the suite, run each assertion, capture the response.
2. Compare the response against the stored snapshot file.
3. If the response differs, report the diff and fail.
4. If no snapshot exists, report it as a new (uncaptured) snapshot.

### Diff display

When a snapshot changes, the runner computes a line-by-line unified diff using an LCS (longest common subsequence) algorithm. The diff is printed with `+`/`-` markers, similar to `git diff`. The diff implementation is in `internal/report/diff.go`.

### When to use snapshots vs expectations

Snapshots are useful when you want to detect any change in a tool's output, even changes you did not anticipate. Expectations (`contains`, `json_path`, etc.) are useful when you want to check specific properties and ignore everything else. Snapshots are stricter but more brittle; expectations are more targeted but may miss unexpected changes.

---

## Security Model

mcp-assert runs arbitrary server commands as subprocesses. Understanding the trust boundaries matters.

### Trusted input: YAML files

Assertion YAML files are fully trusted. They specify commands to execute, arguments to pass, and files to read. A malicious YAML file can run arbitrary code via the `server.command` field. This is by design: the YAML files are part of your test suite, committed to version control, and reviewed like any other code.

### Untrusted input: server responses

Server responses are untrusted. A malicious or buggy server could return arbitrary text, but this text is only used for:

- String comparison against expected values (no execution)
- JSON parsing for `json_path` checks (using Go's `encoding/json`, which is safe)
- Display in terminal output (ANSI escape injection is possible but limited to the user's own terminal)

Server responses are never passed to `exec`, `eval`, or used to construct file paths for writing.

### Command injection surface

The `--server` CLI flag and the `server.command` YAML field specify shell commands. These are split using `strings.Fields()`, not passed through a shell. This means shell metacharacters (`&&`, `|`, `;`, `$()`) are treated as literal arguments, not interpreted. A `--server` value of `my-server; rm -rf /` would try to launch a binary literally named `my-server;` with arguments `rm`, `-rf`, `/`.

However, environment variable expansion in `server.env` uses `os.ExpandEnv`, which resolves `${VAR}` from the process environment. If environment variables contain malicious values, those values are passed to the server process. This is standard behavior for any tool that forwards environment variables.

### Fixture path traversal

The `{{fixture}}` template is replaced with an absolute path. The fixture isolation mechanism copies the entire directory to a temp location, so path traversal attempts (e.g., `{{fixture}}/../../etc/passwd`) resolve to paths inside the temp copy, not the original filesystem. However, if fixture isolation is disabled (no `--fixture` flag), there is no protection: the server sees whatever paths the YAML specifies.

### Docker isolation

When `--docker <image>` is used, the server runs inside a fresh container. The fixture directory is mounted read-write via `-v`. The container is destroyed after the assertion (`--rm`). This provides filesystem and process isolation but not network isolation (the container shares the host network by default).

---

## Performance Characteristics

### What dominates test duration

For most assertion suites, **server startup time** dominates. Each assertion starts a fresh server process. Typical startup times:

| Server type | Startup time | Example |
|-------------|-------------|---------|
| Node.js MCP servers | 200-500ms | filesystem, memory, fetch |
| Python MCP servers | 300-800ms | mcp-server-time, sqlite |
| Go MCP servers | 10-50ms | agent-lsp, grafana-mcp |
| JVM MCP servers | 2-5s | Spring AI servers |
| Docker-wrapped servers | 1-3s | Any server + `--docker` |

A suite of 25 assertions against a Node.js server takes roughly 10-15 seconds. The same suite against a Go server takes 2-3 seconds. Tool call time is usually negligible (under 100ms) unless the tool does real work (network requests, file I/O).

### What's cheap

- **YAML parsing**: ~1ms per file. Negligible even for 100+ assertions.
- **Assertion checking**: pure string/JSON operations, under 1ms per assertion.
- **Report generation**: under 10ms for all formats combined.
- **Fixture copying**: depends on fixture size, but typically under 50ms for small test data.

### Optimization strategies

**Reduce server restarts.** Use `setup` steps to perform multiple operations within a single assertion's server lifetime. For example, instead of two assertions (one to create a file, one to read it), use one assertion with a setup step that creates the file and an assert block that reads it.

**Use `skip_unless_env` for slow tests.** Auth-gated assertions that call real APIs should be skipped in fast CI runs and only enabled when credentials are available.

**Use `--trials 1` in CI.** The `--trials N` flag runs each assertion N times for reliability measurement, multiplying total duration. Use `--trials 1` for pass/fail CI and reserve `--trials 5+` for dedicated reliability analysis.

**Parallelize at the suite level.** mcp-assert itself runs assertions sequentially within a suite. To parallelize, split your assertions into multiple suite directories and run separate `mcp-assert run` processes. The GitHub Action supports this via matrix strategies.

---

## Wire Format Examples

These are the actual JSON-RPC messages exchanged between mcp-assert and the server for each block type. Understanding the wire format helps when debugging assertion failures or writing new block types.

### Tool call (`assert:`)

Request:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "read_file",
    "arguments": {
      "path": "/tmp/mcp-assert-fixture-abc123/test-data/hello.txt"
    }
  }
}
```

Response (success):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      { "type": "text", "text": "Hello, world!\n" }
    ],
    "isError": false
  }
}
```

Response (application error):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      { "type": "text", "text": "File not found: /tmp/missing.txt" }
    ],
    "isError": true
  }
}
```

### Resource read (`assert_resources:`)

Request:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "resources/read",
  "params": {
    "uri": "test://static/greeting"
  }
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "contents": [
      {
        "uri": "test://static/greeting",
        "mimeType": "text/plain",
        "text": "Hello from the resource"
      }
    ]
  }
}
```

### Prompt get (`assert_prompts:`)

Request:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "prompts/get",
  "params": {
    "name": "code_review",
    "arguments": { "language": "go" }
  }
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "description": "Review Go code for best practices",
    "messages": [
      {
        "role": "user",
        "content": { "type": "text", "text": "Review the following Go code..." }
      }
    ]
  }
}
```

### Completion (`assert_completion:`)

Request:
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "completion/complete",
  "params": {
    "ref": { "type": "ref/prompt", "name": "code_review" },
    "argument": { "name": "language", "value": "py" }
  }
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "completion": {
      "values": ["python", "pytorch"],
      "hasMore": false
    }
  }
}
```

### Logging (`assert_logging:`)

Set level request:
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "logging/setLevel",
  "params": { "level": "debug" }
}
```

During the subsequent tool call, the server sends log notifications:
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/message",
  "params": {
    "level": "debug",
    "logger": "my-server",
    "data": "Processing request..."
  }
}
```

mcp-assert captures these notifications and checks them against `contains_level`, `contains_data`, and `min_messages` expectations.

### Sampling (bidirectional)

During a tool call, the server requests an LLM completion from the client:

Server to client:
```json
{
  "jsonrpc": "2.0",
  "id": 100,
  "method": "sampling/createMessage",
  "params": {
    "messages": [
      { "role": "user", "content": { "type": "text", "text": "Summarize this code" } }
    ],
    "maxTokens": 1000
  }
}
```

Client (mcp-assert) responds with the mock:
```json
{
  "jsonrpc": "2.0",
  "id": 100,
  "result": {
    "role": "assistant",
    "content": { "type": "text", "text": "This code implements a REST API..." },
    "model": "mock",
    "stopReason": "end_turn"
  }
}
```

The server uses the mock response to complete its tool execution and returns the final result to the client.
