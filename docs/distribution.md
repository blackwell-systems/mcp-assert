# Distribution Strategy

## Core thesis

mcp-assert spreads through the bugs it finds, not through marketing. Every issue filed and PR opened on a popular MCP server is passive promotion with built-in credibility.

## The scan-and-contribute flywheel

```
Scan server → Find bugs → File issue/PR → Link mcp-assert → Maintainers discover tool → Adopt it
```

1. Run `mcp-assert coverage` + `mcp-assert run` against a popular MCP server
2. Document any bugs, spec violations, or missing error handling
3. File issue or open PR with the fix, including "Found by [mcp-assert](https://github.com/blackwell-systems/mcp-assert)"
4. Ship the assertion suite as an example in our repo
5. Server maintainers and their community discover mcp-assert organically

This is the playbook that made eslint, clippy, and staticcheck ubiquitous — they spread through the bugs they find.

### Proven: first result

| Server | Bug found | Issue |
|--------|-----------|-------|
| `@modelcontextprotocol/server-filesystem` | `read_media_file` returns `type: "blob"` — not a valid MCP content type per the 2025-11-25 spec | [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029) |

## Target servers

Prioritized by stars, community size, and likelihood of finding issues.

### Tier 1: Official Anthropic servers (highest credibility)

| Server | Language | Status | Stars |
|--------|----------|--------|-------|
| `server-filesystem` | TypeScript | 92% coverage, 1 bug filed | Part of modelcontextprotocol/servers |
| `server-memory` | TypeScript | Suite shipped, CI passing | Part of modelcontextprotocol/servers |
| `server-sqlite` | Python | Suite shipped, CI passing | Part of modelcontextprotocol/servers |
| `server-github` | TypeScript | Not started | Part of modelcontextprotocol/servers |
| `server-postgres` | TypeScript | Not started | Part of modelcontextprotocol/servers |
| `server-brave-search` | TypeScript | Not started | Part of modelcontextprotocol/servers |

### Tier 2: High-star community servers

| Server | Language | Stars | Why target |
|--------|----------|-------|------------|
| `mark3labs/mcp-go` | Go | 1000+ | Go MCP SDK — testing the SDK's example servers validates the ecosystem |
| `PrefectHQ/fastmcp` | Python | 25000+ | Most popular Python MCP framework — **11 assertions shipped, 100% tool coverage** |
| `supabase/mcp` | TypeScript | 500+ | Database tools — deterministic query assertions are a natural fit |
| `firebase/mcp` | TypeScript | — | Google ecosystem, high visibility |
| `stripe/agent-toolkit` | TypeScript | — | Payment tools — correctness matters |
| `linear/mcp-server` | TypeScript | — | Project management tools |

### Tier 3: Niche servers with engaged communities

| Server | Language | Why target |
|--------|----------|------------|
| `tavily/mcp-server` | Python | Search tools — response shape assertions |
| `browserbase/mcp-server` | TypeScript | Browser automation — state management testing |
| `Nix MCP servers` | Various | Nix community is quality-focused and vocal |

## Contribution format

For each server we scan:

### 1. Assertion suite (PR to our repo)

```
examples/<server-name>/
  ├── fixtures/           # Minimal test data
  ├── read_query.yaml     # One YAML per tool tested
  ├── list_items.yaml
  └── error_handling.yaml
```

### 2. Bug report (issue on their repo)

Template:
```markdown
## Bug: [tool_name] returns [problem]

**Found by:** automated testing with [mcp-assert](https://github.com/blackwell-systems/mcp-assert)

**Reproduction:**
[YAML assertion that demonstrates the bug]

**Expected:** [what the MCP spec says should happen]
**Actual:** [what happens]

**Spec reference:** [link to MCP spec section]
```

### 3. Fix PR (when the fix is obvious)

Even better than an issue — a PR with the fix and the assertion that proves it works. Include the assertion YAML in the PR description so the maintainer can verify.

## Distribution channels

### Frictionless adoption (highest priority)

| Channel | Status | Effort | Impact |
|---------|--------|--------|--------|
| **GitHub Action** | Planned | 1 day | `uses: blackwell-systems/mcp-assert-action@v1` — one line in any workflow |
| **GoReleaser** | Planned | 1 hour | Tagged releases, `go install ...@v0.1.0`, GitHub Releases binaries |
| **Homebrew** | Planned | 2 hours | `brew install mcp-assert` |
| **PyPI wrapper** | Planned | 1 day | `pip install mcp-assert` — downloads Go binary |
| **npm wrapper** | Planned | 1 day | `npx mcp-assert` — same pattern |

### Content (medium priority)

| Channel | Status | Description |
|---------|--------|-------------|
| **Blog post: dogfooding** | Draft material in `docs/dogfooding-findings.md` | "We built mcp-assert, used it on ourselves, and found 6 real bugs" |
| **Blog post: filesystem audit** | Material from scan results | "We tested Anthropic's official filesystem server and found a spec violation" |
| **MCP community Discord/forums** | Not started | Post when we have 3+ server suites with bugs found |
| **Hacker News** | Not started | Post when GitHub Action is live (frictionless try-it) |
| **r/LocalLLaMA, r/ClaudeAI** | Not started | After HN, when there's momentum |

### Passive (ongoing)

| Channel | Description |
|---------|-------------|
| **Bug reports on MCP servers** | Every issue filed mentions mcp-assert |
| **awesome-mcp-servers** | Already listed for agent-lsp; submit mcp-assert under testing tools |
| **Example suites in repo** | Each new suite is a landing page for that server's community |
| **Coverage reports** | `mcp-assert coverage` output in README shows the tool in action |

## Metrics to track

| Metric | Target | How to measure |
|--------|--------|---------------|
| Server suites shipped | 10 by end of Q2 | Count `examples/` directories |
| Bugs filed upstream | 5 by end of Q2 | GitHub issue links in dogfooding doc |
| External contributors | 1 by end of Q2 | GitHub contributor count |
| GitHub stars | 50 by end of Q2 | GitHub |
| GitHub Action installs | Track via marketplace | After Action ships |

## Non-goals

- **Paid tier** — mcp-assert is free and open source. The value is ecosystem positioning, not revenue.
- **SaaS dashboard** — no hosted version. CI-native, single binary.
- **LLM-as-judge features** — stay in our lane. Deterministic assertions only. Don't dilute the positioning.
