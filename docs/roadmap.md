# Roadmap

## Next Up

| Item | Priority | Description |
|------|----------|-------------|
| **MCP server leaderboard** | High | Static page on the docs site ranking servers by coverage score. Data already exists for 28 servers. Free SEO, gives server authors a reason to improve and adopt mcp-assert. |
| **C# server suites** | Medium | C# MCP servers remain untested. `modelcontextprotocol/csharp-sdk` has examples. Last major language gap. |
| **Reference suite registry** | Medium | A canonical set of protocol conformance assertions any MCP server can run against, independent of server-specific fixtures. Single source of truth for "does this server speak MCP correctly?" |
| **MCP registry integration** | Medium | Surface the mcp-assert badge on Glama and Smithery listings. Servers that pass a reference suite get a "verified" marker. |
| **Blog post** | Ready | "We tested 28 MCP servers from Anthropic, Google, OpenAI, and Microsoft. Here's what we found." The scorecard data is the content; needs prose around it. |
| **Nix flake** | Low | Nix users are quality-focused and vocal. High signal in a niche community. |

## MCP Protocol Coverage

10 of 12 MCP protocol methods covered. Two gaps remain:

| Protocol area | Status |
|--------------|--------|
| **Cancellation** (`$/cancelRequest`) | Not covered |
| **Ping** keepalive | Not covered |

Everything else is covered: tools, resources, prompts, sampling, roots, elicitation, progress, logging, pagination, completion.

## Assertion Engine

| Item | Priority | Description |
|------|----------|-------------|
| **Structured recovery actions** | Medium | When an assertion fails, return machine-readable guidance ("call tool X to fix") not just an error string. Agents consuming mcp-assert output could self-correct. |
| **Invariant drift detection** | Medium | Snapshot state before a tool call, compare after. Like `file_unchanged` but for arbitrary state. |

## Bigger Bets

| Item | Priority | Description |
|------|----------|-------------|
| **VS Code extension** | Low | Run assertions from the editor. Click-to-run on YAML files, inline pass/fail markers. |

## Recently Shipped

| Item | Version | Description |
|------|---------|-------------|
| **pytest plugin** | Unreleased | `pip install pytest-mcp-assert`, then `pytest --mcp-suite evals/`. Each YAML becomes a pytest test item. |
| **Badge snippet on pass** | Unreleased | CLI and GitHub Action output ready-to-paste badge markdown when all assertions pass. |
| **SSE transport fix** | 0.4.0 | `Start()` call was missing for SSE/HTTP clients. Found by dogfooding. |
| **FastMCP SSE suite** | 0.4.0 | 11 assertions against FastMCP over SSE transport. First SSE coverage. |
| **generate HTTP/SSE support** | 0.4.0 | `--transport http\|sse` and `--headers` flags for remote server assertion generation. |
| **Grafana, Playwright, Google, OpenAI, arxiv suites** | Unreleased | 28 servers scanned total. All 4 major tech companies covered. |
| **4 upstream fix PRs** | Unreleased | mcp-go, antvis, grafana, arxiv. |
| **External adoption** | Unreleased | antvis maintainer engaged, asked about mcp-assert. Anthropic filesystem bug fix submitted by community member via our issue. |
