# rmcp Counter Server Examples

These assertions test the Counter example from [4t145/rmcp](https://github.com/4t145/rmcp), the Rust SDK for MCP.

## Setup

Build the example server from the rmcp repository:

```bash
git clone --depth 1 https://github.com/4t145/rmcp.git /tmp/rmcp
cd /tmp/rmcp && cargo build --example std_io
```

Then run the suite with the server path:

```bash
mcp-assert run --suite examples/rmcp-counter --server "/tmp/rmcp/target/debug/examples/std_io"
```

## Coverage

14 assertions covering 100% of tools (6/6) plus resources and prompts:

| Category | Tools/Methods | Assertions |
|----------|--------------|------------|
| Tools | `increment`, `decrement`, `get_value`, `say_hello`, `echo`, `sum` | 10 |
| Resources | `resources/list`, `resources/read` | 2 |
| Prompts | `prompts/list`, `prompts/get` | 2 |

## Known bugs

**`get_value` mutates state** ([4t145/rmcp#41](https://github.com/4t145/rmcp/issues/41)): The `get_value` tool is documented as "Get the current counter value" but actually decrements the counter. Two assertions target this bug and will fail until the upstream fix is merged.
