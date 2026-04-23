# Roadmap

## Ecosystem Credibility

| Item | Status | Description |
|------|--------|-------------|
| **File upstream bug** | **Shipped** | Filed [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029) — `read_media_file` returns `type: "blob"`, violating MCP spec. |
| **Community server suites** | Planned | Add example assertions for 2-3 popular community servers (Brave Search, GitHub, Slack) to demonstrate breadth beyond official Anthropic servers. |
| **External adoption** | Planned | Get one MCP server author to use mcp-assert and report results. A single "I tested my server with mcp-assert" post is worth more than any feature. |

## Distribution

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **GitHub Action** | **Shipped** | **Highest** | [`blackwell-systems/mcp-assert-action@v1`](https://github.com/blackwell-systems/mcp-assert-action) — one line in any workflow. Downloads binary, runs assertions, uploads JUnit XML + badge. |
| **GoReleaser** | **Shipped** | High | v0.1.1 released. Cross-compiled binaries for linux/darwin/windows × amd64/arm64. `go install ...@v0.1.1`. |
| **Homebrew formula** | Planned | High | `brew install mcp-assert` |
| **PyPI wrapper** | Planned | High | `pip install mcp-assert` — downloads the Go binary. Python MCP server authors won't `go install`. |
| **npm wrapper** | Planned | Medium | `npx mcp-assert` — same pattern, TypeScript audience. |

The GitHub Action is the single highest-leverage distribution move. If adding mcp-assert to a CI pipeline is one `uses:` line, adoption is frictionless. Every MCP server repo can add it in 30 seconds.

## Technical Depth

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **Snapshot testing** | **Shipped** | High | `mcp-assert snapshot --update` captures tool responses as `.snapshots.json`. Subsequent runs compare against saved snapshots. Like `jest --updateSnapshot`. |
| **--watch mode** | **Shipped** | Medium | `mcp-assert watch` reruns assertions on YAML file change. Polls every 2s, clears terminal between runs. |
| **pass@k in reports** | **Shipped** | Medium | Reliability metrics in JUnit XML (`<properties>`) and markdown (reliability table) when `--trials > 1`. |
| **--coverage-json** | **Shipped** | Medium | `--coverage-json <path>` on `coverage` command writes machine-readable coverage data. |
| **Setup output capture** | **Shipped** | **High** | `capture:` field on setup steps extracts values via jsonpath, injects as `{{variable}}` into subsequent steps. Session lifecycle tests now use real session IDs. |
| **Trajectory assertions** | Planned | Low | Assert on the sequence of tool calls in a multi-step workflow, not just single tool responses. Requires capturing the full call trace, not just the final result. |

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

Current assertions test single tool calls: "call X, check the response." Trajectory assertions test sequences: "the agent called A, then B, then C — verify the ordering and arguments."

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

## Bigger Bets

| Item | Status | Description |
|------|--------|-------------|
| **MCP server leaderboard** | Planned | Public page showing coverage scores for popular MCP servers. Run mcp-assert against each, publish results. Servers compete on correctness. |
| **Assertion generation** | **Shipped** | `mcp-assert generate --server <cmd> --output <dir>` queries `tools/list`, reads input schemas, creates one stub YAML per tool. Combined with `snapshot --update`, this gets 100% coverage with zero manual assertion writing. |
| **VS Code extension** | Planned | Run assertions from the editor. Click-to-run on YAML files, inline pass/fail markers, coverage gutter annotations. |
