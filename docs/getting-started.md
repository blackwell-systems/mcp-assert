# Getting Started

> **Which path should I use?**
> - Just want to see what happens? Use `audit --server "my-server"` (zero-config, no YAML, instant results)
> - Want to stress-test with bad inputs? Use `fuzz --server "my-server"` (adversarial testing, no YAML)
> - Have a running MCP server? Use `init evals --server "my-server"` (generates stubs + captures baselines)
> - Want to start from a template? Use `init evals` (creates one commented YAML to customize)
> - Know your server config already? Jump to [Write an assertion by hand](#write-an-assertion-by-hand) below

## Install

Pick one:

```bash
# npm (no Go required)
npx @blackwell-systems/mcp-assert

# pip (no Go required)
pip install mcp-assert

# Go
go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest

# Homebrew
brew install blackwell-systems/tap/mcp-assert

# Scoop (Windows)
scoop bucket add blackwell-systems https://github.com/blackwell-systems/scoop-bucket
scoop install mcp-assert

# curl | sh (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/install.sh | sh
```

## Zero-setup testing (no YAML required)

Before writing any YAML, you can test any server with two commands:

```bash
# Quick health check: calls each tool once with valid input
mcp-assert audit --server "npx my-mcp-server"

# Adversarial testing: throws bad inputs at every tool
mcp-assert fuzz --server "npx my-mcp-server"
```

`audit` answers "does this server crash on normal input?" `fuzz` answers "does this server crash on bad input?" Both require zero setup, no YAML files, no configuration. They connect to the server, discover tools via `tools/list`, and test automatically.

Fuzz generates category-based adversarial inputs from each tool's JSON Schema: empty strings, null values, wrong types, missing required fields, boundary numbers, injection payloads, and random mutations. Every run is reproducible via `--seed`:

```bash
# Reproduce a specific fuzz run
mcp-assert fuzz --server "npx my-mcp-server" --seed 42 --runs 100
```

A tool **passes** if it responds (with content or `isError: true`). A tool **fails** if it crashes, hangs, or leaks internal errors (stack traces, panics). Rejecting bad input gracefully is correct behavior.

### Schema quality

Check tool schemas for agent usability issues without calling any tools:

```bash
mcp-assert lint --server "npx my-mcp-server"
```

Reports missing descriptions, untyped parameters, and free-text strings without constraints. Use `--threshold 0` in CI to gate on errors.

When you're ready for deeper testing (expected outputs, multi-step flows, state verification), move on to YAML assertions:

## One-step suite generation (recommended)

If you have a running MCP server, generate a complete test suite in one command:

```bash
mcp-assert init evals --server "my-mcp-server" --fixture ./fixtures
```

> If `--fixture` is omitted, `{{fixture}}` appears literally in generated YAMLs. Pass `--fixture ./fixtures` when your server tools use file paths.

This connects to your server, queries `tools/list`, generates one assertion stub per tool, and captures real response snapshots as baselines. You get 100% tool coverage with zero manual assertion writing.

```
Generating assertion stubs from tools/list...
  created read_file.yaml
  created list_directory.yaml
  created search_files.yaml

3 tools discovered, 3 stubs created, 0 skipped (already exist)

Capturing snapshots...
  NEW    read_file returns expected result
  NEW    list_directory returns expected result
  NEW    search_files returns expected result

Suite created successfully:
  Tools found:       3
  Stubs created:     3
  Snapshots captured: 3

Next steps:
  Run the suite:   mcp-assert run --suite evals --server "my-mcp-server"
```

Then run the suite:

```bash
mcp-assert run --suite evals/ --server "my-mcp-server"
```

Edit the generated YAMLs to replace `TODO` placeholders with realistic argument values and add more specific assertions (contains, json_path, etc.) as needed.

## Scaffold a template (manual)

If you prefer to write assertions by hand, scaffold a template without a server:

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

> **Tip:** If a position-sensitive assertion fails with "no identifier found" or "column is beyond end of line", run with `--fix` to get a suggested correction:
> ```bash
> mcp-assert run --suite evals/ --fix
> ```

You can also run a single YAML file directly instead of an entire directory:

```bash
mcp-assert run --suite evals/read_file.yaml --fixture evals/fixtures
```

This is useful for iterating on one assertion at a time during development.

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

Works the same for a Go server, a Python server, or anything else that speaks MCP: just change `server.command`:

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

Get from zero to 100% coverage in one command:

```bash
mcp-assert init evals --server "my-mcp-server" --fixture ./fixtures
```

This runs `generate` (stub YAMLs from `tools/list`) and `snapshot --update` (capture real outputs) in a single step. Then assert nothing changed:

```bash
mcp-assert run --suite evals/ --server "my-mcp-server"
```

You can also run the steps individually if you need more control:

```bash
# 1. Generate stub assertions for every tool the server exposes
mcp-assert generate --server "my-mcp-server" --output evals/ --fixture ./fixtures

# 2. Capture actual outputs as snapshots
mcp-assert snapshot --suite evals/ --server "my-mcp-server" --update

# 3. Assert nothing changed
mcp-assert run --suite evals/ --server "my-mcp-server"
```

`generate` queries `tools/list`, reads input schemas, and creates one YAML per tool with sensible defaults. `snapshot --update` captures real outputs. `run` asserts against them. Edit the generated YAMLs to replace `TODO` placeholders with real values.

## Auto-generate assertion stubs

Instead of writing every assertion by hand, generate stubs from your server's tool list:

```bash
mcp-assert generate --server "my-mcp-server" --output evals/ --fixture ./fixtures
```

This queries `tools/list`, reads each tool's input schema, and creates one YAML file per tool with sensible defaults. Tools detected as destructive (annotated with `destructiveHint: true` or not marked read-only) are generated with `skip: true` so they won't run until you review them.

To include all tools without skipping destructive ones:

```bash
mcp-assert generate --server "my-mcp-server" --output evals/ --include-writes
```

Edit the generated YAMLs to replace `TODO` placeholders with real values, then run them normally with `mcp-assert run`.

## Next steps

- [Writing Assertions](writing-assertions.md): YAML format, all 18 assertion types + 4 trajectory types, 8 block types, setup steps, capture, fixtures, trajectory assertions
- [CLI Reference](cli.md): full command reference with flags and examples
- [CI Integration](ci-integration.md): GitHub Action, JUnit XML, regression detection
- [Badge](badge.md): add the "Works with mcp-assert" badge to your README
- [Examples](examples.md): 65 suites across 58 servers in 8 languages (606 assertions)
