# Scan-and-Contribute Scorecard

Servers tested by mcp-assert, bugs found, issues filed.

## Summary

| Metric | Count |
|--------|-------|
| Servers scanned | 27 |
| Server suites | 25 (including HTTP transport variant, SSE variant, prompts, resources, completion, logging, GitHub MCP, and rmcp suites) |
| Languages tested | 4 (Go, TypeScript, Python, Rust) |
| Transports tested | 3 (stdio, SSE, HTTP) |
| Total assertions | 343 (323 server + 20 trajectory) |
| Upstream bugs found | 14 (5 servers affected) |
| Upstream issues filed | 5 (1 unfiled: repo archived) |
| Clean scans (no bugs) | 20 |
| Internal bugs fixed | 6 |

## Server Results

### Anthropic Official Servers

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `@modelcontextprotocol/server-filesystem` | TypeScript | stdio | 14 | 92% (13/14) | 1 | [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029). `read_media_file` returns `type: "blob"`, violating MCP 3035-11-25 spec |
| `@modelcontextprotocol/server-memory` | TypeScript | stdio | 5 | - | 0 | Clean |
| `mcp-server-time` | Python | stdio | 5 | 100% (2/2 tools) | 0 | Clean. UTC, named timezone, conversion, invalid timezone rejection. |
| `mcp-server-fetch` | Python | stdio | 3 | 100% (1/1 tool) | 0 | Clean. URL fetch, invalid URL rejection, unreachable host handling. |
| `mcp-server-git` | Python | stdio | 7 | 58% (7/12 tools) | 0 | Clean. Status, log, branch, diff, show, invalid repo/ref rejection. |
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

### JVM (Kotlin/Java)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `jamesward/hello-spring-mcp-server` | Kotlin | HTTP | 3 | 100% (2/2 tools) | 0 | Clean. First JVM server tested. Spring Boot + Spring AI MCP. |

### OpenAI

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `openai/sample-deep-research-mcp` | Python | stdio | 4 | 100% (2/2 tools) | 0 | Clean. "Cupcake MCP" sample server. Search and fetch against static JSON dataset. |

### Google Cloud

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `@google-cloud/storage-mcp` | TypeScript | stdio | 6 | 35% (6/17 tools, no GCP credentials) | 0 | Clean. Bucket metadata, object listing, IAM policy, input validation. Graceful error handling without credentials. |

### Microsoft

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `@playwright/mcp` | TypeScript | stdio | 10 | 48% (10/21 tools) | 0 | Clean. Navigate, snapshot, screenshot, JS evaluate, console, network, resize, close, invalid URL rejection, empty page handling. |

### Research (Python)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `blazickjp/arxiv-mcp-server` | Python | stdio | 5 | 50% (5/10 tools) | 1 | [blazickjp/arxiv-mcp-server#92](https://github.com/blazickjp/arxiv-mcp-server/issues/92). `get_abstract` returns error content in body but `isError` flag not set for invalid paper IDs. |

### Observability (Go)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `grafana/mcp-grafana` | Go | stdio | 10 | 20% (10/50 tools, no Grafana backend) | 1 | [grafana/mcp-grafana#792](https://github.com/grafana/mcp-grafana/issues/792). `get_assertions` returns internal error (-32603) instead of `isError:true` on invalid timestamp. |

### Internal (agent-lsp)

| Server | Language | Transport | Assertions | Coverage | Bugs fixed |
|--------|----------|-----------|------------|----------|------------|
| agent-lsp + gopls | Go | stdio | 63 | 100% (50/50 tools) | 6: `character`→`column` param rename, `format_range` 0-indexed docs, undocumented `simulate_edit_atomic` params, missing warmup pattern, shared fixture mutation, SSE/HTTP transport `Start()` not called |
| agent-lsp skill protocols | N/A (inline trace) | N/A | 20 | 20/20 skills | Trajectory assertions: all 20 skills have required tool call sequences, safety gates, and absence checks verified |

## Bug Details

### Bug #1: Anthropic filesystem: invalid MCP content type

- **Severity:** Spec violation
- **Tool:** `read_media_file`
- **What:** Returns `type: "blob"` which is not a valid MCP content type. The spec allows `text`, `image`, `audio`, `resource_link`, `resource`.
- **Impact:** Any MCP client receiving this response crashes at the transport layer.
- **Issue:** [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029)
- **Status:** Fix submitted ([#4044](https://github.com/modelcontextprotocol/servers/pull/4044)), pending merge

### Bug #2: mcp-go SDK: stdio transport crash on slow tools

- **Severity:** Transport crash
- **Tool:** `longRunningOperation` in `examples/everything`
- **What:** Tool calls `time.Sleep()` which creates a timing window for `fmt.Printf` hooks to corrupt the stdio JSON-RPC stream.
- **Impact:** Any mcp-go server with debug hooks and slow tool handlers will crash over stdio.
- **Issue:** [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826)
- **Status:** Fix submitted ([#828](https://github.com/mark3labs/mcp-go/pull/828)), pending merge

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

### Bug #5: grafana/mcp-grafana: get_assertions returns internal error on invalid input

- **Severity:** Input validation gap
- **Tool:** `get_assertions`
- **What:** Passing a non-ISO-8601 timestamp (e.g., `"not-a-date"`) returns MCP internal error (-32603) instead of `isError: true`. The `time.Time` field is unmarshalled by the mcp-go SDK before the tool handler runs; invalid input triggers an unmarshal error that the SDK converts to an internal error rather than a tool error.
- **Impact:** MCP clients treat -32603 as a server crash. Agents can't self-correct from internal errors the way they can from `isError: true` responses. All other tools in the server validate input correctly.
- **Issue:** [grafana/mcp-grafana#792](https://github.com/grafana/mcp-grafana/issues/792)
- **Status:** Fix submitted ([#793](https://github.com/grafana/mcp-grafana/pull/793)), pending merge

### Bug #6: arxiv-mcp-server: get_abstract returns error content without isError flag

- **Severity:** Missing isError flag
- **Tool:** `get_abstract`
- **What:** Calling with an invalid paper ID (e.g., `0000.00000`) returns `{"status": "error", "message": "Paper 0000.00000 not found on arXiv"}` in the content body, but `isError` is not set to `true`.
- **Impact:** Agents checking the `isError` flag treat this as a successful call. The agent may present "Paper not found" as a valid result instead of retrying or reporting failure.
- **Issue:** [blazickjp/arxiv-mcp-server#92](https://github.com/blazickjp/arxiv-mcp-server/issues/92)
- **Status:** Fix submitted ([#93](https://github.com/blazickjp/arxiv-mcp-server/pull/93)), pending merge

## Observations

**Bug rate:** 14 bugs across 5 of 27 servers scanned. Two transport/protocol-level, one logic bug in example code, nine unhandled exception crashes in a charting server, one input validation gap in Grafana's MCP server, one missing isError flag in the arxiv server.

**Clean scans are valuable too.** fastmcp's clean result (25K-star framework, zero bugs) validates the Python MCP ecosystem's foundations. We document clean scans as positive signals, not wasted effort.

**The flywheel works.** Each issue filed links back to mcp-assert. Maintainers discovering the tool through bug reports is organic adoption: no marketing required.

**Transport bugs are invisible to unit tests.** Both bugs were in the transport layer: they'd never show up in the server's own unit tests because those test tool logic, not MCP protocol compliance. This is mcp-assert's core value proposition.
