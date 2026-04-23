# Examples

mcp-assert ships with example assertions for five MCP servers in three languages. All built-in examples use stdio transport (the default). For HTTP/SSE transport examples, see [Writing Assertions](writing-assertions.md#httpsse-transport).

## Filesystem server. TypeScript

**Directory:** `examples/filesystem/`

Tests the official `@modelcontextprotocol/server-filesystem`. 14 assertions: read file, read multiple files, read text file, list directory, list directory with sizes, directory tree, get file info, search files, write file, edit file, create directory, move file, list allowed directories, and a **negative test** that verifies path traversal is rejected.

```bash
npm install -g @modelcontextprotocol/server-filesystem
mcp-assert run --suite examples/filesystem --fixture examples/filesystem/fixtures
```

## Memory server. TypeScript

**Directory:** `examples/memory/`

Tests the official `@modelcontextprotocol/server-memory`. 5 assertions with **stateful setup**: create entities, add observations, create relations, search nodes, and verify empty search returns nothing.

```bash
npm install -g @modelcontextprotocol/server-memory
mcp-assert run --suite examples/memory
```

## SQLite server. Python

**Directory:** `examples/sqlite/`

Tests the official `mcp-server-sqlite` (Python). 6 assertions: list tables, SELECT queries, COUNT, JOINs, describe table schema, and error handling for invalid SQL. Fixture is a pre-built `.db` file.

```bash
uvx mcp-server-sqlite  # or: pip install mcp-server-sqlite
mcp-assert run --suite examples/sqlite --fixture examples/sqlite/fixtures
```

## fastmcp testing_demo. Python

**Directory:** `examples/fastmcp-testing-demo/`

Tests the `testing_demo` example server from [PrefectHQ/fastmcp](https://github.com/PrefectHQ/fastmcp) (25K stars), the most popular Python MCP framework. 16 assertions covering all three MCP feature categories:

- **Tools (11):** `add` (sum, negative numbers, zero, missing argument error), `greet` (default greeting, custom greeting, empty name), and `async_multiply` (product, zero, negative, fractional). 100% tool coverage.
- **Resources (3):** `resources/list` (verifies `demo://info` is exposed), `resources/read demo://info` (static server info), `resources/read demo://greeting/Alice` (parameterized template resource).
- **Prompts (2):** `prompts/list` (verifies `hello` and `explain` are exposed), `prompts/get hello` with `name: "Alice"` argument.

```bash
git clone --depth 1 https://github.com/PrefectHQ/fastmcp.git /tmp/fastmcp
mcp-assert run --suite examples/fastmcp-testing-demo
```

!!! note
    The assertions reference the server at `/tmp/fastmcp/examples/testing_demo/server.py`. Clone the fastmcp repo to `/tmp/fastmcp` before running, or use `--server` to override the path.

## agent-lsp. Go

**Directory:** `examples/agent-lsp-go/`

Tests [agent-lsp](https://github.com/blackwell-systems/agent-lsp) with gopls. 60 assertions covering all 50 tools: navigation, refactoring, analysis, session lifecycle, workspace, and build.

```bash
mcp-assert run --suite examples/agent-lsp-go --fixture /path/to/go/fixtures
```
