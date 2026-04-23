# Examples

mcp-assert ships with example assertions for four MCP servers in three languages.

## Filesystem server — TypeScript

**Directory:** `examples/filesystem/`

Tests the official `@modelcontextprotocol/server-filesystem`. 14 assertions: read file, read multiple files, read text file, list directory, list directory with sizes, directory tree, get file info, search files, write file, edit file, create directory, move file, list allowed directories, and a **negative test** that verifies path traversal is rejected.

```bash
npm install -g @modelcontextprotocol/server-filesystem
mcp-assert run --suite examples/filesystem --fixture examples/filesystem/fixtures
```

## Memory server — TypeScript

**Directory:** `examples/memory/`

Tests the official `@modelcontextprotocol/server-memory`. 5 assertions with **stateful setup**: create entities, add observations, create relations, search nodes, and verify empty search returns nothing.

```bash
npm install -g @modelcontextprotocol/server-memory
mcp-assert run --suite examples/memory
```

## SQLite server — Python

**Directory:** `examples/sqlite/`

Tests the official `mcp-server-sqlite` (Python). 6 assertions: list tables, SELECT queries, COUNT, JOINs, describe table schema, and error handling for invalid SQL. Fixture is a pre-built `.db` file.

```bash
uvx mcp-server-sqlite  # or: pip install mcp-server-sqlite
mcp-assert run --suite examples/sqlite --fixture examples/sqlite/fixtures
```

## agent-lsp — Go

**Directory:** `examples/agent-lsp-go/`

Tests [agent-lsp](https://github.com/blackwell-systems/agent-lsp) with gopls. 51 assertions covering all 50 tools: navigation, refactoring, analysis, session lifecycle, workspace, and build.

```bash
mcp-assert run --suite examples/agent-lsp-go --fixture /path/to/go/fixtures
```
