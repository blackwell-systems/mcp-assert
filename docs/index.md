# mcp-assert

Test any MCP server in any language. No SDK required. No LLM required.

A single Go binary that connects to your MCP server over stdio, SSE, or HTTP, calls your tools, and asserts the results. Define assertions in YAML, run them in CI. Works with servers written in Go, TypeScript, Python, Rust, Java: anything that speaks MCP.

## Why

Most MCP tools are deterministic: `read_file` returns file contents, `read_query` returns rows, `get_references` returns locations. Given the same input, the correct output is knowable in advance. You don't need an LLM to grade it: you need `assert.Equal`.

Existing MCP eval frameworks use LLM-as-judge for everything: send a prompt, get a response, ask GPT "was this good?" on a 1-5 scale. This adds cost, latency, and false variance to tests that should be instant and exact.

mcp-assert tests MCP server tools the way you test code: given this input, assert this output.

### When to use what

| Your tool returns... | Use |
|---|---|
| Structured data (files, rows, locations, symbols) | **mcp-assert**: deterministic assertions |
| Predictable state changes (rename, create, delete) | **mcp-assert**: assert the state after |
| Error responses for bad input | **mcp-assert**. `is_error` and `contains` |
| Natural language (summaries, explanations, descriptions) | **LLM-as-judge**: quality is subjective |
| Creative content (commit messages, code suggestions) | **LLM-as-judge**: many correct answers |

Most MCP servers are heavy on the first three and light on the last two. If your server returns data, mcp-assert covers it. If your server generates prose, you need a different tool.

## Quick Start

```bash
# Install (pick one)
npx @blackwell-systems/mcp-assert                                          # npm (no Go required)
pip install mcp-assert                                                      # pip (no Go required)
go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest   # Go
brew install blackwell-systems/tap/mcp-assert                               # Homebrew
scoop install mcp-assert                                                    # Scoop (Windows)
curl -fsSL https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/install.sh | sh

# Audit a server in one command (zero-config, no YAML)
mcp-assert audit --server "npx my-mcp-server"

# Or scaffold assertions and run them
mcp-assert init evals                   # Or: init evals --server "my-server"
mcp-assert run --suite evals/ --fixture evals/fixtures
```

See [Getting Started](getting-started.md) for a full walkthrough.

## How It Differs

| Dimension | Existing MCP evals | mcp-assert |
|---|---|---|
| Grading | LLM-as-judge (subjective, costly) | Deterministic assertions (exact, free) |
| Speed | Seconds per test (LLM round-trip) | Milliseconds per test (no LLM) |
| CI cost | API calls on every run | Zero external dependencies |
| Reliability | Not measured | pass@k / pass^k per assertion |
| Regression | Not supported | Baseline comparison, fail on backslide |
| Docker | Not supported | Per-assertion container isolation |
| Multi-language | Not supported | Same assertion across N language servers |
| Bidirectional MCP | Not supported | Client capabilities: roots, sampling, elicitation |
| Trajectory testing | Not supported | Tool call sequence validation (no server needed) |
