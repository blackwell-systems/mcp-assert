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

### Three modes

**Explorer**: Connect to any MCP server. See all tools, prompts, resources in a tree. Click a tool to see its JSON Schema. Click "Call" to invoke with editable args. Response displayed with syntax highlighting. Interactive version of the audit command.

**Tracer**: Proxies between agent and server (builds on `intercept`). Every tool call appears in a live timeline via WebSocket. Click to expand: request args, response body, duration, isError. Filter by tool name, status, duration. Export session as trajectory YAML for regression testing.

**Debugger**: Run a suite from the UI. Failures appear in a list. Click a failure: see request, actual response, expected values, specific expectation that failed. Side-by-side diff view. "Fix" button suggests YAML edits (visual version of `--fix` mode).

### Frontend approach

Lightweight: Preact or vanilla JS with a CSS framework. No React/webpack/npm build step. Single `index.html` + JS bundle embedded in the Go binary. Zero external dependencies at runtime.

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
| **getsentry/XcodeBuildMCP suite** | Unreleased | 10 assertions, 27 tools discovered, 100% clean. First macOS-specific server. Server #39. |
| **`mcp-assert audit` command** | Unreleased | Zero-config quality audit. Connects, discovers tools, calls each with schema-generated inputs, reports quality score. Generates starter YAML for CI. Discovery on-ramp to the YAML workflow. |
| **`skip_unless_env` field** | Unreleased | Conditional assertion skipping based on env vars. Live-backend and no-auth assertions coexist in same suite. |
| **Per-assertion Docker isolation** | Unreleased | `docker:` field in server YAML. Fresh container per assertion for safe write testing. |
| **Coverage expansion** | Unreleased | SQLite 100%, Memory 100%, engram 100%. Anthropic git 92%, Playwright 67%. |
| **Perplexity, Peekaboo, CodeGraphContext, deep-research suites** | Unreleased | 39 servers, 472 assertions, 6 languages, 15 bugs. |
| **pytest plugin** | 0.5.0 | `pip install pytest-mcp-assert`. Published to PyPI via release pipeline. |
| **Badge snippet on pass** | 0.5.0 | CLI and GitHub Action output ready-to-paste badge markdown. |
| **SSE transport fix** | 0.4.0 | `Start()` missing for SSE/HTTP clients. Found by dogfooding. |
