# Roadmap

## Ecosystem Credibility

| Item | Priority | Description |
|------|----------|-------------|
| **External adoption** | High | Get one MCP server author to use mcp-assert and report results. |
| **Underserved language suites** | Medium | Server suites for Rust, Java, and C# MCP servers: communities with few or no dedicated testing tools. Rust: `rmcp` SDK. Java: MCP SDK for Spring, Quarkus MCP. C#: modelcontextprotocol/csharp-sdk. High-signal distribution targets. |
| **"Works with mcp-assert" badge** | Medium | A standard badge + registry that MCP server authors can claim after passing a reference suite. Drives discoverability: every badge is a backlink. |
| **Reference suite registry** | Medium | A canonical set of suites that any conforming server can run against, independent of server-specific fixtures. Addresses the duplicate suite problem (agent-lsp and mcp-assert each maintaining their own copy). Single source of truth. |

## Distribution

| Item | Priority | Description |
|------|----------|-------------|
| **Homebrew formula** | High | `brew install mcp-assert` |
| **PyPI wrapper** | High | `pip install mcp-assert`: downloads the Go binary. Python MCP server authors won't `go install`. |
| **npm wrapper** | Medium | `npx mcp-assert`: same pattern, TypeScript audience. |
| **MCP registry integration** | Medium | Surface the mcp-assert test badge prominently on Glama and Smithery listings. Servers that pass a reference suite get a "verified" marker in the registry. |

## Technical Depth

| Item | Priority | Description |
|------|----------|-------------|
| **Automatic fixture isolation** | **Critical** | Per-test fixture copy in the runner so write-tests can never contaminate read-tests. Today any assertion that commits edits to disk modifies the shared fixture directory, shifting line numbers for subsequent tests. Docker already does this per container; the stdio path needs the same guarantee. Single highest-leverage DX improvement. |
| **`--fix` mode for position errors** | High | When a position-sensitive assertion fails with "no identifier found" or "column is beyond end of line", suggest the correct line/column. Re-run scanning nearby positions, emit a suggested YAML patch. |
| **Watch mode diff view** | Medium | When an assertion flips from PASS to FAIL in `--watch` mode, show a diff of expected vs actual response. |
| **Live agent trajectory capture** | Low | Intercept tool calls made by a real agent session in real time (without needing a pre-recorded trace or audit log). Extends the existing `trace:` and `audit_log:` sources. |

## MCP Protocol Surface Coverage

Tracking coverage of every method defined in the MCP 2025-11-25 specification.

### Server features

| Protocol area | Methods | Status |
|--------------|---------|--------|
| **Tools** | `tools/list`, `tools/call`, `notifications/tools/list_changed` | Covered |
| **Resources** | `resources/list`, `resources/read`, `resources/subscribe`, `resources/unsubscribe` | Covered |
| **Prompts** | `prompts/list`, `prompts/get`, `notifications/prompts/list_changed` | Covered |

### Client features (server-initiated requests)

| Protocol area | Methods | Status |
|--------------|---------|--------|
| **Sampling** | `sampling/createMessage` | Covered (`client_capabilities.sampling`) |
| **Roots** | `roots/list`, `notifications/roots/list_changed` | Covered (`client_capabilities.roots`) |
| **Elicitation** | `elicitation/create` | Covered (accept, decline, cancel flows) |

### Utilities and protocol mechanics

| Protocol area | Methods | Status |
|--------------|---------|--------|
| **Progress** | `notifications/progress` during tool execution | Covered (`capture_progress` + `min_progress`) |
| **Logging** | `logging/setLevel`, `notifications/message` | Covered (`assert_logging`) |
| **Pagination** | Cursor-based pagination on list endpoints | Covered (`json_path` + `cursor:` field) |
| **Completion** | `completion/complete` (argument autocompletion) | Covered (`assert_completion`) |
| **Cancellation** | `$/cancelRequest` | Not covered |
| **Ping** | `ping` keepalive | Not covered |

### Coverage summary

| Category | Covered | Total | Notes |
|----------|---------|-------|-------|
| Server features | 3/3 | 3 | Tools, Resources (list/read/subscribe), Prompts (list/get) |
| Client features | 3/3 | 3 | Sampling, roots, elicitation |
| Utilities | 4/6 | 6 | Progress, logging, pagination, completion covered; cancellation and ping remain |
| **Total** | **10/12** | **12** | Cancellation and ping are the only gaps |

## Scope Map

### Axis 1: Server capabilities

| Capability | Status | Notes |
|-----------|--------|-------|
| **Stdio servers** | Supported | Launch as subprocess, pipe stdin/stdout |
| **HTTP servers (streamable)** | Supported | `transport: http` with `url:` field |
| **SSE servers (legacy)** | Supported | `transport: sse` with `url:` field |
| **Bidirectional (sampling, roots, elicitation)** | Supported | `client_capabilities` in server YAML config |
| **Authenticated servers (OAuth, API keys)** | Partial | Token injection via `server.env:` with `${VAR}` expansion. OAuth refresh cycles not yet supported. |
| **Streaming/long-running tools** | Partial | Progress notifications captured. Full streaming not yet supported. |
| **Multi-server composition** | Not yet | Testing the composition layer requires intercepting outgoing calls. |

### Axis 2: Test patterns

| Pattern | Status | Notes |
|---------|--------|-------|
| **Single tool responses** | Supported | 15 assertion types + 4 trajectory types |
| **Multi-step workflows** | Supported | `setup:` steps with `capture:` for chaining outputs |
| **Trajectory validation** | Supported | Inline `trace:` and `audit_log:` sources |
| **Live agent trajectory capture** | Planned | Intercept tool calls in real time |
| **Snapshot regression** | Supported | `snapshot --update` captures outputs; subsequent runs detect changes |
| **Cross-language conformance** | Supported | Matrix mode runs same assertions across N servers |

## Bigger Bets

| Item | Priority | Description |
|------|----------|-------------|
| **MCP server leaderboard** | Medium | Public page showing coverage scores for popular MCP servers. Servers compete on correctness. |
| **VS Code extension** | Low | Run assertions from the editor. Click-to-run on YAML files, inline pass/fail markers, coverage gutter annotations. |
