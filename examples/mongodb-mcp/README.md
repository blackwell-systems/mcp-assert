# MongoDB MCP Server Examples

These assertions test [mongodb-js/mongodb-mcp-server](https://github.com/mongodb-js/mongodb-mcp-server) (1K stars), the official MongoDB MCP server.

## Setup

```bash
npx -y mongodb-mcp-server@latest --readOnly
```

Most tools require a MongoDB connection string. Without one, the knowledge/assistant tools and error handling can be tested.

For full coverage, set a connection string:

```bash
MDB_MCP_CONNECTION_STRING="mongodb://localhost:27017" mcp-assert run --suite examples/mongodb-mcp
```

## Coverage

4 assertions covering 4/15 tools:

| Tool | Notes |
|------|-------|
| `list-knowledge-sources` | Works without connection |
| `search-knowledge` | Built-in MongoDB knowledge base |
| `find` | Error handling (no connection) |
| `connect` | Error handling (invalid connection string) |

Not covered: `aggregate`, `collection-indexes`, `collection-schema`, `collection-storage-size`, `count`, `db-stats`, `explain`, `export`, `list-collections`, `list-databases`, `mongodb-logs` (all require a live MongoDB instance).

## Notes

Clean scan: no bugs found. Error messages are exemplary: clear, actionable, and include LLM-specific guidance ("do not invent connection strings"). This is the gold standard for MCP error handling.
