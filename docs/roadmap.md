# Roadmap

## Next Up

| Item | Priority | Description |
|------|----------|-------------|
| **`mcp-assert audit` command** | High | Zero-config quality audit. `mcp-assert audit --server "npx my-server"` connects, lists tools, auto-generates inputs from JSON schemas, calls every tool, and reports: which crash, which return isError properly, which return internal errors. No YAML writing needed. Produces a quality score. This is the "point and shoot" version of what we do manually today with generate + run. The YAML workflow stays for CI regression testing; audit is for discovery. |
| **Blog post** | Ready | "We tested 38 MCP servers from Anthropic, Google, OpenAI, Microsoft, and AWS. Here's what we found." The scorecard data is the content; needs prose around it. Publish on docs site (mkdocs already deployed). |
| **MCP server leaderboard** | High | Static page on docs site ranking servers by coverage score and pass rate. Data exists for 38 servers. Becomes valuable once there's external traffic (blog post drives traffic). |
| **antvis CI integration PR** | Blocked on #292 merge | antvis maintainer asked us to add mcp-assert to their CI. Submit follow-up PR with `evals/` directory (25 assertions) + GitHub Actions workflow using `mcp-assert-action@v1`. This is the first external adoption. |
| **C# server suites** | Medium | `modelcontextprotocol/csharp-sdk` has examples. Last major language gap (7th language). |
| **Reference suite registry** | Medium | Canonical protocol conformance assertions any MCP server can run. Independent of server-specific fixtures. "Does this server speak MCP correctly?" |
| **Nix flake** | Low | Nix users are quality-focused and vocal. |

## Platform Direction (exploratory)

The idea: mcp-assert as the engine behind a broader MCP server quality platform.

- **Quality registry** (mcp-assert.dev): public page showing every MCP server's test results. Server authors claim their listing, add the badge, run CI. Users check before adopting.
- **Hosted audit**: paste your server's URL, get results in the browser. No CLI install.
- **ProtoMCP-style explorer**: visual server browser backed by real test data (ProtoMCP at protomcp.io is the GUI explorer; mcp-assert is the testing engine).

No concrete plan yet. Depends on adoption signal from antvis CI integration and PR merges. If server authors start using mcp-assert in CI, the platform play becomes viable.

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

## Bigger Bets

| Item | Priority | Description |
|------|----------|-------------|
| **VS Code extension** | Low | Run assertions from the editor. Click-to-run on YAML files. |

## Recently Shipped

| Item | Version | Description |
|------|---------|-------------|
| **`skip_unless_env` field** | Unreleased | Conditional assertion skipping based on env vars. Live-backend and no-auth assertions coexist in same suite. |
| **Per-assertion Docker isolation** | Unreleased | `docker:` field in server YAML. Fresh container per assertion for safe write testing. |
| **Coverage expansion** | Unreleased | SQLite 100%, Memory 100%, engram 100%. Anthropic git 92%, Playwright 67%. |
| **Perplexity, Peekaboo, CodeGraphContext, deep-research suites** | Unreleased | 38 servers, 462 assertions, 6 languages, 15 bugs. |
| **pytest plugin** | 0.5.0 | `pip install pytest-mcp-assert`. Published to PyPI via release pipeline. |
| **Badge snippet on pass** | 0.5.0 | CLI and GitHub Action output ready-to-paste badge markdown. |
| **SSE transport fix** | 0.4.0 | `Start()` missing for SSE/HTTP clients. Found by dogfooding. |
