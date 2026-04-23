# mcp-assert

[![Blackwell Systems](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems)
[![Go](https://img.shields.io/badge/go-1.23+-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Test any MCP server in any language. No SDK required. No LLM required.

A single Go binary that connects to your MCP server over stdio, SSE, or HTTP, calls your tools, and asserts the results. Define assertions in YAML, run them in CI. Works with servers written in Go, TypeScript, Python, Rust, Java, or anything else that speaks MCP.

## Why

Most MCP tools are deterministic: `read_file` returns file contents, `read_query` returns rows, `get_references` returns locations. Given the same input, the correct output is knowable in advance. You don't need an LLM to grade it. You need `assert.Equal`.

### When to use what

| Your tool returns... | Use |
|---|---|
| Structured data (files, rows, locations, symbols) | **mcp-assert**: deterministic assertions |
| Predictable state changes (rename, create, delete) | **mcp-assert**: assert the state after |
| Error responses for bad input | **mcp-assert**: `is_error` and `contains` |
| Natural language (summaries, explanations, descriptions) | **LLM-as-judge**: quality is subjective |
| Creative content (commit messages, code suggestions) | **LLM-as-judge**: many correct answers |

Most MCP servers are heavy on the first three and light on the last two. If your server returns data, mcp-assert covers it. If your server generates prose, you need a different tool.

## Why not just write tests in Go/Python/etc?

You could. The assertion logic is straightforward. What you'd have to build yourself:

- **MCP protocol bootstrapping.** stdio transport, JSON-RPC framing, initialize/initialized handshake, tool call request/response lifecycle. This is ~200 lines of boilerplate per test suite, and easy to get wrong.
- **Server-agnostic test runner.** Your Go tests are coupled to your Go server. mcp-assert tests any server from any language with the same YAML. Switch `server.command` from `npx my-ts-server` to `python -m my_server` and the assertions don't change.
- **Eval-framework features.** pass@k/pass^k reliability metrics, baseline regression detection, JUnit XML output, Docker isolation, cross-language matrix mode. These are eval concerns, not unit test concerns.

The value isn't in the assertion logic. It's in not writing MCP client boilerplate, having one tool that works across every MCP server regardless of implementation language, and getting CI-grade reporting for free.

## Quick Start

```bash
# Install
go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest

# Scaffold your first assertion
mcp-assert init evals

# Run it
mcp-assert run --suite evals/ --fixture evals/fixtures
```

See the [Getting Started guide](https://blackwell-systems.github.io/mcp-assert/getting-started/) for a full walkthrough.

## Zero-Effort Coverage

```bash
# Generate stub assertions for every tool the server exposes
mcp-assert generate --server "my-mcp-server" --output evals/ --fixture ./fixtures

# Capture actual outputs as snapshots
mcp-assert snapshot --suite evals/ --server "my-mcp-server" --update

# Assert nothing changed
mcp-assert run --suite evals/ --server "my-mcp-server"
```

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

## CI Integration

Add to any GitHub Actions workflow with the [mcp-assert GitHub Action](https://github.com/blackwell-systems/mcp-assert-action):

```yaml
- name: Assert MCP server correctness
  run: |
    go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest
    mcp-assert ci --suite evals/ --threshold 95 --junit results.xml
```

See the [CI Integration guide](https://blackwell-systems.github.io/mcp-assert/ci-integration/) for JUnit XML, markdown summaries, badges, and regression detection.

## Documentation

Full documentation is available at [blackwell-systems.github.io/mcp-assert](https://blackwell-systems.github.io/mcp-assert):

- [Getting Started](https://blackwell-systems.github.io/mcp-assert/getting-started/): install, scaffold, first run
- [Writing Assertions](https://blackwell-systems.github.io/mcp-assert/writing-assertions/): YAML format, all 14 assertion types, setup steps, capture, fixtures
- [CLI Reference](https://blackwell-systems.github.io/mcp-assert/cli/): full command reference with flags and examples
- [Examples](https://blackwell-systems.github.io/mcp-assert/examples/): 4 example suites across 3 languages
- [CI Integration](https://blackwell-systems.github.io/mcp-assert/ci-integration/): GitHub Action, JUnit XML, regression detection
- [Architecture](https://blackwell-systems.github.io/mcp-assert/architecture/): internals and design decisions
- [Roadmap](https://blackwell-systems.github.io/mcp-assert/roadmap/): what's shipped and what's next
- [Dogfooding](https://blackwell-systems.github.io/mcp-assert/dogfooding/): real bugs found by testing our own servers

## License

MIT
