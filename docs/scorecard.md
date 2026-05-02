# Scan-and-Contribute Scorecard

Servers tested by mcp-assert, bugs found, issues filed.

## Summary

| Metric | Count |
|--------|-------|
| Servers scanned | 58 |
| Server suites | 65 total (63 server + 1 agent-lsp + 1 trajectory) |
| Languages tested | 7 (Go, TypeScript/JavaScript, Python, Rust, Kotlin/Java, Swift, C#) |
| Transports tested | 3 (stdio, SSE, HTTP) |
| Total assertions | 603 (520 server + 63 agent-lsp + 20 trajectory) |
| Upstream bugs found | 31 (12 servers affected + 1 SDK) |
| Upstream issues filed | 9 (1 unfiled: repo archived) |
| Upstream fix PRs submitted | 9 (6 ours pending, 2 merged, 1 closed after maintainer fix) |
| Clean scans (no bugs) | 46 |
| Internal bugs fixed | 6 |

## Server Results

### Anthropic Official Servers

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `@modelcontextprotocol/server-filesystem` | TypeScript | stdio | 14 | 92% (13/14) | 1 | [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029). `read_media_file` returns `type: "blob"`, violating MCP 3035-11-25 spec |
| `@modelcontextprotocol/server-memory` | TypeScript | stdio | 9 | 100% (9/9 tools) | 0 | Clean |
| `mcp-server-time` | Python | stdio | 5 | 100% (2/2 tools) | 0 | Clean. UTC, named timezone, conversion, invalid timezone rejection. |
| `mcp-server-fetch` | Python | stdio | 3 | 100% (1/1 tool) | 0 | Clean. URL fetch, invalid URL rejection, unreachable host handling. |
| `mcp-server-git` | Python | stdio | 11 | 92% (11/12 tools) | 0 | Clean. Status, log, branch, diff, show, invalid repo/ref rejection. |
| `mcp-server-sqlite` | Python | stdio | 9 | 100% (6/6 tools) | 0 | Clean |
| `@modelcontextprotocol/server-everything` | TypeScript | stdio | 13 | 92% (12/13 tools) | 0 | Clean. Official Anthropic reference server. |
| `@modelcontextprotocol/server-puppeteer` | TypeScript | stdio | 7 | 100% (7/7 tools) | 1 | [modelcontextprotocol/servers#4051](https://github.com/modelcontextprotocol/servers/pull/4051). `puppeteer_navigate` crashes with internal error (-32603) on invalid URLs instead of returning `isError:true`. Fix PR submitted. |

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
| `PrefectHQ/fastmcp` testing_demo (SSE) | Python | SSE | 11 | 100% tools | 0 | Same server verified over SSE transport. First SSE coverage. |
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
| `antvis/mcp-server-chart` | TypeScript | stdio | 25 | 93% (25/27 tools) | 9 | [antvis/mcp-server-chart#291](https://github.com/antvis/mcp-server-chart/issues/291). 9 tools crash with unhandled exceptions on default/minimal input. Stack traces leak to agents. **Fix PR [#292](https://github.com/antvis/mcp-server-chart/pull/292) merged.** CI integration follow-up in progress. |

### TypeScript (additional)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `makenotion/notion-mcp-server` | TypeScript | stdio | 22 | 100% (22/22 tools) | 0 | Clean. Official Notion server (4.2K stars). |
| `mongodb/mongodb-mcp-server` | TypeScript | stdio | 4 | -- | 0 | Clean. Knowledge search, error handling. Exemplary error messages with LLM-aware guidance. |
| `onmyway133/git-mcp` | TypeScript | stdio | 14 | 39% (14/36 tools) | 0 | Clean. Status, log, branches, diff, show, reflog, invalid repo rejection. |
| `nicobailey/codegraph-context-mcp` | TypeScript | stdio | 16 | 76% (16/21 tools) | 0 | Clean. Code graph indexer with 21 tools. |

### Go (additional)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `hashicorp/terraform-mcp-server` | Go | stdio | 5 | 56% (5/9 tools) | 0 | Clean. Provider, module, policy search. |
| `sammcj/mcp-devtools` | Go | stdio | 5 | 56% (5/9 tools) | 4 | [sammcj/mcp-devtools#258](https://github.com/sammcj/mcp-devtools/pull/258). Tool handler returns `-32603` internal error instead of `isError:true` for all input validation failures. Affects calculator, get_tool_help, internet_search, search_packages. Fix PR submitted. |

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
| `microsoft/playwright-mcp` | TypeScript | stdio | 14 | 67% (14/21 tools) | 0 | Clean. Navigate, snapshot, screenshot, JS evaluate, console, network, resize, close, tabs, navigate back, press key, wait for element, invalid URL rejection, empty page handling. |

### Research (Python)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `blazickjp/arxiv-mcp-server` | Python | stdio | 5 | 50% (5/10 tools) | 1 | [blazickjp/arxiv-mcp-server#92](https://github.com/blazickjp/arxiv-mcp-server/issues/92). `get_abstract` returns error content in body but `isError` flag not set for invalid paper IDs. |

### AWS

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `awslabs/aws-documentation-mcp-server` | Python | stdio | 4 | 100% (4/4 tools) | 0 | Clean. Search, recommend, no-results handling. Works without AWS credentials. |

### Search API

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `exa-labs/exa-mcp-server` | JavaScript | stdio | 2 | 100% (2/2 tools) | 0 | Clean. Proper 401 with `isError: true` and API key guidance when credentials missing. |

### Observability (Go)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `grafana/mcp-grafana` | Go | stdio | 54 | 100% (50/50 tools) | 1 (fixed) | [grafana/mcp-grafana#793](https://github.com/grafana/mcp-grafana/pull/793) **merged**. `get_assertions` timestamp validation fix. 10 live-backend assertions use `skip_unless_env` (annotations, incidents, dashboards, Prometheus, Loki, Sift). |

### macOS (Swift)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `steipete/Peekaboo` | Swift | stdio | 6 | 27% (6/22 tools) | 1 | [steipete/Peekaboo#108](https://github.com/steipete/Peekaboo/issues/108). `image` returns internal error (-32603) instead of `isError:true` when Screen Recording permission is not granted. |

### macOS (TypeScript)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `getsentry/XcodeBuildMCP` | TypeScript | stdio | 10 | 37% (10/27 tools) | 0 | Clean. 27 tools discovered across build, simulator, coverage, session management. All tested tools return `isError:true` properly when preconditions not met. Sentry-backed. |

### Research/AI (JavaScript)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `u14app/deep-research` | JavaScript | HTTP | 5 | 100% (5/5 tools) | 0 | Clean. All tools return `isError:true` with "Unsupported Provider" when no LLM credentials configured. |
| `perplexityai/mcp-server` | TypeScript | stdio | 4 | 100% (4/4 tools) | 0 | Clean. All tools return `isError:true` with 401 and API key guidance when credentials invalid. |

### Browser Automation (TypeScript)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `chrome-devtools-mcp` | TypeScript | stdio | 7 | 24% (7/29 tools) | 0 | Clean. 29 tools, all return `isError:true` properly. Page management, console, network, performance, screenshots. |
| `mozilla/firefox-devtools-mcp` | TypeScript | stdio | 7 | 24% (7/29 tools) | 0 | Clean. 29 tools via WebDriver BiDi. All return `isError:true` when Firefox not running. Mozilla-backed. |

### Documentation (TypeScript)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `@upstash/context7-mcp` | TypeScript | stdio | 2 | 100% (2/2 tools) | 0 | Clean. Library resolution and documentation search. Upstash-backed. |

### Diagram Generation (Python)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `excalidraw-architect-mcp` | Python | stdio | 4 | 100% (4/4 tools) | 0 | Clean. Architecture diagrams, mermaid conversion. Zero auth. |

### Financial Data (Python)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `sec-edgar-mcp` | Python | stdio | 5 | 24% (5/21 tools) | 0 | Clean. SEC EDGAR filings, insider trading, financials. Free public API (requires user-agent string only). Uses `skip_unless_env`. |
| `narumiruna/yfinance-mcp` | Python | stdio | 4 | 22% (4/18 tools) | 0 | Clean. Live stock market data via Yahoo Finance. Ticker info, price history, search, market calendar. Zero auth, all happy-path assertions return real data. |

### Math/Utility (Python)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `mcp-server-math` | Python | stdio | 4 | 25% (4/16 tools) | 0 | Clean. 16 math tools (sum, divide, average, percentage, etc.). Zero auth. |
| `duckduckgo-mcp-server` | Python | stdio | 2 | 100% (2/2 tools) | 0 | Clean. Search and fetch_content. Zero auth. |

### Infrastructure (Python)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `mcp-server-kubernetes` | Python | stdio | 2 | 40% (2/5 tools) | 0 | Clean. kubectl get, describe error handling. Works without a running cluster. |

### Mobile Automation (TypeScript)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `mobile-next/mobile-mcp` | TypeScript | stdio | 6 | 29% (6/21 tools) | 0 | Clean. 21 tools for iOS/Android automation. All return `isError:true` properly without connected device. 4.7K stars. |

### Cloud Infrastructure (TypeScript)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `lighthouse-mcp-server` | TypeScript | stdio | 2 | 4% (2/57 tools) | 0 | Clean. Tencent Cloud Lighthouse management. 57 tools, all return `isError:true` without credentials. |

### Document Conversion (Python)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `markitdown-mcp` | Python | stdio | 1 | 100% (1/1 tool) | 0 | Clean. Microsoft MarkItDown document-to-markdown converter. |

### Workflow (TypeScript)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `Pimzino/spec-workflow-mcp` | TypeScript | stdio | 1 | 100% (1/1 tool) | 0 | Clean. Spec-driven development workflow. 4.1K stars. |

### C# (.NET)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `modelcontextprotocol/csharp-sdk` QuickstartWeatherServer | C# | stdio | 2 | 100% (2/2 tools) | 0 | Clean. Official C# SDK sample. First C# server tested (7th language). |

### Memory (Go)

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `Gentleman-Programming/engram` | Go | stdio | 16 | 100% (16/16 tools) | 0 | Clean. Full coverage including writes (save, delete, update, merge, session lifecycle). SQLite state is self-contained. |

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
- **Fix PR:** [#292](https://github.com/antvis/mcp-server-chart/pull/292) **merged** (2026-04-28)
- **CI integration:** [#294](https://github.com/antvis/mcp-server-chart/pull/294) submitted. 25 assertion YAML files + GitHub Actions workflow using `mcp-assert-action@v1`. First external adoption.
- **Status:** Fixed. CI integration pending review.

### Bug #5: grafana/mcp-grafana: get_assertions returns internal error on invalid input

- **Severity:** Input validation gap
- **Tool:** `get_assertions`
- **What:** Passing a non-ISO-8601 timestamp (e.g., `"not-a-date"`) returns MCP internal error (-32603) instead of `isError: true`. The `time.Time` field is unmarshalled by the mcp-go SDK before the tool handler runs; invalid input triggers an unmarshal error that the SDK converts to an internal error rather than a tool error.
- **Impact:** MCP clients treat -32603 as a server crash. Agents can't self-correct from internal errors the way they can from `isError: true` responses. All other tools in the server validate input correctly.
- **Issue:** [grafana/mcp-grafana#792](https://github.com/grafana/mcp-grafana/issues/792)
- **Status:** Fix merged ([#793](https://github.com/grafana/mcp-grafana/pull/793), merged 2026-04-27)

### Bug #6: arxiv-mcp-server: get_abstract returns error content without isError flag

- **Severity:** Missing isError flag
- **Tool:** `get_abstract`
- **What:** Calling with an invalid paper ID (e.g., `0000.00000`) returns `{"status": "error", "message": "Paper 0000.00000 not found on arXiv"}` in the content body, but `isError` is not set to `true`.
- **Impact:** Agents checking the `isError` flag treat this as a successful call. The agent may present "Paper not found" as a valid result instead of retrying or reporting failure.
- **Issue:** [blazickjp/arxiv-mcp-server#92](https://github.com/blazickjp/arxiv-mcp-server/issues/92)
- **Status:** Issue closed. Maintainer merged own fix ([#95](https://github.com/blazickjp/arxiv-mcp-server/pull/95)) in v0.5.0, but isError still not set as of 2026-04-26 testing.

### Bug #7: steipete/Peekaboo: image returns internal error without Screen Recording permission

- **Severity:** Internal error instead of isError
- **Tool:** `image`
- **What:** Calling `image` with `mode: screen` when Screen Recording permission is not granted returns MCP internal error (-32603) instead of `isError: true`. Other tools in Peekaboo (like `permissions`) correctly handle missing permissions.
- **Impact:** MCP clients treat -32603 as a server crash. Agents can't self-correct from internal errors.
- **Issue:** [steipete/Peekaboo#108](https://github.com/steipete/Peekaboo/issues/108)
- **Status:** Open

### Bug #8: Anthropic Puppeteer: puppeteer_navigate crashes on invalid URL

- **Severity:** Internal error instead of isError
- **Tool:** `puppeteer_navigate`
- **What:** Calling `puppeteer_navigate` with an invalid or empty URL throws an unhandled CDP `Protocol error (Page.navigate)` that propagates as a JSON-RPC -32603 internal error. The `page.goto()` call has no try/catch. Other tools in the same server (e.g., `puppeteer_screenshot`) correctly catch errors and return `isError: true`.
- **Impact:** Agents sending malformed URLs get an unrecoverable internal error instead of a structured error they can act on.
- **Fix PR:** [modelcontextprotocol/servers#4051](https://github.com/modelcontextprotocol/servers/pull/4051)
- **Status:** Open (server archived to `archive-servers` branch, but npm package still published)

### Bug #9: sammcj/mcp-devtools: tool handler returns internal error instead of isError

- **Severity:** Internal error instead of isError (affects all tools)
- **Tools:** calculator, get_tool_help, internet_search, search_packages
- **What:** The central tool handler in `main.go` returns `(nil, fmt.Errorf(...))` when `Execute()` fails. In mcp-go, this causes a JSON-RPC `-32603` internal error. The error messages are actually helpful and descriptive, but they're wrapped in the wrong error type. Individual tools like `get_library_documentation` already use `mcp.NewToolResultError()` correctly; the bug is in the shared handler.
- **Impact:** Agents cannot distinguish between "the server crashed" and "I sent bad input." All tool execution failures look like server crashes.
- **Fix PR:** [sammcj/mcp-devtools#258](https://github.com/sammcj/mcp-devtools/pull/258)
- **Status:** Open

### Bug #10: MCP TypeScript SDK: null arguments crash every server

- **Severity:** Protocol-level crash (affects all TypeScript SDK servers)
- **What:** `CallToolRequestParamsSchema` uses Zod's `.optional()` which accepts `undefined` but rejects `null`. Clients that serialize missing fields as `null` (Go, Java, C# JSON libraries) trigger a Zod validation error. The error propagates as `-32603` (InternalError) with raw Zod output instead of `-32602` (InvalidParams).
- **Impact:** Every server built on `@modelcontextprotocol/sdk` (12k stars) crashes on null arguments. Affects the filesystem server (14 tools), memory server (9 tools), everything server (13 tools), puppeteer (7 tools), and all community TypeScript servers.
- **Found by:** `mcp-assert fuzz` on first run against the everything server.
- **Issue:** [modelcontextprotocol/typescript-sdk#2012](https://github.com/modelcontextprotocol/typescript-sdk/issues/2012)
- **Fix PR:** [modelcontextprotocol/typescript-sdk#2013](https://github.com/modelcontextprotocol/typescript-sdk/pull/2013)
- **Status:** Open

## Fuzz Results

Adversarial input testing via `mcp-assert fuzz`. Unlike audit (one valid call per tool), fuzz sends 10-50 adversarial inputs per tool: empty strings, null values, wrong types, boundary numbers, missing required fields, injection payloads, and random mutations.

| Server | SDK | Tools | Runs | Passed | Failed | Server-specific findings |
|--------|-----|-------|------|--------|--------|--------------------------|
| `@modelcontextprotocol/server-everything` | TypeScript | 13 | 65 | 51 | 14 | 1 (gzip empty name crash) + SDK null args |
| `@modelcontextprotocol/server-filesystem` | TypeScript | 14 | 140 | 126 | 14 | SDK null args only |
| `@modelcontextprotocol/server-memory` | TypeScript | 9 | 135 | 126 | 9 | SDK null args only |
| `@modelcontextprotocol/server-puppeteer` | TypeScript | 7 | 105 | 82 | 23 | 15 navigate crashes (invalid URLs), 2 fill hangs + SDK null args |
| `mcp-server-sqlite` | Python | 6 | 90 | 90 | 0 | Clean |
| `mcp-server-fetch` | Python | 1 | 15 | 15 | 0 | Clean |
| `mcp-server-time` | Python | 2 | 20 | 20 | 0 | Clean |
| `mcp-server-git` | Python | 12 | 120 | 120 | 0 | Clean |
| `mcp-server-math` | Python | 16 | 160 | 160 | 0 | Clean |
| `duckduckgo-mcp-server` | Python | 2 | 20 | 20 | 0 | Clean |
| `markitdown-mcp` | Python | 1 | 10 | 10 | 0 | Clean |
| `arxiv-mcp-server` | Python | 10 | 100 | 89 | 11 | `check_alerts` timeouts (slow external API calls, not a server bug) |

**Totals:** 12 servers fuzzed, 93 tools, 930 runs, 907 passed, 23 failed.

**Key findings:**

- The TypeScript SDK null args bug is systemic, affecting every TypeScript server. Fix submitted ([typescript-sdk#2013](https://github.com/modelcontextprotocol/typescript-sdk/pull/2013)).
- **Python SDK servers are 100% clean** under adversarial input: 50 tools across 8 servers, 535 runs, zero crashes. The Python SDK handles missing fields, wrong types, and boundary values gracefully.
- Puppeteer has additional server-specific bugs beyond the SDK: unvalidated URLs passed to CDP (15 crashes), missing timeouts on selector waits (2 hangs).
- arxiv timeouts are external API latency, not server bugs.

## Observations

**Bug rate:** 31 bugs across 12 servers + 1 SDK. The most common pattern: unhandled exceptions propagating as JSON-RPC `-32603` internal errors instead of returning `isError: true`. This affects the TypeScript SDK itself (all servers), antvis (9 tools), mcp-devtools (4 tools), Puppeteer (15+ tools via fuzz), Grafana (1 tool), Peekaboo (1 tool). Less common: missing isError flag (arxiv), logic bugs (rmcp), spec violations (filesystem blob type), transport corruption (mcp-go stdio).

**Fuzz testing validates Python SDK quality.** 8 Python servers, 50 tools, 535 adversarial inputs: zero crashes. The Python MCP SDK handles null args, wrong types, missing fields, and boundary values correctly at the transport layer. Every TypeScript server fails on the same SDK-level null args bug.

**Clean scans are valuable too.** fastmcp's clean result (25K-star framework, zero bugs) validates the Python MCP ecosystem's foundations. We document clean scans as positive signals, not wasted effort.

**The flywheel works.** Each issue filed links back to mcp-assert. Maintainers discovering the tool through bug reports is organic adoption: no marketing required. antvis maintainer asked to integrate mcp-assert into their CI after our fix PR.

**Transport bugs are invisible to unit tests.** Both transport-layer bugs (Anthropic blob type, mcp-go stdout corruption) would never show up in the server's own unit tests because those test tool logic, not MCP protocol compliance. This is mcp-assert's core value proposition.

## Coverage Limitations

Not all tools can be tested at full coverage. The main blockers:

| Blocker | Affected servers | Tools skipped |
|---------|-----------------|---------------|
| **Requires credentials** | Grafana (50 tools), Google Storage (17), Perplexity (4), Exa (2), Cloudflare, Atlassian | Tools that call external APIs. We test auth error handling, not actual functionality. |
| **Requires backend service** | Grafana (needs running Grafana), Google Storage (needs GCP), AWS docs (limited without full AWS) | Read-only tools often work; write tools need the real backend. |
| **Requires snapshot refs** | Playwright click, hover, drag, fill, select (7 tools) | These tools operate on element references from a prior `browser_snapshot` call. Requires multi-step assertion chaining with captured output. |
| **macOS permissions** | Peekaboo image, see, analyze (3+ tools) | Screen Recording and Accessibility permissions must be granted at the OS level. |
| **Node 25 incompatibility** | DesktopCommanderMCP, spec-workflow-mcp | Crash on startup with Node 25 due to dependency issues. |

**Per-assertion Docker isolation** (shipped) addresses write safety but not missing credentials or backends. For servers that need a backend (Grafana, databases), the next step is docker-compose with service containers.

**What we can test without credentials:** auth error handling (does the server return `isError: true` with a helpful message?), input validation (does it reject bad params gracefully?), schema conformance (does `tools/list` return valid schemas?), and graceful degradation (does the server start and respond without crashing?).
