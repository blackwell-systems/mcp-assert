# Scan-and-Contribute Scorecard

Servers tested by mcp-assert, bugs found, issues filed.

## Summary

| Metric | Count |
|--------|-------|
| Servers scanned | 16 |
| Server suites | 20 (including HTTP transport variant, prompts, resources, completion, logging, GitHub MCP, and rmcp suites) |
| Languages tested | 4 (Go, TypeScript, Python, Rust) |
| Transports tested | 3 (stdio, SSE, HTTP) |
| Total assertions | 274 (254 server + 20 trajectory) |
| Upstream bugs found | 12 (3 servers affected) |
| Upstream issues filed | 3 (1 unfiled: repo archived) |
| Clean scans (no bugs) | 11 |
| Internal bugs fixed | 5 |

## Server Results

### Anthropic Official Servers

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `@modelcontextprotocol/server-filesystem` | TypeScript | stdio | 14 | 92% (13/14) | 1 | [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029). `read_media_file` returns `type: "blob"`, violating MCP 2745-11-25 spec |
| `@modelcontextprotocol/server-memory` | TypeScript | stdio | 5 | - | 0 | Clean |
| `mcp-server-sqlite` | Python | stdio | 6 | - | 0 | Clean |

### Community Framework SDKs

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `mark3labs/mcp-go` everything | Go | stdio | 9 | 100% | 1 | [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826). `longRunningOperation` crashes stdio transport (fmt.Printf to stdout corrupts JSON-RPC) |
| `mark3labs/mcp-go` everything | Go | HTTP | 5 | 100% | 0 | Transport conformance: same tools pass over HTTP |
| `mark3labs/mcp-go` everything (prompts) | Go | stdio | 4 | 100% | 0 | Clean. `prompts/list` (2 prompts), `prompts/get` for static and template prompts, pagination pattern documented. |
| `mark3labs/mcp-go` everything (resources) | Go | stdio | 4 | 100% | 0 | Clean. `resources/list`, `resources/read`, `resources/subscribe`, `resources/unsubscribe`. |
| `mark3labs/mcp-go` typed_tools | Go | stdio | 3 | 100% | 0 | Clean |
| `mark3labs/mcp-go` structured | Go | stdio | 6 | 100% | 0 | Clean |
| `mark3labs/mcp-go` roots_server | Go | stdio | 1 | 100% | 0 | Clean. Verified bidirectional roots/list via `client_capabilities.roots` |
| `mark3labs/mcp-go` sampling_server | Go | stdio | 3 | 100% | 0 | Clean. Verified bidirectional sampling/createMessage via `client_capabilities.sampling` |
| `mark3labs/mcp-go` elicitation | Go | stdio | 4 | 100% | 0 | Clean. `create_project`, `cancel_flow`, `decline_flow`, `validation_constraints` verified via `client_capabilities.elicitation`. |
| `mark3labs/mcp-go` everything (completion) | Go | stdio | 3 | 100% | 0 | Clean. `completion/complete` for prompt argument, resource URI, and empty prefix. |
| `mark3labs/mcp-go` everything (logging) | Go | stdio | 2 | 100% | 0 | Clean. `logging/setLevel` with info level, log message capture after tool call. |
| `PrefectHQ/fastmcp` testing_demo | Python | stdio | 16 | 100% tools + resources + prompts | 0 | Clean. All three MCP feature categories: 11 tool assertions (100% coverage), 3 resource assertions (list, read static, read parameterized), 2 prompt assertions (list, get with argument). |
| `github/github-mcp-server` | Go | stdio | 20 | -- (read-only subset, 17 tools across 7 toolsets) | 0 | Clean. Context, repos, git, issues, pull requests, users, gists toolsets. |

### Rust SDK

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `4t145/rmcp` counter | Rust | stdio | 14 | 100% (6/6 tools + resources + prompts) | 1 | `get_value` decrements counter instead of reading. Repo archived, issue cannot be filed. |
| `rust-mcp-stack/rust-mcp-filesystem` | Rust | stdio | 23 | 92% (22/24 tools) | 0 | Clean. Read, list, search, write, edit, zip/unzip, path traversal rejection. |

### Python (additional)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `haris-musa/excel-mcp-server` | Python | stdio | 15 | 52% (13/25 tools) | 0 | Clean. Workbook, sheets, data round-trip, formulas, charts, pivot tables, formatting, merge, validation. |

### TypeScript (additional)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `antvis/mcp-server-chart` | TypeScript | stdio | 16 | 59% (16/27 tools) | 9 | [antvis/mcp-server-chart#291](https://github.com/antvis/mcp-server-chart/issues/291). 9 tools crash with unhandled exceptions on default/minimal input. Stack traces leak to agents. |

### Internal (agent-lsp)

| Server | Language | Transport | Assertions | Coverage | Bugs fixed |
|--------|----------|-----------|------------|----------|------------|
| agent-lsp + gopls | Go | stdio | 63 | 100% (50/50 tools) | 5: `character`→`column` param rename, `format_range` 0-indexed docs, undocumented `simulate_edit_atomic` params, missing warmup pattern, shared fixture mutation |
| agent-lsp skill protocols | N/A (inline trace) | N/A | 20 | 20/20 skills | Trajectory assertions: all 20 skills have required tool call sequences, safety gates, and absence checks verified |

## Bug Details

### Bug #1: Anthropic filesystem: invalid MCP content type

- **Severity:** Spec violation
- **Tool:** `read_media_file`
- **What:** Returns `type: "blob"` which is not a valid MCP content type. The spec allows `text`, `image`, `audio`, `resource_link`, `resource`.
- **Impact:** Any MCP client receiving this response crashes at the transport layer.
- **Issue:** [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029)
- **Status:** Open

### Bug #2: mcp-go SDK: stdio transport crash on slow tools

- **Severity:** Transport crash
- **Tool:** `longRunningOperation` in `examples/everything`
- **What:** Tool calls `time.Sleep()` which creates a timing window for `fmt.Printf` hooks to corrupt the stdio JSON-RPC stream.
- **Impact:** Any mcp-go server with debug hooks and slow tool handlers will crash over stdio.
- **Issue:** [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826)
- **Status:** Open, linked to project board

### Bug #3: rmcp SDK example: get_value mutates state

- **Severity:** Logic bug in example code
- **Tool:** `get_value` in `examples/servers/src/common/counter.rs`
- **What:** Tool is documented as "Get the current counter value" but actually decrements the counter (`*counter -= 1`). Not idempotent.
- **Impact:** Every developer learning from this example copies a getter that mutates state. An agent calling `get_value` to "check" the counter unknowingly decrements it.
- **Status:** Cannot file issue (repo archived March 2025). Documented in assertion suite. Superseded by `rust-mcp-stack/rust-mcp-sdk`.

### Bug #4: antvis/mcp-server-chart: 9 tools crash with unhandled exceptions

- **Severity:** Unhandled exceptions (9 tools affected)
- **Tools:** `generate_fishbone_diagram`, `generate_mind_map`, `generate_organization_chart`, `generate_flow_diagram`, `generate_network_graph`, `generate_funnel_chart`, `generate_venn_chart`, `generate_district_map`, `generate_radar_chart`
- **What:** Tools throw raw JavaScript exceptions (TypeError, G6 graph errors) on default/minimal input. Exceptions propagate as MCP error -32603 with full stack traces instead of returning `isError: true` with helpful messages.
- **Impact:** LLM agents receive cryptic Node.js stack traces they cannot recover from. The `generate_radar_chart` tool also crashes with populated data when field names don't match undocumented expectations.
- **Issue:** [antvis/mcp-server-chart#291](https://github.com/antvis/mcp-server-chart/issues/291)
- **Status:** Open

## Observations

**Bug rate:** 12 bugs across 3 of 14 servers scanned. Two transport/protocol-level, one logic bug in example code, nine unhandled exception crashes in a charting server.

**Clean scans are valuable too.** fastmcp's clean result (25K-star framework, zero bugs) validates the Python MCP ecosystem's foundations. We document clean scans as positive signals, not wasted effort.

**The flywheel works.** Each issue filed links back to mcp-assert. Maintainers discovering the tool through bug reports is organic adoption: no marketing required.

**Transport bugs are invisible to unit tests.** Both bugs were in the transport layer: they'd never show up in the server's own unit tests because those test tool logic, not MCP protocol compliance. This is mcp-assert's core value proposition.
