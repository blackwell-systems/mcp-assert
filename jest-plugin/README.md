# jest-mcp-assert

Jest integration for [mcp-assert](https://github.com/blackwell-systems/mcp-assert). Run MCP server assertion YAML files as Jest tests.

Same YAML files work across Jest, Vitest, pytest, and the CLI.

## Install

```bash
npm install -D jest-mcp-assert @blackwell-systems/mcp-assert
```

## Usage

### Suite mode (auto-discover YAML files)

```ts
// mcp.test.ts
import { describeMcpSuite } from 'jest-mcp-assert'

describeMcpSuite('my mcp server', 'evals/')
```

Each YAML file in `evals/` becomes a Jest test case.

### Single assertion mode

```ts
import { runMcpAssert } from 'jest-mcp-assert'

test('echo tool returns message', () => {
  runMcpAssert('evals/echo.yaml')
})

test('greet tool with timeout', () => {
  runMcpAssert('evals/greet.yaml', { timeout: '60s' })
})
```

### Options

```ts
runMcpAssert('evals/echo.yaml', {
  timeout: '30s',           // per-assertion timeout
  fixture: 'test/fixtures', // {{fixture}} substitution
  server: 'node server.js', // server override
  binary: '/path/to/mcp-assert', // explicit binary path
})
```

## How it works

jest-mcp-assert is a thin bridge (~100 lines). It shells out to the mcp-assert Go binary with `--json`, parses the result, and maps it to Jest pass/fail outcomes. The Go binary handles all MCP protocol logic.

Binary resolution (in order):
1. Explicit `binary` option
2. `mcp-assert` on PATH (brew, go install, curl install)
3. `@blackwell-systems/mcp-assert` npm package

## Writing assertions

Assertions are YAML files. See the [mcp-assert docs](https://blackwell-systems.github.io/mcp-assert) for the full reference.

```yaml
name: echo returns the input message
server:
  command: node
  args: ["server.js"]
assert:
  tool: echo
  args:
    message: "hello"
  expect:
    not_error: true
    contains: ["hello"]
```

## License

MIT
