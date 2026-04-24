# GitHub MCP Server Examples

These assertions test [github/github-mcp-server](https://github.com/github/github-mcp-server), the most popular MCP server (28K+ stars).

## Setup

The assertions reference `github-mcp-server` in PATH. Build and install the server before running:

```bash
git clone --depth 1 https://github.com/github/github-mcp-server.git /tmp/github-mcp-server
cd /tmp/github-mcp-server && go build -o github-mcp-server ./cmd/github-mcp-server
```

Then add the binary to your PATH, or override the command when running:

```bash
mcp-assert run --suite examples/github-mcp --server "/tmp/github-mcp-server/github-mcp-server stdio"
```

## Authentication

Requires a GitHub personal access token with `repo` and `read:user` scopes. Create one at https://github.com/settings/tokens and set it as an environment variable:

```bash
GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_TOKEN mcp-assert run --suite examples/github-mcp
```

## Coverage

6 assertions covering read-only tools: `get_me`, `search_repositories`, `get_file_contents`, `list_issues`, `search_code`, `list_branches`.
