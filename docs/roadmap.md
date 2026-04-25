# Roadmap

## Ecosystem Credibility

| Item | Priority | Description |
|------|----------|-------------|
| **External adoption** | High | Get one MCP server author to use mcp-assert and report results. |
| **C# server suites** | Medium | C# MCP servers remain untested. `modelcontextprotocol/csharp-sdk`. Last major language gap. |
| **Reference suite registry** | Medium | A canonical set of suites any conforming server can run against, independent of server-specific fixtures. Single source of truth. |

## Distribution

| Item | Priority | Description |
|------|----------|-------------|
| **MCP registry integration** | Medium | Surface the mcp-assert test badge on Glama and Smithery listings. Servers that pass a reference suite get a "verified" marker. |
| **Hacker News launch** | Ready | Everything is polished: 303 assertions, 19 servers, 12 bugs, 5 languages, demo GIF, custom badges. |
| **Nix flake** | Low | Nix users are quality-focused and vocal. High signal in a niche community. |

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
| **Authenticated servers (OAuth, API keys)** | Mostly supported | Token/API key injection via `server.env:` (stdio) and `server.headers:` (HTTP/SSE) with `${VAR}` expansion. Interactive browser-based OAuth flows not yet automated. |
| **Streaming/long-running tools** | Partial | Progress notifications captured. Full streaming not yet supported. |
| **Multi-server composition** | Not yet | Testing the composition layer requires intercepting outgoing calls. |

### Axis 2: Test patterns

| Pattern | Status | Notes |
|---------|--------|-------|
| **Single tool responses** | Supported | 18 assertion types + 4 trajectory types |
| **Multi-step workflows** | Supported | `setup:` steps with `capture:` for chaining outputs |
| **Trajectory validation** | Supported | Inline `trace:` and `audit_log:` sources |
| **Live agent trajectory capture** | Supported | `intercept` command proxies stdio and captures tool calls in real time |
| **Snapshot regression** | Supported | `snapshot --update` captures outputs; subsequent runs detect changes |
| **Cross-language conformance** | Supported | Matrix mode runs same assertions across N servers |

## Framework Integrations

| Item | Priority | Description |
|------|----------|-------------|
| **pytest plugin** | High | Thin Python wrapper that calls the Go binary and reports results as pytest test cases. `pip install pytest-mcp-assert`, then `pytest --mcp-suite evals/`. Each YAML assertion becomes a pytest item with pass/fail/skip semantics, fixtures, markers, and `-k` filtering. The Go binary remains the single source of truth for assertion logic. |

## Assertion Engine

| Item | Priority | Description |
|------|----------|-------------|
| **Structured recovery actions** | Medium | When an assertion fails, return machine-readable guidance ("call tool X to fix") not just an error string. Agents consuming mcp-assert output could self-correct. Inspired by Centian's RecoveryAction pattern. |
| **Invariant drift detection** | Medium | Snapshot a file or state before a tool call, compare after. Like `file_unchanged` but for arbitrary state (e.g., "no new diagnostics introduced"). More powerful than post-hoc file comparison. |
| **`generate` HTTP transport support** | Low | The `generate` command currently only works with stdio servers. Add HTTP/SSE support so remote servers can be auto-discovered. |

## Bigger Bets

| Item | Priority | Description |
|------|----------|-------------|
| **MCP server leaderboard** | Medium | Public page showing coverage scores for popular MCP servers. Servers compete on correctness. |
| **VS Code extension** | Low | Run assertions from the editor. Click-to-run on YAML files, inline pass/fail markers, coverage gutter annotations. |
