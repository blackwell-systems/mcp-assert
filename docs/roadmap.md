# Roadmap

## Ecosystem Credibility

| Item | Status | Description |
|------|--------|-------------|
| **File upstream bugs** | **Shipped** | 2 bugs filed: [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029) (filesystem `read_media_file` returns invalid `blob` type) and [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826) (everything server `longRunningOperation` crashes stdio transport). |
| **Community server suites** | **Shipped** | 29 assertions across 4 community servers: 3 mark3labs/mcp-go SDK examples (everything, typed_tools, structured_input_and_output) + PrefectHQ/fastmcp testing_demo (11 assertions, 100% tool coverage). Go and Python framework coverage. Scan-and-contribute flywheel validated. |
| **External adoption** | Planned | Get one MCP server author to use mcp-assert and report results. The mcp-go bug report is the first touchpoint: watch for maintainer response. |

## Distribution

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **GitHub Action** | **Shipped** | **Highest** | [`blackwell-systems/mcp-assert-action@v1`](https://github.com/blackwell-systems/mcp-assert-action): one line in any workflow. Downloads binary, runs assertions, uploads JUnit XML + badge. |
| **GoReleaser** | **Shipped** | High | v0.1.1 released. Cross-compiled binaries for linux/darwin/windows × amd64/arm64. `go install ...@v0.1.1`. |
| **Homebrew formula** | Planned | High | `brew install mcp-assert` |
| **PyPI wrapper** | Planned | High | `pip install mcp-assert`: downloads the Go binary. Python MCP server authors won't `go install`. |
| **npm wrapper** | Planned | Medium | `npx mcp-assert`: same pattern, TypeScript audience. |

The GitHub Action is the single highest-leverage distribution move. If adding mcp-assert to a CI pipeline is one `uses:` line, adoption is frictionless. Every MCP server repo can add it in 30 seconds.

## Technical Depth

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **HTTP/SSE transport** | **Shipped** | **High** | Test MCP servers over HTTP (streamable HTTP) and SSE (legacy), not just stdio. Set `transport: sse` or `transport: http` with a `url` field in assertion YAML. Uses mcp-go's `NewSSEMCPClient` and `NewStreamableHttpClient`. Docker isolation remains stdio-only. |
| **Snapshot testing** | **Shipped** | High | `mcp-assert snapshot --update` captures tool responses as `.snapshots.json`. Subsequent runs compare against saved snapshots. Like `jest --updateSnapshot`. |
| **--watch mode** | **Shipped** | Medium | `mcp-assert watch` reruns assertions on YAML file change. Polls every 2s, clears terminal between runs. |
| **pass@k in reports** | **Shipped** | Medium | Reliability metrics in JUnit XML (`<properties>`) and markdown (reliability table) when `--trials > 1`. |
| **--coverage-json** | **Shipped** | Medium | `--coverage-json <path>` on `coverage` command writes machine-readable coverage data. |
| **Setup output capture** | **Shipped** | **High** | `capture:` field on setup steps extracts values via jsonpath, injects as `{{variable}}` into subsequent steps. Session lifecycle tests now use real session IDs. |
| **Client capabilities (bidirectional)** | Planned | **High** | Mock client-side capabilities so servers that use sampling, roots, or elicitation can be tested. No other MCP testing tool supports this. |
| **Trajectory assertions** | **Shipped** | **Critical** | 4 types (order, presence, absence, args_contain). Inline trace or audit log source. 20 example assertions covering all agent-lsp skill protocols (rename, safe-edit, refactor, cross-repo, dead-code, docs, edit-export, edit-symbol, explore, extract-function, fix-all, format-code, generate, impact, implement, local-symbols, simulate, test-correlation, understand, verify). 21 new tests. Runs in 0ms (no server). |

### Trajectory assertions detail

**Why this is critical, not nice-to-have:**

mcp-assert currently tests tools in isolation: "call X, check the response." But agent-lsp's 20 skills define exact tool call sequences. A skill is correct only if the agent calls the right tools in the right order. Testing each tool individually proves they work, but doesn't prove the agent follows the protocol.

Examples of skill protocols that trajectory assertions would validate:

| Skill | Required sequence | What breaks if skipped |
|-------|------------------|----------------------|
| `/lsp-rename` | prepare_rename -> confirm -> rename_symbol -> get_diagnostics | Renaming without prepare_rename skips validation (cursor on keyword, built-in type) |
| `/lsp-refactor` | get_references -> simulate_edit_atomic -> apply_edit -> get_diagnostics -> run_tests | Editing without blast radius check risks breaking callers silently |
| `/lsp-safe-edit` | simulate_edit_atomic -> (check net_delta) -> apply_edit | Applying without simulation skips error detection |
| `/lsp-impact` | get_references -> call_hierarchy -> type_hierarchy | Incomplete blast radius analysis |

**Proposed YAML format:**

```yaml
name: lsp-rename follows protocol
trajectory:
  - tool: prepare_rename
    before: rename_symbol
  - tool: rename_symbol
    args_contain:
      new_name: "Entity"
    after: prepare_rename
  - tool: get_diagnostics
    after: rename_symbol
graders:
  - type: order
    assert: prepare_rename before rename_symbol before get_diagnostics
  - type: presence
    assert: all_of [prepare_rename, rename_symbol, get_diagnostics]
  - type: absence
    assert: none_of [apply_edit]  # rename_symbol handles writes internally
```

**Implementation approaches:**

1. **Audit trail parsing:** agent-lsp's `--audit-log` writes every tool call as JSONL. mcp-assert reads the log after a skill execution and checks the call sequence. No transport changes needed.

2. **MCP transport interception:** mcp-assert acts as a proxy between the agent and the server, capturing every tool call in transit. More general but harder to implement.

3. **Setup-driven replay:** Use the existing setup/capture mechanism to execute a sequence of tool calls, then assert on the final state. This works today but doesn't verify an agent's autonomous behavior.

Approach #1 (audit trail) is the pragmatic first step. It works specifically with agent-lsp and requires no changes to mcp-assert's transport layer. The YAML format can be designed now and extended to approach #2 later.

### Client capabilities detail

MCP is bidirectional: servers can request things from the client (sampling, roots, elicitation). mcp-assert currently only acts as a tool-calling client. Adding client capability mocks would let it test servers that depend on these features.

**YAML format:**

```yaml
server:
  command: sampling-server
  client_capabilities:
    roots: ["{{fixture}}"]              # respond to roots/list with these paths
    sampling:                           # respond to sampling/createMessage
      response: "mock LLM response"
    elicitation:                        # respond to elicitation/create
      response:
        name: "Alice"
        confirm: true
```

**Implementation phases:**

| Phase | Capability | Effort | What it unblocks |
|-------|-----------|--------|------------------|
| 1 | **Roots** | Low | Return a list of workspace paths. mcp-go client supports `WithRoots()`. Unblocks mcp-go `roots_server` example. |
| 2 | **Elicitation** | Medium | Return preset key-value pairs for server-initiated prompts. Unblocks mcp-go `elicitation` example. |
| 3 | **Sampling** | Medium | Return mock LLM responses with configurable text, model, and stop reason. Unblocks any server that uses MCP sampling for agent behavior. |

**Why this matters:** The mcp-go SDK has 3 example servers (roots_server, sampling_server, elicitation) that are currently untestable by any MCP testing tool. Building this would make mcp-assert the only tool that can fully simulate an MCP client environment: not just "call tools and check responses" but "participate in the full MCP protocol."

**The mock response pattern is key.** The server doesn't care if the LLM response is real. It just needs a response to continue its workflow. Same for roots (just return paths) and elicitation (just return values). The assertion still checks the tool result: client capabilities are setup for the server to function.

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

### Trajectory assertions detail

Current assertions test single tool calls: "call X, check the response." Trajectory assertions test sequences: "the agent called A, then B, then C: verify the ordering and arguments."

```yaml
trajectory:
  - tool: start_lsp
    args_contain: { root_dir: "{{fixture}}" }
  - tool: get_references
    before: rename_symbol
  - tool: rename_symbol
    args_contain: { new_name: "Entity" }
  - tool: get_diagnostics
    after: rename_symbol
```

This bridges mcp-assert (tool correctness) with skill evaluation (workflow correctness). Low priority because it requires capturing tool call traces, which means either wrapping the MCP transport or parsing audit logs.

## Scope Map

mcp-assert expands along two axes. The current implementation is already broad; the planned items extend it further.

### Axis 1: Server capabilities

What kinds of servers can be fully tested?

| Capability | Status | Notes |
|-----------|--------|-------|
| **Stdio servers** | Supported | Launch as subprocess, pipe stdin/stdout |
| **HTTP servers (streamable)** | Supported | `transport: http` with `url:` field |
| **SSE servers (legacy)** | Supported | `transport: sse` with `url:` field |
| **Bidirectional (sampling, roots, elicitation)** | Planned | Servers that make requests back to the client. Requires mock client capabilities. |
| **Authenticated servers (OAuth, API keys)** | Partial | Simple token injection works today via `server.env:`. OAuth refresh cycles need client capabilities expansion. |
| **Streaming/long-running tools** | Partial | Servers that stream progress notifications during execution. The `longRunningOperation` bug in mcp-go exposed this gap. Requires client-side notification handling. |
| **Multi-server composition** | Not yet | Tools that call other MCP servers. Testing the composition layer requires intercepting outgoing calls. |

### Axis 2: Test patterns

What can be asserted?

| Pattern | Status | Notes |
|---------|--------|-------|
| **Single tool responses** | Supported | 14 assertion types: contains, json_path, is_error, net_delta, etc. |
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
