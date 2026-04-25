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

20 assertions covering read-only tools across 6 toolsets:

| Toolset | Tools | Assertions |
|---------|-------|------------|
| Context | `get_me` | 1 |
| Repos | `get_file_contents`, `search_repositories`, `search_code`, `list_branches`, `list_commits`, `list_tags`, `list_releases`, `get_latest_release`, `get_release_by_tag` | 11 |
| Git | `get_repository_tree` | 2 |
| Issues | `list_issues`, `search_issues` | 2 |
| Pull Requests | `list_pull_requests`, `search_pull_requests` | 2 |
| Users | `search_users` | 1 |
| Gists | `list_gists` | 1 |

All assertions are read-only. No write operations are performed against GitHub.
