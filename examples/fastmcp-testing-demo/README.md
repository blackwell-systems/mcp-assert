# fastmcp testing_demo Examples

These assertions test the `testing_demo` example server from [PrefectHQ/fastmcp](https://github.com/PrefectHQ/fastmcp) (25K+ stars), the most popular Python MCP framework.

## Setup

The assertions reference `/tmp/fastmcp/examples/testing_demo/server.py`. Clone the fastmcp repository before running:

```bash
git clone --depth 1 https://github.com/PrefectHQ/fastmcp.git /tmp/fastmcp
```

Then run the assertions:

```bash
mcp-assert run --suite examples/fastmcp-testing-demo
```

If you cloned to a different location, use `--server` to override:

```bash
mcp-assert run --suite examples/fastmcp-testing-demo \
  --server "uvx fastmcp run /path/to/fastmcp/examples/testing_demo/server.py"
```

## Coverage

16 assertions covering all three MCP feature categories:

- **Tools (11):** `add` (4 assertions), `greet` (3), `async_multiply` (4). 100% tool coverage.
- **Resources (3):** `resources/list`, `resources/read` (static and parameterized)
- **Prompts (2):** `prompts/list`, `prompts/get` with arguments
