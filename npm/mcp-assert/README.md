# mcp-assert

The testing standard for MCP servers. Lint, test, and fuzz over real stdio/SSE/HTTP transport.

A single Go binary that acts as an MCP client: connects to your server, calls your tools with predefined inputs, and evaluates responses against assertions you define in YAML. Includes 24 static analysis rules with auto-fix.

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
# Zero-config audit (finds crashes)
mcp-assert audit --server "npx my-mcp-server"

# Static analysis (finds schema issues)
mcp-assert lint --server "npx my-mcp-server"

# Auto-fix schema issues
mcp-assert lint --server "npx my-mcp-server" --fix

# Adversarial fuzzing (finds edge cases)
mcp-assert fuzz --server "npx my-mcp-server"

# Run YAML assertions
mcp-assert run --suite evals/

# CI with threshold
mcp-assert ci --suite evals/ --threshold 95
```

## Lint (24 Rules)

Static analysis catches issues before runtime:

```bash
mcp-assert lint --server "npx my-mcp-server"

  E  E103   create_user    Required parameter "email" has no description
  W  W114   update_user    Input schema is 4 levels deep
  W  W112   (server)       Server exposes 27 tools. LLM accuracy degrades beyond 20

# Auto-generate fixes
mcp-assert lint --server "npx my-mcp-server" --fix

# Strict mode for CI (warnings become errors)
mcp-assert lint --server "npx my-mcp-server" --strict
```

## Vitest Integration

```bash
npm install -D @blackwell-systems/vitest-mcp-assert
```

```ts
import { describeMcpSuite } from '@blackwell-systems/vitest-mcp-assert'
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
- [Error Reference](https://github.com/blackwell-systems/mcp-assert/blob/main/docs/ERROR_REFERENCE.md) (24 error codes)
- [Scorecard](https://blackwell-systems.github.io/mcp-assert/scorecard/) (102 servers scanned, 4,794 issues found)
- [GitHub Action](https://github.com/marketplace/actions/mcp-assert)
- [Vitest plugin](https://www.npmjs.com/package/@blackwell-systems/vitest-mcp-assert)
- [pytest plugin](https://pypi.org/project/pytest-mcp-assert/)

## License

MIT
