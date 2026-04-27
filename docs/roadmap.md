# Roadmap

## Next Up

| Item | Priority | Description |
|------|----------|-------------|
| **Blog post** | Ready | "We tested 50 MCP servers from Anthropic, Google, OpenAI, Microsoft, Mozilla, Sentry, and AWS. Here's what we found." The scorecard data is the content; needs prose around it. Publish on docs site (mkdocs already deployed). |
| **MCP server leaderboard** | High | Static page on docs site ranking servers by coverage score and pass rate. Data exists for 50 servers. Becomes valuable once there's external traffic (blog post drives traffic). |
| **antvis CI integration PR** | Blocked on #292 merge | antvis maintainer asked us to add mcp-assert to their CI. Submit follow-up PR with `evals/` directory (25 assertions) + GitHub Actions workflow using `mcp-assert-action@v1`. This is the first external adoption. |
| **Download stats dashboard** | Medium | Script or tool that queries PyPI, npm, GitHub releases, and Homebrew APIs to produce a unified download report. Optionally append to CSV in repo for historical tracking. |
| **C# server suites** | Medium | `modelcontextprotocol/csharp-sdk` has examples. Last major language gap (7th language). |
| **Reference suite registry** | Medium | Canonical protocol conformance assertions any MCP server can run. Independent of server-specific fixtures. "Does this server speak MCP correctly?" |
| **Docker images** | Low | Per-runtime images (node, python, go) for running `mcp-assert audit/ci` without installing the binary. Useful for CI without install (`docker run ghcr.io/blackwell-systems/mcp-assert:node ci --suite evals/`) and as the backend for the hosted audit experiment. Not needed until hosted audit or Docker Hub pull metrics become a priority. |
| **Nix flake** | Low | Nix users are quality-focused and vocal. |

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

| PR/Issue | Repo | Status | What happens when it merges |
|----------|------|--------|----------------------------|
| antvis/mcp-server-chart#292 | Fix: isError on chart failures | Open, maintainer engaged | Submit CI integration PR immediately |
| grafana/mcp-grafana#793 | Fix: timestamp validation | **Merged** (2026-04-27) | Scorecard updated, assertion unskipped |
| mark3labs/mcp-go#828 | Fix: stderr hooks | Open | Update scorecard |
| modelcontextprotocol/servers#4044 | Fix: blob content type (community) | Open | Update scorecard, unskip filesystem assertion |
| modelcontextprotocol/servers#4051 | Fix: puppeteer_navigate isError | Open (archived branch) | Update scorecard, unskip assertion |
| sammcj/mcp-devtools#258 | Fix: isError instead of internal error | Open | Update scorecard, unskip assertions |
| steipete/Peekaboo#108 | Issue: internal error on missing perms | Open | Swift fix, not pursuing PR |

## Coverage Expansion Opportunities

| Server | Current | Potential | Blocker |
|--------|---------|-----------|---------|
| Playwright | 67% (14/21) | ~85% | click/hover/drag need snapshot element refs (multi-step chaining) |
| Google Storage | 35% (6/17) | ~80% | Needs GCP credentials (use skip_unless_env) |
| Grafana | 34% (17/50) | ~60% | Needs running Grafana instance (docker-compose with service container) |
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
| **getsentry/XcodeBuildMCP suite** | 0.6.0 | 10 assertions, 27 tools discovered, 100% clean. First macOS-specific server. Server #39. |
| **`mcp-assert audit` command** | 0.6.0 | Zero-config quality audit. Connects, discovers tools, calls each with schema-generated inputs, reports quality score. Generates starter YAML for CI. Discovery on-ramp to the YAML workflow. |
| **`skip_unless_env` field** | 0.6.0 | Conditional assertion skipping based on env vars. Live-backend and no-auth assertions coexist in same suite. |
| **Per-assertion Docker isolation** | 0.6.0 | `docker:` field in server YAML. Fresh container per assertion for safe write testing. |
| **Coverage expansion** | 0.6.0 | SQLite 100%, Memory 100%, engram 100%. Anthropic git 92%, Playwright 67%. |
| **11 new server suites** | 0.6.0 | 48 servers, 504 assertions, 6 languages, 20 bugs. XcodeBuildMCP, Puppeteer, Context7, Chrome DevTools, Firefox DevTools, Excalidraw, SEC EDGAR, mcp-devtools, mcp-math, DuckDuckGo, Kubernetes, plus Perplexity, Peekaboo, CodeGraphContext, deep-research from pre-release. |
| **pytest plugin** | 0.5.0 | `pip install pytest-mcp-assert`. Published to PyPI via release pipeline. |
| **Badge snippet on pass** | 0.5.0 | CLI and GitHub Action output ready-to-paste badge markdown. |
| **SSE transport fix** | 0.4.0 | `Start()` missing for SSE/HTTP clients. Found by dogfooding. |
