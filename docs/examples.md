# Examples

mcp-assert ships with 18 suites (17 server suites + 1 trajectory suite) for 12 MCP servers in three languages (Go, TypeScript, Python), plus a trajectory suite that runs without a server. All built-in server suites use stdio transport except the HTTP conformance suite. 188 total assertions.

## Summary

| Suite | Server | Language | Transport | Assertions |
|-------|--------|----------|-----------|------------|
| `examples/filesystem/` | `@modelcontextprotocol/server-filesystem` | TypeScript | stdio | 14 |
| `examples/memory/` | `@modelcontextprotocol/server-memory` | TypeScript | stdio | 5 |
| `examples/sqlite/` | `mcp-server-sqlite` | Python | stdio | 6 |
| `examples/fastmcp-testing-demo/` | PrefectHQ/fastmcp testing_demo | Python | stdio | 16 |
| `examples/agent-lsp-go/` | agent-lsp + gopls | Go | stdio | 63 |
| `examples/mcp-go-everything/` | mark3labs/mcp-go everything | Go | stdio | 9 |
| `examples/mcp-go-everything-http/` | mark3labs/mcp-go everything | Go | HTTP | 5 |
| `examples/mcp-go-everything-prompts/` | mark3labs/mcp-go everything | Go | stdio | 4 |
| `examples/mcp-go-everything-resources/` | mark3labs/mcp-go everything | Go | stdio | 4 |
| `examples/mcp-go-typed-tools/` | mark3labs/mcp-go typed_tools | Go | stdio | 3 |
| `examples/mcp-go-structured/` | mark3labs/mcp-go structured | Go | stdio | 6 |
| `examples/mcp-go-roots/` | mark3labs/mcp-go roots_server | Go | stdio | 1 |
| `examples/mcp-go-sampling/` | mark3labs/mcp-go sampling_server | Go | stdio | 3 |
| `examples/mcp-go-elicitation/` | mark3labs/mcp-go elicitation | Go | stdio | 4 |
| `examples/mcp-go-everything-completion/` | mark3labs/mcp-go everything | Go | stdio | 3 |
| `examples/mcp-go-everything-logging/` | mark3labs/mcp-go everything | Go | stdio | 2 |
| `examples/github-mcp/` | github/github-mcp-server | Go | stdio | 20 |
| `examples/trajectory/` | Inline trace (no server) | N/A | N/A | 20 |

---

## Filesystem server. TypeScript

**Directory:** `examples/filesystem/`

Tests the official `@modelcontextprotocol/server-filesystem`. 14 assertions: read file, read multiple files, read text file, list directory, list directory with sizes, directory tree, get file info, search files, write file, edit file, create directory, move file, list allowed directories, and a **negative test** that verifies path traversal is rejected. 92% tool coverage (13/14 tools; `read_media_file` excluded due to [upstream spec violation](https://github.com/modelcontextprotocol/servers/issues/4029)).

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

The assertions reference the server at `/tmp/fastmcp/examples/testing_demo/server.py`. Clone the repo before running:

```bash
git clone --depth 1 https://github.com/PrefectHQ/fastmcp.git /tmp/fastmcp
mcp-assert run --suite examples/fastmcp-testing-demo
```

You can also use `--server` to override the path if you cloned to a different location.

## agent-lsp. Go

**Directory:** `examples/agent-lsp-go/`

Tests [agent-lsp](https://github.com/blackwell-systems/agent-lsp) with gopls. 63 assertions covering all 50 tools: navigation, refactoring, analysis, session lifecycle, workspace, and build. 100% tool coverage.

```bash
git clone --depth 1 https://github.com/blackwell-systems/agent-lsp.git /tmp/agent-lsp
mcp-assert run --suite examples/agent-lsp-go --fixture /tmp/agent-lsp/test/fixtures/go
```

## mcp-go everything. Go (stdio)

**Directory:** `examples/mcp-go-everything/`

Tests the `mark3labs/mcp-go` SDK's `examples/everything` server over stdio. 9 assertions: echo, add, image content, resource link, notification, and long-running operation (skipped due to known [stdio transport bug](https://github.com/mark3labs/mcp-go/issues/826)). 100% tool coverage for non-buggy tools.

```bash
# Build the everything server from the mcp-go repo
mcp-assert run --suite examples/mcp-go-everything
```

## mcp-go everything. Go (HTTP)

**Directory:** `examples/mcp-go-everything-http/`

Same tools as the stdio suite, tested over streamable HTTP transport. 5 assertions. Demonstrates transport-agnostic testing: the same tool expectations work over both stdio and HTTP without changing the `expect:` blocks.

```bash
# Start the everything server with HTTP transport, then:
mcp-assert run --suite examples/mcp-go-everything-http
```

## mcp-go everything prompts. Go

**Directory:** `examples/mcp-go-everything-prompts/`

Tests `prompts/list` and `prompts/get` on the mcp-go everything server. 4 assertions: list prompts, get static prompt, get template prompt with arguments, and a pagination pattern example using `json_path` on cursor values.

```bash
mcp-assert run --suite examples/mcp-go-everything-prompts
```

## mcp-go everything resources. Go

**Directory:** `examples/mcp-go-everything-resources/`

Tests `resources/list`, `resources/read`, `resources/subscribe`, and `resources/unsubscribe` on the mcp-go everything server. 4 assertions: list resources (verifies `test://static/resource` is exposed), read a specific resource by URI, subscribe to a resource, and subscribe then unsubscribe.

```bash
mcp-assert run --suite examples/mcp-go-everything-resources
```

## mcp-go typed_tools. Go

**Directory:** `examples/mcp-go-typed-tools/`

Tests the `mark3labs/mcp-go` `typed_tools` example server. 3 assertions: greeting with required parameters, greeting with optional parameters, and error handling for missing required input. 100% tool coverage.

```bash
mcp-assert run --suite examples/mcp-go-typed-tools
```

## mcp-go structured. Go

**Directory:** `examples/mcp-go-structured/`

Tests the `mark3labs/mcp-go` `structured_input_and_output` example server. 6 assertions: weather data, user profile, image assets, and manual structured results. Demonstrates `json_path` assertions on structured response objects. 100% tool coverage.

```bash
mcp-assert run --suite examples/mcp-go-structured
```

## mcp-go roots. Go

**Directory:** `examples/mcp-go-roots/`

Tests the `mark3labs/mcp-go` `roots_server` example. 1 assertion: the `roots` tool calls `roots/list` back to the client, and mcp-assert responds with the configured workspace roots via `client_capabilities.roots`. 100% tool coverage.

```bash
mcp-assert run --suite examples/mcp-go-roots --fixture examples/mcp-go-roots/fixtures
```

## mcp-go sampling. Go

**Directory:** `examples/mcp-go-sampling/`

Tests the `mark3labs/mcp-go` `sampling_server` example. 3 assertions: `ask_llm` with a custom system prompt, `ask_llm` without a system prompt, and `greet` (verifying non-sampling tools work normally when `client_capabilities.sampling` is set). mcp-assert responds to `sampling/createMessage` requests with a mock LLM response. 100% tool coverage.

```bash
mcp-assert run --suite examples/mcp-go-sampling
```

## mcp-go elicitation. Go

**Directory:** `examples/mcp-go-elicitation/`

Tests the `mark3labs/mcp-go` elicitation example server. 4 assertions: `create_project`, `cancel_flow`, `decline_flow`, and `validation_constraints`. Each triggers an `elicitation/create` request, and mcp-assert responds with preset field values via `client_capabilities.elicitation`.

```bash
mcp-assert run --suite examples/mcp-go-elicitation
```

## mcp-go everything completion. Go

**Directory:** `examples/mcp-go-everything-completion/`

Tests `completion/complete` on the mcp-go everything server. 3 assertions: completion for a prompt argument, completion for a resource URI, and completion with an empty prefix.

```bash
mcp-assert run --suite examples/mcp-go-everything-completion
```

## mcp-go everything logging. Go

**Directory:** `examples/mcp-go-everything-logging/`

Tests `logging/setLevel` on the mcp-go everything server. 2 assertions: setting the log level and capturing log messages after a tool call.

```bash
mcp-assert run --suite examples/mcp-go-everything-logging
```


## GitHub MCP Server. Go

**Directory:** `examples/github-mcp/`

Tests [github/github-mcp-server](https://github.com/github/github-mcp-server) (28K+ stars), the most popular MCP server. 20 assertions targeting 17 read-only tools across 7 toolsets: context (`get_me`), repos (`get_file_contents`, `search_repositories`, `search_code`, `list_branches`, `list_commits`, `list_tags`, `list_releases`, `get_latest_release`, `get_release_by_tag`), git (`get_repository_tree`), issues (`list_issues`, `search_issues`), pull requests (`list_pull_requests`, `search_pull_requests`), users (`search_users`), gists (`list_gists`). Requires a `GITHUB_PERSONAL_ACCESS_TOKEN` environment variable.

The assertions reference `github-mcp-server` in PATH. Build and install the server before running:

```bash
git clone --depth 1 https://github.com/github/github-mcp-server.git /tmp/github-mcp-server
cd /tmp/github-mcp-server && go build -o github-mcp-server ./cmd/github-mcp-server

# Add to PATH or use --server to override:
mcp-assert run --suite examples/github-mcp --server "/tmp/github-mcp-server/github-mcp-server stdio"
```

Or if `github-mcp-server` is already in your PATH:

```bash
# Create a token at https://github.com/settings/tokens with repo + read:user scopes
GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_TOKEN mcp-assert run --suite examples/github-mcp
```

## Trajectory assertions (no server)

**Directory:** `examples/trajectory/`

20 trajectory assertions covering all 20 agent-lsp skill protocols. These assertions validate tool call sequences using inline traces (no live server required). Each assertion captures the required tool call sequence, safety gates, and absence checks for one skill:

| Skill | Key constraints |
|-------|----------------|
| `/lsp-rename` | `prepare_rename` before `rename_symbol`; no `apply_edit` |
| `/lsp-safe-edit` | `simulate_edit_atomic` before `apply_edit` |
| `/lsp-refactor` | blast-radius check before any edit |
| `/lsp-impact` | `get_references` + `call_hierarchy` + `type_hierarchy` |
| `/lsp-simulate` | no `apply_edit` (simulate only) |
| `/lsp-verify` | diagnostics + build + tests |
| (and 14 more) | each skill has a dedicated `trajectory_<skill>_protocol.yaml` |

```bash
mcp-assert run --suite examples/trajectory
```

Trajectory assertions use the `trace:` field (inline YAML) instead of `server:`. To validate real agent behavior, replace `trace:` with `audit_log: path/to/agent.jsonl` pointing at an agent's recorded tool call log.

For the trajectory YAML format, see [Writing Assertions](writing-assertions.md#trajectory-assertions).
