# Roadmap

## Next Up

| Item | Priority | Description |
|------|----------|-------------|
| **MCP server leaderboard** | High | Static page on docs site ranking servers by coverage score, pass rate, and lint score. Data exists for 58 servers. |
| **Reference suite registry** | Medium | Canonical protocol conformance assertions any MCP server can run. Independent of server-specific fixtures. "Does this server speak MCP correctly?" |
| **Download stats dashboard** | Medium | Script or tool that queries PyPI, npm, GitHub releases, and Homebrew APIs to produce a unified download report. |
| **Docker images** | Low | Per-runtime images (node, python, go) for running `mcp-assert audit/ci` without installing the binary. |
| **Nix flake** | Low | Nix users are quality-focused and vocal. |
| **Lint: tool ambiguity detection** | Low | Compare all tool names/descriptions pairwise, flag pairs an LLM would confuse. |

## `mcp-assert ui` (deferred, feature-gated)

Web UI for mcp-assert: explore servers, debug assertion failures, trace agent sessions. **Deferred until demand is proven.** The CLI is the primary interface; MCP server developers live in terminals.

### Status: design complete, build deferred

The design work is done (four modes, frontend stack decision, ProtoMCP analysis). Build is deferred pending adoption signal. The hosted audit experiment (below) tests demand with minimal investment first.

### Feature gating via Go build tags

The UI is optional. Default binary has no frontend overhead:

```go
//go:build ui

package ui

//go:embed dist/*
var Files embed.FS
```

```bash
# Default: CLI only, small binary
go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest

# With UI: includes embedded frontend (~5-15MB larger)
go install -tags ui github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest
```

Without `-tags ui`, the `mcp-assert ui` command prints "built without UI support" and exits. The distribution pipeline (npm, pip, brew, scoop, Docker) ships the default CLI-only binary. A separate `mcp-assert-ui` binary or install flag provides the UI variant. Same pattern as Prometheus, Caddy, and other Go tools with optional features.

### Four modes (if built)

**Phase 1**: Explorer (browse tools, call interactively, save as assertion) + Debugger (visual assertion failure inspector with diff view).

**Phase 2**: Agent (connect LLM, watch ReAct loop, record trajectory) + Tracer (proxy between external agent and server, live timeline).

### Frontend stack (if built)

Preact + Tailwind CSS + esbuild. Embedded via `embed.FS`. Inspired by ProtoMCP's layout (three-column: server list, main content, JSON-RPC log panel). Our differentiation: "Save as assertion" button, expected vs actual diff, CI export, testing layer ProtoMCP lacks.

## Hosted Audit Experiment

Test demand before building the full UI. Minimal investment, maximum signal.

**What**: a landing page at mcp-assert.dev where users paste a server command and get an audit report. No CLI install needed.

**How**: backend runs `mcp-assert audit --json` in a container, renders results as a static HTML report. Server-side rendering, no SPA, no Preact. A form and a report page.

**Why before the UI**: proves whether anyone wants browser-based MCP testing. If yes, invest in the richer UI. If no, saved weeks of frontend work. The CLI and CI workflow remain the core product regardless.

| Tier | What | Pricing |
|------|------|---------|
| **Free (OSS)** | CLI, GitHub Action, all assertion types | Free forever |
| **Hosted audit** | Paste a server URL, get results in the browser. No CLI install. | Free: 5 audits/month. Paid: unlimited. |
| **Quality registry** | Public leaderboard. Server authors claim listings, add verified badge, show CI status. | Free listing. Verified badge: paid. |
| **Continuous monitoring** | Run assertion suite on schedule against live servers. Alert on regression (Slack, email, PagerDuty). | $29/mo per server, $99/mo teams |
| **Team dashboard** | Shared view of org's MCP servers, coverage, pass rates, trends. Role-based access, audit logs. | Enterprise pricing |

### Monetization sequence

```
OSS CLI (free) → hosted audit (freemium, test demand) → UI (if demand exists) → registry (paid) → monitoring (SaaS)
```

Viability depends on MCP ecosystem growth. If MCP becomes the standard agent-to-tool protocol (Anthropic, OpenAI, Google all pushing it), the quality layer is infrastructure.

## Open PRs and Issues

| PR/Issue | Repo | Status | Description |
|----------|------|--------|-------------|
| mark3labs/mcp-go#839 | Fix: listenForever infinite retry on 404 | Open | Session terminated not detected |
| mark3labs/mcp-go#828 | Fix: stderr hooks | Open | stdio transport corruption |
| modelcontextprotocol/servers#4095 | Issue: filesystem schema quality | Open | 16 required params undescribed |
| modelcontextprotocol/servers#4044 | Fix: blob content type | Open | MCP spec violation |
| modelcontextprotocol/servers#4051 | Fix: puppeteer_navigate isError | Open | Unhandled CDP error |
| github/github-mcp-server#2425 | Issue: schema quality (112 findings) | Open | 20 errors, 92 warnings |
| makenotion/notion-mcp-server#280 | Issue: schema quality (36 findings) | Open | 8 required params undescribed |
| modelcontextprotocol/typescript-sdk#2013 | Fix: null args crash | Open | Affects all TS SDK servers |
| modelcontextprotocol/go-sdk#929 | Fix: HTTP response body leak | Open | Leaked on every session close |
| modelcontextprotocol/python-sdk#2536 | Fix: lost-wakeup race | Open | Concurrent pollers hang forever |
| modelcontextprotocol/php-sdk#297 | Fix: URI scheme validation | Open | RFC 3986 compliance |
| sammcj/mcp-devtools#258 | Fix: isError instead of internal error | Open | 4 tools affected |

### Merged

| PR/Issue | Repo | Date |
|----------|------|------|
| mark3labs/mcp-go#838 | Fix: isError for input validation in example | 2026-05-03 |
| antvis/mcp-server-chart#292 | Fix: isError on chart failures | 2026-04-28 |
| grafana/mcp-grafana#793 | Fix: timestamp validation | 2026-04-27 |

## Coverage Expansion Opportunities

| Server | Current | Potential | Blocker |
|--------|---------|-----------|---------|
| Playwright | 67% (14/21) | ~85% | click/hover/drag need snapshot element refs (multi-step chaining) |
| Google Storage | 35% (6/17) | ~80% | Needs GCP credentials (use skip_unless_env) |
| Grafana | 100% (50/50) | 100% | Complete. 10 live-backend assertions use `skip_unless_env`. |
| git-mcp (idosal) | 39% (14/36) | ~60% | Many write tools need valid repo state |
| Perplexity | 100% auth errors only | 100% real | Needs API key ($5 free credits) |

## MCP Protocol Coverage

10 of 12 MCP protocol methods covered. Two gaps remain (low priority, rarely used):

| Protocol area | Status |
|--------------|--------|
| **Cancellation** (`$/cancelRequest`) | Not covered |
| **Ping** keepalive | Not covered |

## Assertion Engine

| Item | Priority | Description |
|------|----------|-------------|
| **Structured recovery actions** | Medium | When an assertion fails, return machine-readable guidance. Agents consuming mcp-assert output could self-correct. |
| **Invariant drift detection** | Medium | Snapshot state before a tool call, compare after. |

## Recently Shipped

| Item | Version | Description |
|------|---------|-------------|
| **`mcp-assert lint` command** | 0.9.0 | Static schema analysis for agent usability. 7 lint codes (E101-E103, E301, W101-W103). Checks missing descriptions, types, constraints. Found 254 issues across 11 servers including official filesystem (16 errors) and GitHub MCP (112 issues). |
| **`mcp-assert fuzz` command** | 0.8.0 | Zero-setup adversarial testing. Category-based input generation from JSON Schema. Found bugs in 5 official MCP SDKs. |
| **Shared server flags + connectAndInitialize** | 0.9.0 | Internal refactoring: shared flags prevent drift, shared connection logic eliminates duplication across audit/fuzz/lint/generate. |
| **`mcp-assert audit` command** | 0.6.0 | Zero-config quality audit. Connects, discovers tools, calls each with schema-generated inputs, reports quality score. |
| **Blog post** | -- | "We Tested 55 MCP Servers. Here's What Breaks." Published, needs update to current numbers. |
| **antvis CI integration** | -- | PR #292 merged. Follow-up #294 submitted (closed by maintainer). First external adoption. |
| **C# server suites** | 0.6.0 | 7th language. QuickstartWeatherServer. |
| **`skip_unless_env` field** | 0.6.0 | Conditional assertion skipping based on env vars. Live-backend and no-auth assertions coexist in same suite. |
| **Per-assertion Docker isolation** | 0.6.0 | `docker:` field in server YAML. Fresh container per assertion for safe write testing. |
| **Coverage expansion** | 0.6.0 | SQLite 100%, Memory 100%, engram 100%. Anthropic git 92%, Playwright 67%. |
| **11 new server suites** | 0.6.0 | 55 servers, 570 assertions, 7 languages, 20 bugs. XcodeBuildMCP, Puppeteer, Context7, Chrome DevTools, Firefox DevTools, Excalidraw, SEC EDGAR, mcp-devtools, mcp-math, DuckDuckGo, Kubernetes, plus Perplexity, Peekaboo, CodeGraphContext, deep-research from pre-release. |
| **Vitest plugin** | 0.7.0 | `npm install -D vitest-mcp-assert`. Auto-discover YAML files or per-test control. Same bridge architecture as pytest plugin. |
| **pytest plugin** | 0.5.0 | `pip install pytest-mcp-assert`. Published to PyPI via release pipeline. |
| **Badge snippet on pass** | 0.5.0 | CLI and GitHub Action output ready-to-paste badge markdown. |
| **SSE transport fix** | 0.4.0 | `Start()` missing for SSE/HTTP clients. Found by dogfooding. |
