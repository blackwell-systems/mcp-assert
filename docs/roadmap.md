# Roadmap

## Next Up

| Item | Priority | Description |
|------|----------|-------------|
| **`mcp-assert ui` command** | High | Web-based GUI served by the Go binary. Three modes: Explorer (connect to a server, browse tools, call them interactively), Tracer (live tool call timeline between agent and server via WebSocket), Debugger (visual assertion failure inspector with request/response diff). Frontend embedded via `embed.FS`, no separate install. Reuses existing `createMCPClient`, `generateArgsFromSchema`, `runAssertion`, `auditSingleTool`. This is the foundation for hosted audit and the quality registry. |
| **Blog post** | Ready | "We tested 38 MCP servers from Anthropic, Google, OpenAI, Microsoft, and AWS. Here's what we found." The scorecard data is the content; needs prose around it. Publish on docs site (mkdocs already deployed). |
| **MCP server leaderboard** | High | Static page on docs site ranking servers by coverage score and pass rate. Data exists for 39 servers. Becomes valuable once there's external traffic (blog post drives traffic). |
| **antvis CI integration PR** | Blocked on #292 merge | antvis maintainer asked us to add mcp-assert to their CI. Submit follow-up PR with `evals/` directory (25 assertions) + GitHub Actions workflow using `mcp-assert-action@v1`. This is the first external adoption. |
| **C# server suites** | Medium | `modelcontextprotocol/csharp-sdk` has examples. Last major language gap (7th language). |
| **Reference suite registry** | Medium | Canonical protocol conformance assertions any MCP server can run. Independent of server-specific fixtures. "Does this server speak MCP correctly?" |
| **Nix flake** | Low | Nix users are quality-focused and vocal. |

## `mcp-assert ui` Design

### Architecture

```
mcp-assert ui --server "npx my-server" --port 7890

┌─────────────────────────────────┐
│  Go binary (mcp-assert)        │
│  ├─ HTTP server (embed.FS)     │
│  ├─ WebSocket (live trace)     │
│  ├─ REST API (/api/tools,      │
│  │   /api/call, /api/run)      │
│  └─ MCP client (reuses all     │
│      existing runner code)      │
└─────────────────────────────────┘
```

### Four modes (two phases)

**Phase 1 (launch)**: Explorer + Debugger. Self-contained, no LLM key needed, demonstrates core value.

**Explorer**: Connect to any MCP server. See all tools, prompts, resources in a tree. Click a tool to see its JSON Schema. Click "Call" to invoke with editable args. Response displayed with syntax highlighting. "Save as assertion" button turns any call into a YAML test case. Interactive version of the audit command.

**Debugger**: Run a suite from the UI. Failures appear in a list. Click a failure: see request, actual response, expected values, specific expectation that failed. Side-by-side diff view. "Fix" button suggests YAML edits (visual version of `--fix` mode). "Export suite" generates YAML + GitHub Actions workflow.

**Phase 2 (after launch)**: Agent + Tracer. Require LLM config and WebSocket proxy infrastructure.

**Agent**: Connect an LLM (OpenAI, Anthropic, etc.), let it drive the server's tools via ReAct loop. Watch the tool call chain in real time. Tool confirmation mode (approve/deny before execution). Record the full session as a trajectory YAML for CI regression testing. This is ProtoMCP's agent mode plus assertions.

**Tracer**: Proxy between an external agent (Claude Code, Cursor, etc.) and an MCP server. Every tool call appears in a live timeline via WebSocket. Click to expand: request args, response body, duration, isError. Filter by tool name, status, duration. Export session as trajectory YAML. Builds on the existing `intercept` command.

**The funnel**: Explorer ("does my server work?") leads to Debugger ("why did this fail?") leads to Agent ("how does an LLM use my tools?") leads to Tracer ("what is my production agent doing?"). Each mode feeds the next.

### Frontend stack

**Preact + Tailwind CSS**, compiled via esbuild to a single `bundle.js`, embedded in the Go binary via `//go:embed`. Same API as React (JSX, useState, useEffect), 3KB instead of 45KB. esbuild compiles in ~50ms.

**Dev workflow**: edit JSX, run `esbuild` (one command, 50ms), built JS committed to repo. Users never run a build step; the frontend is already inside the Go binary they download.

**Why Preact over alternatives**:
- vs React: same API, 1/15th the size. Matters for an embedded binary.
- vs Vanilla JS: component reuse (ToolCard, SchemaForm, TraceEntry), reactive state for WebSocket streams, list rendering. Vanilla JS becomes unmanageable at 10+ interactive components.
- vs HTMX: wrong fit for real-time WebSocket data streams and complex client-side state (trace timeline, form editing).

**Inspiration from ProtoMCP** (SahanUday/ProtoMCP): three-column layout (server list | main content | JSON-RPC log panel), auto-generated forms from JSON Schema, real-time trace timeline with color-coded events, tool confirmation mode for destructive calls. Our differentiation: "Save as assertion" button, expected vs actual diff, CI export, all three transports (stdio/SSE/HTTP), and the testing/assertion layer ProtoMCP completely lacks.

### Scaling path

The single binary with embedded UI scales for the local tool (one user, localhost, 1-10 servers). Grafana uses the same pattern at millions of lines of frontend TypeScript.

For the hosted platform (multi-user, persistent storage, queued jobs, billing), the same Go engine (`internal/runner`, `internal/assertion`, `internal/report`) gets wrapped in a production web service with a database, auth, and CDN-served frontend. No rewrite; the local UI is both a standalone product and a prototype for the hosted version.

```
Phase 1: mcp-assert ui       → single binary, localhost, embedded frontend
Phase 2: mcp-assert-cloud    → deployed service, same Go engine, production frontend
```

Build local first. Adoption proves demand. Demand justifies hosted.

## Platform Direction

The `ui` command is the local version. The platform is the hosted version of the same UI, with accounts and persistence.

### Monetization sequence

```
OSS CLI (free) → UI local (free) → hosted audit (freemium) → registry (paid) → monitoring (SaaS)
```

| Tier | What | Pricing |
|------|------|---------|
| **Free (OSS)** | CLI, GitHub Action, all assertion types, local UI | Free forever |
| **Hosted audit** | Paste a server URL, get results in the browser. No CLI install. | Free: 5 audits/month. Paid: unlimited. |
| **Quality registry** | Public leaderboard. Server authors claim listings, add verified badge, show CI status. | Free listing. Verified badge: paid. |
| **Continuous monitoring** | Run assertion suite on schedule against live servers. Alert on regression (Slack, email, PagerDuty). | $29/mo per server, $99/mo teams |
| **Team dashboard** | Shared view of org's MCP servers, coverage, pass rates, trends. Role-based access, audit logs. | Enterprise pricing |

The quality registry (mcp-assert.dev) becomes the "npm audit for MCP": users check before adopting a server, authors add the badge for trust. Revenue comes from verified listings and continuous monitoring.

Viability depends on MCP ecosystem growth. If MCP becomes the standard agent-to-tool protocol (Anthropic, OpenAI, Google all pushing it), the quality layer is infrastructure.

## Open PRs and Issues

| PR/Issue | Repo | Status | What happens when it merges |
|----------|------|--------|----------------------------|
| antvis/mcp-server-chart#292 | Fix: isError on chart failures | Open, maintainer engaged | Submit CI integration PR immediately |
| grafana/mcp-grafana#793 | Fix: timestamp validation | Open, CLA signed | Update scorecard, unskip assertion |
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
| **Perplexity, Peekaboo, CodeGraphContext, deep-research suites** | 0.6.0 | 39 servers, 472 assertions, 6 languages, 15 bugs. |
| **pytest plugin** | 0.5.0 | `pip install pytest-mcp-assert`. Published to PyPI via release pipeline. |
| **Badge snippet on pass** | 0.5.0 | CLI and GitHub Action output ready-to-paste badge markdown. |
| **SSE transport fix** | 0.4.0 | `Start()` missing for SSE/HTTP clients. Found by dogfooding. |
