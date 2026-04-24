# Roadmap

## Ecosystem Credibility

| Item | Status | Description |
|------|--------|-------------|
| **File upstream bugs** | **Shipped** | 2 bugs filed: [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029) (filesystem `read_media_file` returns invalid `blob` type) and [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826) (everything server `longRunningOperation` crashes stdio transport). |
| **Community server suites** | **Shipped** | 29 assertions across 4 community servers: 3 mark3labs/mcp-go SDK examples (everything, typed_tools, structured_input_and_output) + PrefectHQ/fastmcp testing_demo (11 assertions, 100% tool coverage). Go and Python framework coverage. Scan-and-contribute flywheel validated. |
| **External adoption** | Planned | Get one MCP server author to use mcp-assert and report results. The mcp-go bug report is the first touchpoint: watch for maintainer response. |
| **Underserved language suites** | Planned | Server suites for Rust, Java, and C# MCP servers — communities with few or no dedicated testing tools. Rust: `rmcp` SDK. Java: MCP SDK for Spring, Quarkus MCP. C#: modelcontextprotocol/csharp-sdk. These communities are underserved by existing MCP tooling and are high-signal distribution targets: post results in their package/Discord communities. |
| **"Works with mcp-assert" badge** | Planned | A standard badge + registry that MCP server authors can claim after passing a reference suite. Server runs `mcp-assert ci --suite <reference-suite>` in their CI; on pass, they add the badge to their README. Drives discoverability: every badge is a backlink. Reference suites would cover the protocol surface (tools, resources, prompts, pagination, error handling) without being server-specific. |
| **Reference suite registry** | Planned | A canonical set of suites that any conforming server can run against, independent of server-specific fixtures. Addresses the duplicate suite problem (agent-lsp and mcp-assert each maintaining their own copy of the same assertions). Authors link to the registry suite; the registry is the single source of truth. |

## Distribution

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **GitHub Action** | **Shipped** | **Highest** | [`blackwell-systems/mcp-assert-action@v1`](https://github.com/blackwell-systems/mcp-assert-action): one line in any workflow. Downloads binary, runs assertions, uploads JUnit XML + badge. |
| **GoReleaser** | **Shipped** | High | v0.1.1 released. Cross-compiled binaries for linux/darwin/windows × amd64/arm64. `go install ...@v0.1.1`. |
| **Homebrew formula** | Planned | High | `brew install mcp-assert` |
| **PyPI wrapper** | Planned | High | `pip install mcp-assert`: downloads the Go binary. Python MCP server authors won't `go install`. |
| **npm wrapper** | Planned | Medium | `npx mcp-assert`: same pattern, TypeScript audience. |
| **MCP registry integration** | Planned | Medium | Surface the mcp-assert test badge prominently on Glama and Smithery listings. Servers that pass a reference suite get a "verified" marker in the registry. Makes correctness a first-class signal in server discovery, not an afterthought. |

The GitHub Action is the single highest-leverage distribution move. If adding mcp-assert to a CI pipeline is one `uses:` line, adoption is frictionless. Every MCP server repo can add it in 30 seconds.

## Technical Depth

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **Automatic fixture isolation** | Planned | **Critical** | Per-test fixture copy in the runner so write-tests can never contaminate read-tests. Today any assertion that commits edits to disk (apply_edit, commit_session with apply: true) modifies the shared fixture directory, shifting line numbers for all subsequent position-sensitive tests. Two hours of debugging in agent-lsp dogfooding exposed this. Implementation: copy the fixture directory to a temp path for each assertion run; restore after. Docker already does this per container. The stdio path needs the same guarantee. This is the single highest-leverage DX improvement: it would eliminate an entire class of hard-to-diagnose failures for every mcp-assert user who writes edit assertions. |
| **`--fix` mode for position errors** | Planned | High | When a position-sensitive assertion fails with "no identifier found" or "column is beyond end of line", suggest the correct line/column. Strategy: re-run the tool call scanning nearby positions (±3 lines, ±5 columns), report the nearest position that succeeds, and emit a suggested YAML patch. Turns a cryptic LSP error into an actionable one-line fix. |
| **Watch mode diff view** | Planned | Medium | When an assertion flips from PASS to FAIL in `--watch` mode, show a diff of the expected vs actual response rather than just the error message. Makes the assertion development loop much tighter — you see exactly what changed without re-reading the YAML. |
| **`init` one-step suite generation** | Planned | Medium | `mcp-assert init --server <cmd>` runs `generate` (stub YAMLs from tools/list) + `snapshot --update` (capture real outputs) in one command. Result: a complete working suite with 100% tool coverage, zero manual assertion writing. Currently requires two separate commands and manual wiring. |
| **HTTP/SSE transport** | **Shipped** | **High** | Test MCP servers over HTTP (streamable HTTP) and SSE (legacy), not just stdio. Set `transport: sse` or `transport: http` with a `url` field in assertion YAML. Uses mcp-go's `NewSSEMCPClient` and `NewStreamableHttpClient`. Docker isolation remains stdio-only. |
| **Snapshot testing** | **Shipped** | High | `mcp-assert snapshot --update` captures tool responses as `.snapshots.json`. Subsequent runs compare against saved snapshots. Like `jest --updateSnapshot`. |
| **--watch mode** | **Shipped** | Medium | `mcp-assert watch` reruns assertions on YAML file change. Polls every 2s, clears terminal between runs. |
| **pass@k in reports** | **Shipped** | Medium | Reliability metrics in JUnit XML (`<properties>`) and markdown (reliability table) when `--trials > 1`. |
| **--coverage-json** | **Shipped** | Medium | `--coverage-json <path>` on `coverage` command writes machine-readable coverage data. |
| **Setup output capture** | **Shipped** | **High** | `capture:` field on setup steps extracts values via jsonpath, injects as `{{variable}}` into subsequent steps. Session lifecycle tests now use real session IDs. |
| **Client capabilities (bidirectional)** | **Shipped** | **High** | Mock client-side capabilities so servers that use sampling, roots, or elicitation can be tested. Set `client_capabilities` in server YAML: `roots: [paths]`, `sampling: {text, model, stop_reason}`, or `elicitation: {content: {...}}`. Verified against mcp-go `roots_server`, `sampling_server`, and `elicitation` example servers. No other MCP testing tool supports this. |
| **Trajectory assertions** | **Shipped** | **Critical** | 4 types (order, presence, absence, args_contain). Inline trace or audit log source. 20 example assertions covering all agent-lsp skill protocols (rename, safe-edit, refactor, cross-repo, dead-code, docs, edit-export, edit-symbol, explore, extract-function, fix-all, format-code, generate, impact, implement, local-symbols, simulate, test-correlation, understand, verify). 21 new tests. Runs in 0ms (no server). |

### Trajectory assertions detail

Trajectory assertions validate that agents call MCP tools in the correct order. They test skill protocols rather than individual tools: "the agent called `prepare_rename` before `rename_symbol`" rather than "`rename_symbol` returned the right output."

The 20 example assertions in `examples/trajectory/` cover all agent-lsp skill protocols:

| Skill | Required sequence | What breaks if skipped |
|-------|------------------|----------------------|
| `/lsp-rename` | `prepare_rename` before `rename_symbol`, then `get_diagnostics` | Renaming without `prepare_rename` skips validation (cursor on keyword, built-in type) |
| `/lsp-refactor` | blast-radius check before any edit | Editing without blast radius risks breaking callers silently |
| `/lsp-safe-edit` | `simulate_edit_atomic` before `apply_edit` | Applying without simulation skips error detection |
| `/lsp-simulate` | no `apply_edit` present | Simulate-only mode must never write to disk |

YAML format (inline trace):

```yaml
name: lsp-rename follows skill protocol
trace:
  - tool: prepare_rename
    args: { file_path: "main.go", line: 6, column: 6 }
  - tool: rename_symbol
    args: { file_path: "main.go", new_name: "Entity" }
  - tool: get_diagnostics
    args: { file_path: "main.go" }
trajectory:
  - type: order
    tools: ["prepare_rename", "rename_symbol", "get_diagnostics"]
  - type: presence
    tools: ["prepare_rename", "rename_symbol", "get_diagnostics"]
  - type: absence
    tools: ["apply_edit"]
  - type: args_contain
    tool: rename_symbol
    args:
      new_name: "Entity"
```

Use `audit_log: path/to/agent.jsonl` instead of `trace:` to validate real agent behavior from a recorded JSONL file. Live agent trajectory capture (real-time interception) is a planned item in the Scope Map.

### Client capabilities detail

MCP is bidirectional: servers can make requests back to the client (sampling, roots, elicitation). mcp-assert supports all three via `client_capabilities` in the server YAML block. This makes it the only MCP testing tool that can fully simulate a bidirectional MCP client environment.

Shipped YAML format:

```yaml
server:
  command: sampling-server
  client_capabilities:
    roots:
      - "{{fixture}}"                 # respond to roots/list with these paths
    sampling:
      text: "mock LLM response"       # respond to sampling/createMessage
      model: mock
      stop_reason: end_turn
    elicitation:
      content:                        # respond to elicitation/create
        projectName: "myapp"
        framework: "react"
        includeTests: true
```

All three are verified against real mcp-go SDK example servers: `roots_server`, `sampling_server`, and `elicitation`. See [Writing Assertions](writing-assertions.md#client-capabilities-bidirectional-mcp) for full examples.

### Setup output capture detail

Setup steps can now capture values from responses via jsonpath and inject them into subsequent steps using `{{variable}}` syntax:

```yaml
setup:
  - tool: create_simulation_session
    args:
      workspace_root: "{{fixture}}"
      language: go
    capture:
      session_id: "$.session_id"    # jsonpath into response

  - tool: simulate_edit
    args:
      session_id: "{{session_id}}"  # use captured value
      file_path: "{{fixture}}/main.go"
      start_line: 13
      end_line: 13
      new_text: "return 42"

assert:
  tool: evaluate_session
  args:
    session_id: "{{session_id}}"    # same captured value
  expect:
    not_error: true
    contains: ["net_delta"]
```

**What this unlocked:**
- Full session lifecycle: create -> edit -> evaluate -> commit/discard -> destroy
- Any multi-step workflow where step N depends on step N-1's output
- Database tests: INSERT returns an ID, SELECT uses that ID
- Auth flows: login returns a token, subsequent calls use it

9 multi-step workflow assertions now use capture for real session lifecycle testing. See [Writing Assertions](writing-assertions.md#chaining-outputs-between-steps-capture) for the full syntax.

### Snapshot testing detail

The biggest friction in writing assertions is knowing what the expected output looks like. Snapshot mode solves this:

```bash
# First run: capture actual outputs as expected values
mcp-assert snapshot --suite evals/ --server "my-server" --update

# Subsequent runs: assert against saved snapshots
mcp-assert snapshot --suite evals/ --server "my-server"
```

This is the `jest --updateSnapshot` pattern applied to MCP servers. Write a minimal YAML with just the tool call and `expect: {}`, run with `--update`, and mcp-assert fills in the expected output. On subsequent runs, it asserts the output hasn't changed.

### Trajectory assertions detail (Scope Map)

Trajectory assertions are shipped. The remaining planned item is live agent trajectory capture: intercepting tool calls made by a real agent session in real time (without needing a pre-recorded trace or audit log). This is distinct from the inline `trace:` and `audit_log:` sources that are already supported.

## MCP Protocol Surface Coverage

Tracking coverage of every method defined in the MCP 2025-11-25 specification.

### Server features

| Protocol area | Methods | Status | Priority |
|--------------|---------|--------|----------|
| **Tools** | `tools/list`, `tools/call`, `notifications/tools/list_changed` | Covered | — |
| **Resources** | `resources/list`, `resources/read` | Covered | — |
| **Resources (advanced)** | `resources/subscribe`, `resources/unsubscribe`, `notifications/resources/updated`, `notifications/resources/list_changed` | Not covered | Low |
| **Prompts** | `prompts/list`, `prompts/get`, `notifications/prompts/list_changed` | **Covered** | — |

### Client features (server-initiated requests)

| Protocol area | Methods | Status | Priority |
|--------------|---------|--------|----------|
| **Sampling** | `sampling/createMessage` | Covered (`client_capabilities.sampling`) | — |
| **Roots** | `roots/list`, `notifications/roots/list_changed` | Covered (`client_capabilities.roots`) | — |
| **Elicitation** | `elicitation/create` | Covered (`client_capabilities.elicitation`) | — |

### Utilities and protocol mechanics

| Protocol area | Methods | Status | Priority |
|--------------|---------|--------|----------|
| **Progress** | `notifications/progress` during tool execution | **Covered** (`capture_progress` + `min_progress`) | — |
| **Cancellation** | `$/cancelRequest` | Not covered | Low |
| **Logging** | `logging/setLevel`, `notifications/message` | Not covered | Low |
| **Pagination** | Cursor-based pagination on `resources/list`, `tools/list`, `prompts/list` | **Covered** (via `json_path` on marshaled response; `cursor:` field in `list:` block) | — |
| **Completion** | `completion/complete` (argument autocompletion) | Not covered | Low |
| **Ping** | `ping` keepalive | Not covered | Low |
| **Task execution** | `taskSupport` on tools, async task results | Not covered | Medium |

### Coverage summary

| Category | Covered | Total | Notes |
|----------|---------|-------|-------|
| Server features | 3/3 | 3 | Tools, Resources (list/read), and Prompts (list/get) all covered |
| Client features | 3/3 | 3 | Sampling, roots, elicitation all covered |
| Utilities | 2/7 | 7 | Progress (capture_progress + min_progress) and Pagination (json_path on list responses) covered; cancellation, logging, completion, ping, tasks remain |
| **Total** | **8/13** | **13** | | |

All three MCP server feature categories have basic coverage. The next push should deepen coverage within each category (subscriptions, sampling as a test subject, completion) and expand elicitation beyond the single mcp-go example.

### Next coverage targets

| Gap | Priority | Description |
|-----|----------|-------------|
| **Resource subscriptions** | Medium | `resources/subscribe`, `resources/unsubscribe`, `notifications/resources/updated`, `notifications/resources/list_changed`. These let clients watch for resource changes. Requires client-side notification handling and a server that actually fires resource change events. The mcp-go `everything` server may support this. |
| **Sampling as first-class test subject** | Medium | Today sampling is tested only as a client capability (mock responses for servers that call `sampling/createMessage`). The gap: testing servers that ARE sampling providers, where mcp-assert sends `sampling/createMessage` requests and validates the response quality, latency, or content. This would let mcp-assert test LLM gateway servers. |
| **Elicitation breadth** | Medium | Only one elicitation example exists (mcp-go `create_project`). Need 2-3 more examples covering different form patterns: multi-step elicitation, validation constraints, cancel/reject flows. |
| **Logging** | Low | `logging/setLevel` + `notifications/message`. Capture server log output during assertion execution. Useful for debugging flaky tests and verifying servers emit expected log events. |
| **Completion** | Low | `completion/complete` for argument autocompletion. Low priority because few servers implement it, but it's part of the spec. |

---

## Scope Map

mcp-assert expands along two axes. The current implementation is already broad; the planned items extend it further.

### Axis 1: Server capabilities

What kinds of servers can be fully tested?

| Capability | Status | Notes |
|-----------|--------|-------|
| **Stdio servers** | Supported | Launch as subprocess, pipe stdin/stdout |
| **HTTP servers (streamable)** | Supported | `transport: http` with `url:` field |
| **SSE servers (legacy)** | Supported | `transport: sse` with `url:` field |
| **Bidirectional (sampling, roots, elicitation)** | **Supported** | Servers that make requests back to the client. Set `client_capabilities` in server YAML config. Verified against mcp-go roots_server, sampling_server, and elicitation example servers. |
| **Authenticated servers (OAuth, API keys)** | Partial | Simple token injection works today via `server.env:`. OAuth refresh cycles need client capabilities expansion. |
| **Streaming/long-running tools** | Partial | Servers that stream progress notifications during execution. The `longRunningOperation` bug in mcp-go exposed this gap. Requires client-side notification handling. |
| **Multi-server composition** | Not yet | Tools that call other MCP servers. Testing the composition layer requires intercepting outgoing calls. |

### Axis 2: Test patterns

What can be asserted?

| Pattern | Status | Notes |
|---------|--------|-------|
| **Single tool responses** | Supported | 15 assertion types: contains, json_path, is_error, net_delta, min_progress, etc. |
| **Multi-step workflows** | Supported | `setup:` steps with `capture:` for chaining outputs |
| **Inline trace validation** | Supported | Trajectory assertions with inline `trace:` |
| **Audit log validation** | Supported | Trajectory assertions against JSONL audit logs |
| **Live agent trajectory capture** | Planned | Intercept tool calls made by a real agent session, validate the sequence automatically |
| **Snapshot regression** | Supported | `snapshot --update` captures outputs; subsequent runs detect changes |
| **Cross-language conformance** | Supported | Matrix mode runs same assertions across N servers |

### Bottom line

The YAML format and transport abstraction are both designed to accommodate expansion. The fundamental model (call a tool, assert the response) holds for virtually every MCP pattern that exists today. The two main expansion areas: making mcp-assert a more capable client (bidirectional, auth, streaming), and extending what can be asserted about server behavior (live trajectory capture, audit integration).

## Bigger Bets

| Item | Status | Description |
|------|--------|-------------|
| **MCP server leaderboard** | Planned | Public page showing coverage scores for popular MCP servers. Run mcp-assert against each, publish results. Servers compete on correctness. |
| **Assertion generation** | **Shipped** | `mcp-assert generate --server <cmd> --output <dir>` queries `tools/list`, reads input schemas, creates one stub YAML per tool. Combined with `snapshot --update`, this gets 100% coverage with zero manual assertion writing. |
| **VS Code extension** | Planned | Run assertions from the editor. Click-to-run on YAML files, inline pass/fail markers, coverage gutter annotations. |
