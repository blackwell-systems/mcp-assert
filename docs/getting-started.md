# Getting Started

## Install

```bash
go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest
```

## Scaffold your first assertion

```bash
mcp-assert init evals
```

This creates `evals/read_file.yaml` (a commented assertion template) and `evals/fixtures/hello.txt` (a fixture file). Edit the YAML to point at your MCP server, then run it:

```bash
mcp-assert run --suite evals/ --fixture evals/fixtures
```

You should see:

```
PASS  read_file returns file contents  1203ms

1 passed
```

## Write an assertion by hand

If you already know which server you want to test, write the assertion directly:

```yaml
# evals/read_file.yaml
name: read_file returns file contents
server:
  command: npx
  args: ["@modelcontextprotocol/server-filesystem", "{{fixture}}"]
assert:
  tool: read_file
  args:
    path: "{{fixture}}/hello.txt"
  expect:
    not_error: true
    contains: ["Hello, world!"]
```

```bash
mcp-assert run --suite evals/ --fixture ./fixtures
```

## Any language, same assertions

Works the same for a Go server, a Python server, or anything else that speaks MCP — just change `server.command`:

```yaml
# Python server
server:
  command: python
  args: ["-m", "my_mcp_server"]
```

```yaml
# Go server
server:
  command: agent-lsp
  args: ["go:gopls"]
```

## Zero-Effort Coverage

Get from zero to 100% coverage in three commands:

```bash
# 1. Generate stub assertions for every tool the server exposes
mcp-assert generate --server "my-mcp-server" --output evals/ --fixture ./fixtures

# 2. Capture actual outputs as snapshots
mcp-assert snapshot --suite evals/ --server "my-mcp-server" --update

# 3. Assert nothing changed
mcp-assert run --suite evals/ --server "my-mcp-server"
```

`generate` queries `tools/list`, reads input schemas, and creates one YAML per tool with sensible defaults. `snapshot --update` captures real outputs. `run` asserts against them. Edit the generated YAMLs to replace `TODO` placeholders with real values.

## Next steps

- [Writing Assertions](writing-assertions.md) — YAML format, all 14 assertion types, setup steps, capture, fixtures
- [CLI Reference](cli.md) — full command reference with flags and examples
- [CI Integration](ci-integration.md) — GitHub Action, JUnit XML, regression detection
