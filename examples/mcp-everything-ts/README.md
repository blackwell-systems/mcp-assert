# MCP Everything Server (TypeScript) Examples

These assertions test [`@modelcontextprotocol/server-everything`](https://github.com/modelcontextprotocol/servers/tree/main/src/everything), the official Anthropic reference MCP server. This is the canonical implementation used to validate MCP client libraries.

## Setup

```bash
npx -y @modelcontextprotocol/server-everything
```

Then run:

```bash
mcp-assert run --suite examples/mcp-everything-ts --server "npx -y @modelcontextprotocol/server-everything"
```

## Coverage

13 assertions covering 12/13 tools (92%):

| Tool | Assertions | Notes |
|------|------------|-------|
| `echo` | 1 | Echo back input |
| `get-annotated-message` | 1 | Annotated error/success/debug messages |
| `get-env` | 1 | Environment variable access |
| `get-resource-links` | 1 | Resource link references |
| `get-resource-reference` | 1 | Resource references |
| `get-structured-content` | 1 | Structured weather data (enum: New York/Chicago/LA) |
| `get-sum` | 1 | Numeric addition |
| `get-tiny-image` | 1 | Base64 image content |
| `gzip-file-as-resource` | 1 | Gzipped file as embedded resource |
| `toggle-simulated-logging` | 1 | Toggle log simulation |
| `toggle-subscriber-updates` | 1 | Toggle resource subscription updates |
| `trigger-long-running-operation` | 1 | Long-running operation with progress (~10s) |
| `simulate-research-query` | skipped | Requires `taskSupport` (not yet implemented in mcp-assert) |

## Notes

Clean scan: no bugs found. The reference server handles invalid input gracefully with proper MCP validation errors (enum validation, required field checks) instead of crashing with stack traces.
