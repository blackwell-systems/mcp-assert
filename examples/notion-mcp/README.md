# Notion MCP Server Examples

These assertions test [makenotion/notion-mcp-server](https://github.com/makenotion/notion-mcp-server) (4.2K stars), the official Notion MCP server.

## Setup

```bash
npx -y @notionhq/notion-mcp-server
```

Requires a Notion API token for full functionality. Without a token, tools respond without crashing but return auth-related errors.

```bash
OPENAPI_MCP_HEADERS='{"Authorization":"Bearer ntn_YOUR_TOKEN"}' mcp-assert run --suite examples/notion-mcp --server "npx -y @notionhq/notion-mcp-server"
```

## Coverage

22 assertions covering 22/22 tools (100%):

Users, search, blocks (list children, retrieve, update, delete), pages (retrieve, create, update, move), page properties, comments (retrieve, create), data sources (query, retrieve, update, create, list templates), databases.

## Notes

Clean scan: no bugs found. All 22 tools respond cleanly via stdio with `not_error: true` stubs. The server initializes and handles requests gracefully even without authentication.
