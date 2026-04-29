# mcp-assert

Deterministic correctness testing for MCP servers. Works with any language, any transport.

A single Go binary that acts as an MCP client: connects to your server over stdio, SSE, or HTTP, calls your tools with predefined inputs, and evaluates the responses against assertions you define in YAML.

## Install

```bash
npx @blackwell-systems/mcp-assert
```

Or install globally:

```bash
npm install -g @blackwell-systems/mcp-assert
```

## Quick Start

```bash
# Zero-config audit
mcp-assert audit --server "npx my-mcp-server"

# Run YAML assertions
mcp-assert run --suite evals/

# CI with threshold
mcp-assert ci --suite evals/ --threshold 95
```

## Vitest Integration

```bash
npm install -D vitest-mcp-assert
```

```ts
import { describeMcpSuite } from 'vitest-mcp-assert'
describeMcpSuite('mcp server', 'evals/')
```

## GitHub Action

```yaml
- uses: blackwell-systems/mcp-assert-action@v1
  with:
    suite: evals/
```

## Links

- [Documentation](https://blackwell-systems.github.io/mcp-assert/)
- [GitHub](https://github.com/blackwell-systems/mcp-assert)
- [Scorecard](https://blackwell-systems.github.io/mcp-assert/scorecard/) (58 servers scanned, 25 bugs found)
- [GitHub Action](https://github.com/marketplace/actions/mcp-assert)
- [Vitest plugin](https://www.npmjs.com/package/vitest-mcp-assert)
- [pytest plugin](https://pypi.org/project/pytest-mcp-assert/)

## License

MIT
