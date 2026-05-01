# Test Runner Integrations

mcp-assert assertions are defined in YAML. The CLI runs them directly, but you can also run them through your existing test framework. The same YAML files work across all five integrations: CLI, pytest, Vitest, Jest, Bun, and Go test.

All integrations follow the same architecture: a thin bridge that shells out to the `mcp-assert` binary with `--json` output and maps the result to the test framework's pass/fail/skip semantics. No MCP protocol logic is reimplemented; the Go binary handles everything.

## Vitest (TypeScript)

### Install

```bash
npm install -D vitest-mcp-assert @blackwell-systems/mcp-assert
```

The `@blackwell-systems/mcp-assert` package provides the binary. If you already have `mcp-assert` on PATH (via Homebrew, Go, or pip), you can skip it.

### Suite mode (auto-discover YAML files)

```ts
// mcp.test.ts
import { describeMcpSuite } from 'vitest-mcp-assert'

describeMcpSuite('mcp server assertions', 'evals/')
```

Each `.yaml` file in `evals/` becomes a separate Vitest test, named from the YAML's `name:` field:

```
 ✓ mcp server assertions > echo returns the input message (1204ms)
 ✓ mcp server assertions > greet says hello (1156ms)
 ✗ mcp server assertions > broken tool handles error gracefully (1089ms)
```

### Single assertion mode

```ts
import { test } from 'vitest'
import { runMcpAssert } from 'vitest-mcp-assert'

test('echo tool', () => runMcpAssert('evals/echo.yaml'))
test('greet tool', () => runMcpAssert('evals/greet.yaml', { timeout: '60s' }))
```

### Options

```ts
{
  fixture?: string   // Fixture directory (substituted for {{fixture}})
  server?: string    // Override server command
  timeout?: string   // Per-assertion timeout (default: "30s")
  binary?: string    // Path to mcp-assert binary (default: auto-detect)
}
```

## pytest (Python)

### Install

```bash
pip install pytest-mcp-assert
```

The binary is resolved from PATH or from the `mcp-assert` PyPI package (which bundles platform-specific binaries).

### Usage

```bash
pytest --mcp-suite evals/
```

Each `.yaml` file becomes a pytest test item with pass/fail/skip semantics.

### Configuration via pyproject.toml

```toml
[tool.pytest.ini_options]
mcp_suite = "evals/"
mcp_fixture = "fixtures/"
mcp_timeout = "30s"
```

Then just run `pytest` with no extra flags.

### CLI options

| Flag | Description |
|------|-------------|
| `--mcp-suite <dir>` | Directory containing YAML assertion files |
| `--mcp-fixture <dir>` | Fixture directory (substituted for `{{fixture}}`) |
| `--mcp-server <cmd>` | Override server command for all assertions |
| `--mcp-timeout <dur>` | Per-assertion timeout (default: 30s) |
| `--mcp-binary <path>` | Explicit path to mcp-assert binary |

## Binary resolution

Both integrations find the `mcp-assert` binary using a 3-tier strategy:

1. **Explicit path**: `binary` option (Vitest) or `--mcp-binary` flag (pytest)
2. **PATH lookup**: finds `mcp-assert` on your system PATH
3. **Package binary**: resolves from `@blackwell-systems/mcp-assert` (npm) or `mcp_assert` (PyPI)

## Same YAML, everywhere

The assertion YAML format is identical regardless of how you run it. A file that works with `mcp-assert run` works with `pytest --mcp-suite` and `describeMcpSuite()`. No migration, no format differences. Switch runners without changing assertions.

```yaml
name: search returns results for known query
server:
  command: npx
  args:
    - -y
    - my-mcp-server
assert:
  tool: search
  args:
    query: "test"
  expect:
    not_error: true
    contains:
      - "test"
timeout: 30s
```

## Jest (JavaScript/TypeScript)

### Install

```bash
npm install -D jest-mcp-assert @blackwell-systems/mcp-assert
```

### Usage

Auto-discover all YAML files:

```ts
import { describeMcpSuite } from 'jest-mcp-assert'
describeMcpSuite('mcp server', 'evals/')
```

Or run individual assertions:

```ts
import { runMcpAssert } from 'jest-mcp-assert'
test('echo tool', () => { runMcpAssert('evals/echo.yaml') })
```

Same YAML files as pytest and Vitest. See `jest-plugin/README.md` for all options.

## Bun

### Install

```bash
bun add -d bun-mcp-assert @blackwell-systems/mcp-assert
```

### Usage

```ts
import { describeMcpSuite } from "bun-mcp-assert"
describeMcpSuite("mcp server", "evals/")
```

Or individually:

```ts
import { test } from "bun:test"
import { runMcpAssert } from "bun-mcp-assert"
test("echo tool", () => { runMcpAssert("evals/echo.yaml") })
```

Uses native Bun APIs (`Bun.spawnSync`, `Bun.which`). Ships as TypeScript source with no build step. See `bun-plugin/README.md` for all options.

## Go test

### Install

```bash
go get github.com/blackwell-systems/mcp-assert/go-plugin
```

### Usage

Run a single assertion:

```go
func TestEchoTool(t *testing.T) {
    mcpassert.Run(t, "evals/echo.yaml")
}
```

Auto-discover all YAML files in a directory:

```go
func TestMCPServer(t *testing.T) {
    mcpassert.Suite(t, "evals/")
}
```

Each YAML file becomes a `t.Run` subtest. See `go-plugin/README.md` for all options.
