# vitest-mcp-assert

Vitest integration for [mcp-assert](https://github.com/blackwell-systems/mcp-assert). Run MCP server assertion YAML files as Vitest tests.

## Install

```bash
npm install -D vitest-mcp-assert @blackwell-systems/mcp-assert
```

The `@blackwell-systems/mcp-assert` package provides the mcp-assert binary. Alternatively, install it via Homebrew (`brew install blackwell-systems/tap/mcp-assert`) or Go (`go install github.com/blackwell-systems/mcp-assert@latest`).

## Usage

### Suite mode (recommended)

Create a test file that auto-discovers all YAML assertions in a directory:

```ts
// mcp.test.ts
import { describeMcpSuite } from 'vitest-mcp-assert'

describeMcpSuite('mcp server assertions', 'evals/')
```

Each `.yaml` file becomes a separate test, named from the YAML's `name:` field:

```
 ✓ mcp server assertions > echo returns the input message (1204ms)
 ✓ mcp server assertions > greet says hello (1156ms)
 ✗ mcp server assertions > broken tool handles error gracefully (1089ms)
```

### Single assertion mode

For more control, import `runMcpAssert` directly:

```ts
// mcp.test.ts
import { test } from 'vitest'
import { runMcpAssert } from 'vitest-mcp-assert'

test('echo tool', () => runMcpAssert('evals/echo.yaml'))
test('greet tool', () => runMcpAssert('evals/greet.yaml', { timeout: '60s' }))
```

### Options

Both `describeMcpSuite` and `runMcpAssert` accept an options object:

```ts
{
  fixture?: string   // Fixture directory (substituted for {{fixture}})
  server?: string    // Override server command
  timeout?: string   // Per-assertion timeout (default: "30s")
  binary?: string    // Path to mcp-assert binary (default: auto-detect)
}
```

### YAML assertion format

```yaml
name: echo returns the input message
server:
  command: node
  args:
    - build/index.js
assert:
  tool: echo
  args:
    message: hello
  expect:
    contains:
      - hello
timeout: 30s
```

See [mcp-assert documentation](https://blackwell-systems.github.io/mcp-assert/) for the full assertion reference (14 assertion types, resources, prompts, trajectories, and more).

## Binary resolution

The plugin finds `mcp-assert` using a 3-tier strategy:

1. Explicit `binary` option
2. `mcp-assert` on `PATH`
3. `@blackwell-systems/mcp-assert` npm package

## License

MIT
