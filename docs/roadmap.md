# Roadmap

## Ecosystem Credibility

| Item | Status | Description |
|------|--------|-------------|
| **File upstream bug** | Planned | File issue on `@modelcontextprotocol/server-filesystem` for `read_media_file` returning invalid MCP response on non-media files. "Found by mcp-assert" is social proof. |
| **Community server suites** | Planned | Add example assertions for 2-3 popular community servers (Brave Search, GitHub, Slack) to demonstrate breadth beyond official Anthropic servers. |
| **External adoption** | Planned | Get one MCP server author to use mcp-assert and report results. A single "I tested my server with mcp-assert" post is worth more than any feature. |

## Distribution

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **GitHub Action** | Planned | **Highest** | `uses: blackwell-systems/mcp-assert-action@v1` — one line in any workflow. Downloads the binary, runs assertions, uploads JUnit XML. Frictionless adoption for every MCP server repo. |
| **Homebrew formula** | Planned | High | `brew install mcp-assert` |
| **PyPI wrapper** | Planned | High | `pip install mcp-assert` — downloads the Go binary. Python MCP server authors won't `go install`. |
| **npm wrapper** | Planned | Medium | `npx mcp-assert` — same pattern, TypeScript audience. |
| **GoReleaser** | Planned | High | Tagged releases with cross-compiled binaries for `go install ...@v0.1.0` and GitHub Releases download. |

The GitHub Action is the single highest-leverage distribution move. If adding mcp-assert to a CI pipeline is one `uses:` line, adoption is frictionless. Every MCP server repo can add it in 30 seconds.

## Technical Depth

| Item | Status | Priority | Description |
|------|--------|----------|-------------|
| **--watch mode** | Planned | Medium | Rerun assertions on file change. Assertion development loop: edit YAML, save, see result. |
| **Snapshot testing** | Planned | High | `mcp-assert snapshot --update` auto-generates expected values from first run, saves as baseline, asserts against them on subsequent runs. Eliminates manual assertion writing for initial coverage. |
| **pass@k in reports** | Planned | Medium | Include reliability metrics in JUnit XML and markdown output, not just terminal. |
| **--coverage-json** | Planned | Medium | Machine-readable coverage data for dashboards and badges. |
| **Trajectory assertions** | Planned | Low | Assert on the sequence of tool calls in a multi-step workflow, not just single tool responses. Requires capturing the full call trace, not just the final result. |

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
| **Assertion generation from OpenAPI/tool schema** | Planned | Given a server's `tools/list` output, auto-generate stub assertions for every tool. Combined with snapshot mode, this gets to 100% coverage with zero manual work. |
| **VS Code extension** | Planned | Run assertions from the editor. Click-to-run on YAML files, inline pass/fail markers, coverage gutter annotations. |
