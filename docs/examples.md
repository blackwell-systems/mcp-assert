# Examples

mcp-assert ships with 61 suites for 55 MCP servers in 7 languages (Go, TypeScript, Python, Rust, Kotlin/Java, Swift, C#), plus a trajectory suite that runs without a server. 3 transports: stdio, SSE, HTTP. 570 total assertions.

## Summary

| Suite | Server | Language | Transport | Assertions |
|-------|--------|----------|-----------|------------|
| `examples/filesystem/` | `@modelcontextprotocol/server-filesystem` | TypeScript | stdio | 14 |
| `examples/memory/` | `@modelcontextprotocol/server-memory` | TypeScript | stdio | 9 |
| `examples/mcp-time/` | `mcp-server-time` | Python | stdio | 5 |
| `examples/mcp-fetch/` | `mcp-server-fetch` | Python | stdio | 3 |
| `examples/mcp-git/` | `mcp-server-git` | Python | stdio | 11 |
| `examples/sqlite/` | `mcp-server-sqlite` | Python | stdio | 9 |
| `examples/mcp-everything-ts/` | `@modelcontextprotocol/server-everything` | TypeScript | stdio | 13 |
| `examples/fastmcp-testing-demo/` | PrefectHQ/fastmcp testing_demo | Python | stdio | 16 |
| `examples/fastmcp-testing-demo-sse/` | PrefectHQ/fastmcp testing_demo | Python | SSE | 11 |
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
| `examples/rmcp-counter/` | 4t145/rmcp counter | Rust | stdio | 14 |
| `examples/rust-filesystem/` | rust-mcp-stack/rust-mcp-filesystem | Rust | stdio | 23 |
| `examples/excel-mcp/` | haris-musa/excel-mcp-server | Python | stdio | 15 |
| `examples/antvis-chart/` | antvis/mcp-server-chart | TypeScript | stdio | 25 |
| `examples/notion-mcp/` | makenotion/notion-mcp-server | TypeScript | stdio | 22 |
| `examples/terraform-mcp/` | hashicorp/terraform-mcp-server | Go | stdio | 5 |
| `examples/mongodb-mcp/` | mongodb/mongodb-mcp-server | TypeScript | stdio | 4 |
| `examples/spring-mcp/` | jamesward/hello-spring-mcp-server | Kotlin | HTTP | 3 |
| `examples/playwright-mcp/` | microsoft/playwright-mcp | TypeScript | stdio | 14 |
| `examples/openai-deep-research/` | openai/sample-deep-research-mcp | Python | stdio | 4 |
| `examples/google-storage-mcp/` | @google-cloud/storage-mcp | TypeScript | stdio | 6 |
| `examples/grafana-mcp/` | grafana/mcp-grafana | Go | stdio | 54 |
| `examples/arxiv-mcp/` | blazickjp/arxiv-mcp-server | Python | stdio | 5 |
| `examples/aws-docs-mcp/` | awslabs/aws-documentation-mcp-server | Python | stdio | 4 |
| `examples/exa-mcp/` | exa-labs/exa-mcp-server | JavaScript | stdio | 2 |
| `examples/git-mcp-idosal/` | onmyway133/git-mcp | TypeScript | stdio | 14 |
| `examples/perplexity-mcp/` | perplexityai/mcp-server | TypeScript | stdio | 4 |
| `examples/engram/` | Gentleman-Programming/engram | Go | stdio | 16 |
| `examples/codegraph-context/` | nicobailey/codegraph-context-mcp | TypeScript | stdio | 16 |
| `examples/deep-research/` | u14app/deep-research | JavaScript | HTTP | 5 |
| `examples/peekaboo/` | steipete/Peekaboo | Swift | stdio | 6 |
| `examples/puppeteer-mcp/` | @modelcontextprotocol/server-puppeteer | TypeScript | stdio | 7 |
| `examples/chrome-devtools-mcp/` | chrome-devtools-mcp | TypeScript | stdio | 7 |
| `examples/firefox-devtools-mcp/` | mozilla/firefox-devtools-mcp | TypeScript | stdio | 7 |
| `examples/context7-mcp/` | @upstash/context7-mcp | TypeScript | stdio | 2 |
| `examples/csharp-weather/` | modelcontextprotocol/csharp-sdk | C# | stdio | 2 |
| `examples/duckduckgo-mcp/` | duckduckgo-mcp-server | Python | stdio | 2 |
| `examples/excalidraw-architect-mcp/` | excalidraw-architect-mcp | Python | stdio | 4 |
| `examples/kubernetes-mcp/` | mcp-server-kubernetes | Python | stdio | 2 |
| `examples/lighthouse-mcp/` | lighthouse-mcp-server | TypeScript | stdio | 2 |
| `examples/markitdown-mcp/` | markitdown-mcp | Python | stdio | 1 |
| `examples/mcp-devtools/` | sammcj/mcp-devtools | Go | stdio | 5 |
| `examples/mcp-math/` | mcp-server-math | Python | stdio | 4 |
| `examples/mobile-mcp/` | mobile-next/mobile-mcp | TypeScript | stdio | 6 |
| `examples/sec-edgar-mcp/` | sec-edgar-mcp | Python | stdio | 5 |
| `examples/spec-workflow-mcp/` | Pimzino/spec-workflow-mcp | TypeScript | stdio | 1 |
| `examples/xcodebuild-mcp/` | getsentry/XcodeBuildMCP | TypeScript | stdio | 10 |
| `examples/yfinance-mcp/` | narumiruna/yfinance-mcp | Python | stdio | 4 |
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

Tests the official `@modelcontextprotocol/server-memory`. 9 assertions with **stateful setup** and 100% tool coverage (9/9 tools): create entities, add observations, create relations, search nodes, open nodes, delete entities, delete observations, delete relations, and verify empty search returns nothing.

```bash
npm install -g @modelcontextprotocol/server-memory
mcp-assert run --suite examples/memory
```

## SQLite server. Python

**Directory:** `examples/sqlite/`

Tests the official `mcp-server-sqlite` (Python). 9 assertions with 100% tool coverage (6/6 tools): list tables, SELECT queries, COUNT, JOINs, describe table schema, CREATE TABLE, INSERT, write query, and error handling for invalid SQL. Fixture is a pre-built `.db` file.

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

## Anthropic time server. Python

**Directory:** `examples/mcp-time/`

Tests the official `mcp-server-time`. 5 assertions: UTC time, named timezone, time conversion, invalid timezone rejection, and named timezone validation. 100% tool coverage (2/2 tools).

```bash
mcp-assert run --suite examples/mcp-time
```

## Anthropic fetch server. Python

**Directory:** `examples/mcp-fetch/`

Tests the official `mcp-server-fetch`. 3 assertions: URL fetch, invalid URL rejection, and unreachable host handling. 100% tool coverage (1/1 tool).

```bash
mcp-assert run --suite examples/mcp-fetch
```

## Anthropic git server. Python

**Directory:** `examples/mcp-git/`

Tests the official `mcp-server-git`. 11 assertions: status, log, branch, diff, show, commit, add, reset, tag, invalid repo rejection, and invalid ref handling. 92% tool coverage (11/12 tools).

```bash
mcp-assert run --suite examples/mcp-git
```

## Anthropic everything server. TypeScript

**Directory:** `examples/mcp-everything-ts/`

Tests the official `@modelcontextprotocol/server-everything` (Anthropic reference server). 13 assertions: echo, sum, image, resource links, structured content, annotations, env, gzip, long-running operation. 92% tool coverage (12/13 tools, 1 skipped).

```bash
mcp-assert run --suite examples/mcp-everything-ts
```

## fastmcp testing_demo over SSE. Python

**Directory:** `examples/fastmcp-testing-demo-sse/`

Same server as the stdio fastmcp suite, tested over SSE transport. 11 assertions. First SSE transport coverage in the suite collection.

```bash
mcp-assert run --suite examples/fastmcp-testing-demo-sse
```

## rmcp counter. Rust

**Directory:** `examples/rmcp-counter/`

Tests the `4t145/rmcp` SDK's counter example server. 14 assertions: increment, decrement, get_value, sum, echo, say_hello, plus resources and prompts. 100% tool coverage (6/6 tools + resources + prompts). Bug found: `get_value` decrements counter instead of reading it (repo archived).

```bash
mcp-assert run --suite examples/rmcp-counter
```

## rust-mcp-filesystem. Rust

**Directory:** `examples/rust-filesystem/`

Tests `rust-mcp-stack/rust-mcp-filesystem` (145 stars). 23 assertions: read, list, search, write, edit, zip/unzip, head, tail, line ranges, path traversal rejection, duplicate file detection, empty directory detection. 92% tool coverage (22/24 tools). Clean scan.

```bash
mcp-assert run --suite examples/rust-filesystem
```

## excel-mcp-server. Python

**Directory:** `examples/excel-mcp/`

Tests `haris-musa/excel-mcp-server` (3,750 stars). 15 assertions: workbook creation, data round-trip, formulas, charts, pivot tables, formatting, merge cells, validation. 52% tool coverage (13/25 tools). Clean scan.

```bash
mcp-assert run --suite examples/excel-mcp
```

## antvis chart server. TypeScript

**Directory:** `examples/antvis-chart/`

Tests `antvis/mcp-server-chart` (4K stars). 25 assertions covering 25 chart types. 9 bugs found: unhandled JavaScript exceptions on default/minimal input. Filed [antvis/mcp-server-chart#291](https://github.com/antvis/mcp-server-chart/issues/291).

```bash
mcp-assert run --suite examples/antvis-chart
```

## Notion MCP server. TypeScript

**Directory:** `examples/notion-mcp/`

Tests the official `makenotion/notion-mcp-server` (4.2K stars). 22 assertions, 100% tool coverage (22/22 tools). Clean scan.

```bash
mcp-assert run --suite examples/notion-mcp
```

## Terraform MCP server. Go

**Directory:** `examples/terraform-mcp/`

Tests `hashicorp/terraform-mcp-server` (1.3K stars). 5 assertions: provider lookup, module search, policy search. 56% tool coverage (5/9 tools). Clean scan.

```bash
mcp-assert run --suite examples/terraform-mcp
```

## MongoDB MCP server. TypeScript

**Directory:** `examples/mongodb-mcp/`

Tests the official `mongodb/mongodb-mcp-server` (1K stars). 4 assertions: knowledge search, error handling. Clean scan.

```bash
mcp-assert run --suite examples/mongodb-mcp
```

## Spring AI MCP server. Kotlin

**Directory:** `examples/spring-mcp/`

Tests `jamesward/hello-spring-mcp-server`. 3 assertions, 100% tool coverage (2/2 tools). First JVM language in the suite collection. Uses HTTP transport. Clean scan.

```bash
mcp-assert run --suite examples/spring-mcp
```

## Playwright MCP server. TypeScript

**Directory:** `examples/playwright-mcp/`

Tests `microsoft/playwright-mcp` (31K stars). 14 assertions: navigate, snapshot, screenshot, JS evaluate, console messages, network requests, resize, close, tabs, navigate back, press key, wait for element, invalid URL rejection, empty page handling. 67% tool coverage (14/21 tools). Clean scan.

```bash
mcp-assert run --suite examples/playwright-mcp
```

## OpenAI deep research MCP. Python

**Directory:** `examples/openai-deep-research/`

Tests `openai/sample-deep-research-mcp`. 4 assertions: search and fetch against static JSON dataset. 100% tool coverage (2/2 tools). Clean scan.

```bash
mcp-assert run --suite examples/openai-deep-research
```

## Google Cloud Storage MCP. TypeScript

**Directory:** `examples/google-storage-mcp/`

Tests `@google-cloud/storage-mcp`. 6 assertions: bucket metadata, object listing, IAM policy, input validation. 35% tool coverage (6/17 tools, no GCP credentials required). Clean scan.

```bash
mcp-assert run --suite examples/google-storage-mcp
```

## Grafana MCP server. Go

**Directory:** `examples/grafana-mcp/`

Tests `grafana/mcp-grafana`. 17 assertions (14 no-credentials + 3 live-backend via `skip_unless_env`). 1 bug found: `get_assertions` returns internal error (-32603) instead of `isError:true` on invalid timestamps. Filed [grafana/mcp-grafana#792](https://github.com/grafana/mcp-grafana/issues/792).

```bash
mcp-assert run --suite examples/grafana-mcp
```

## arxiv MCP server. Python

**Directory:** `examples/arxiv-mcp/`

Tests `blazickjp/arxiv-mcp-server`. 5 assertions. 1 bug found: `get_abstract` returns error content without `isError` flag. Filed [blazickjp/arxiv-mcp-server#92](https://github.com/blazickjp/arxiv-mcp-server/issues/92).

```bash
mcp-assert run --suite examples/arxiv-mcp
```

## AWS documentation MCP server. Python

**Directory:** `examples/aws-docs-mcp/`

Tests `awslabs/aws-documentation-mcp-server`. 4 assertions: search, recommend, no-results handling. 100% tool coverage (4/4 tools). Clean scan.

```bash
mcp-assert run --suite examples/aws-docs-mcp
```

## Exa search MCP server. JavaScript

**Directory:** `examples/exa-mcp/`

Tests `exa-labs/exa-mcp-server`. 2 assertions: proper 401 with `isError: true` and API key guidance when credentials missing. 100% tool coverage (2/2 tools). Clean scan.

```bash
mcp-assert run --suite examples/exa-mcp
```

## git-mcp. TypeScript

**Directory:** `examples/git-mcp-idosal/`

Tests `onmyway133/git-mcp`. 14 assertions: status, log, branches, diff, show, reflog, stash, tag, blame, grep, cherry-pick, remote, invalid repo rejection. 39% tool coverage (14/36 tools). Clean.

```bash
mcp-assert run --suite examples/git-mcp-idosal
```

## Perplexity MCP server. TypeScript

**Directory:** `examples/perplexity-mcp/`

Tests `perplexityai/mcp-server`. 4 assertions, 100% tool coverage (4/4 tools). All tools return `isError:true` with 401 and API key guidance when credentials are invalid. Clean scan.

```bash
mcp-assert run --suite examples/perplexity-mcp
```

## engram memory server. Go

**Directory:** `examples/engram/`

Tests `Gentleman-Programming/engram`. 16 assertions, 100% tool coverage (16/16 tools). Full coverage including writes (save, delete, update, merge, session lifecycle). SQLite state is self-contained. Clean scan.

```bash
mcp-assert run --suite examples/engram
```

## CodeGraphContext MCP server. TypeScript

**Directory:** `examples/codegraph-context/`

Tests `nicobailey/codegraph-context-mcp`. 16 assertions. Code graph indexer with 21 tools. Clean scan.

```bash
mcp-assert run --suite examples/codegraph-context
```

## u14app deep-research. JavaScript (HTTP)

**Directory:** `examples/deep-research/`

Tests `u14app/deep-research` via HTTP transport. 5 assertions. All tools return `isError:true` with "Unsupported Provider" when no LLM credentials configured. Clean scan.

```bash
mcp-assert run --suite examples/deep-research
```

## Peekaboo. Swift

**Directory:** `examples/peekaboo/`

Tests `steipete/Peekaboo`. 6 assertions. First Swift MCP server in the suite collection. 1 bug found: `image` returns internal error (-32603) instead of `isError:true` when Screen Recording permission is not granted. Filed [steipete/Peekaboo#108](https://github.com/steipete/Peekaboo/issues/108).

```bash
mcp-assert run --suite examples/peekaboo
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
