# bun-mcp-assert

Bun test integration for [mcp-assert](https://github.com/blackwell-systems/mcp-assert). Run MCP server assertion YAML files as Bun tests.

Same YAML files work across Bun, Jest, Vitest, pytest, and the CLI.

## Install

```bash
bun add -d bun-mcp-assert @blackwell-systems/mcp-assert
```

## Usage

### Suite mode (auto-discover YAML files)

```ts
// mcp.test.ts
import { describeMcpSuite } from "bun-mcp-assert"

describeMcpSuite("my mcp server", "evals/")
```

Each YAML file in `evals/` becomes a Bun test case. Run with `bun test`.

### Single assertion mode

```ts
import { test } from "bun:test"
import { runMcpAssert } from "bun-mcp-assert"

test("echo tool returns message", () => {
  runMcpAssert("evals/echo.yaml")
})

test("greet tool with timeout", () => {
  runMcpAssert("evals/greet.yaml", { timeout: "60s" })
})
```

### Options

```ts
runMcpAssert("evals/echo.yaml", {
  timeout: "30s",           // per-assertion timeout
  fixture: "test/fixtures", // {{fixture}} substitution
  server: "bun server.ts",  // server override
  binary: "/path/to/mcp-assert", // explicit binary path
})
```

## How it works

bun-mcp-assert is a thin bridge. It uses `Bun.spawnSync` to call the mcp-assert Go binary with `--json`, parses the result, and maps it to Bun test pass/fail outcomes.

Binary resolution (in order):
1. Explicit `binary` option
2. `Bun.which("mcp-assert")` (PATH lookup, native)
3. `@blackwell-systems/mcp-assert` npm package

## License

MIT
