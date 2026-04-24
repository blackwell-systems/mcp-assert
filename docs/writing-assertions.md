# Writing Assertions

Every assertion YAML file has four parts: **name**, **server**, **assert**, and optionally **setup**.

## Assertion File Format

```yaml
# Required: shown in test output and reports.
name: Human-readable description

# Required: how to start the MCP server. mcp-assert launches this process,
# connects over stdio, and handles the MCP initialize handshake.
server:
  command: path/to/mcp-server        # Binary or command to run
  args: ["arg1", "arg2"]             # CLI arguments
  env:                               # Optional environment variables
    KEY: value

# Optional: tool calls that run before the assertion, in order.
# Use for creating state the assertion depends on.
setup:
  - tool: setup_tool
    args: { key: value }
  - tool: another_setup_tool
    args: { key: value }

# Required: the tool call to test and its expected results.
assert:
  tool: tool_under_test              # MCP tool name
  args: { key: value }              # Arguments passed to the tool
  expect:                            # All checks must pass
    not_error: true
    contains: ["expected", "strings"]
    not_contains: ["unexpected"]
    matches_regex: ["\\d+ items"]
    json_path:
      "$.locations[0].file": "main.go"
    min_results: 3

# Optional: per-assertion timeout (default: 30s).
# The server process is killed if it doesn't respond in time.
timeout: 30s
```

`{{fixture}}` in any argument is replaced with the `--fixture` directory at runtime. Substitution works recursively in strings, arrays, and nested maps.

## Environment variable expansion

Environment variables in `env:` blocks support shell-style expansion. Use `${VAR}` or `$VAR` syntax to reference variables from the host environment:

```yaml
server:
  command: my-mcp-server
  env:
    DATABASE_URL: "postgres://${DB_USER}:${DB_PASS}@localhost/testdb"
    API_KEY: "$MY_API_KEY"
```

If the referenced variable is not set in the host environment, the original string is preserved unchanged (e.g., `${DB_USER}` remains literal). This lets you share assertion files across environments without hardcoding secrets.

## Minimal assertion

The simplest assertion calls one tool and checks the result:

```yaml
name: list_tables returns all tables        # Shown in test output
server:
  command: uvx                              # How to start the MCP server
  args: ["mcp-server-sqlite", "--db-path", "{{fixture}}/test.db"]
assert:
  tool: list_tables                         # MCP tool to call
  args: {}                                  # Arguments (empty object if none)
  expect:
    not_error: true                         # Tool did not return an error
    contains: ["users", "projects"]         # Response includes these strings
timeout: 15s                                # Kill server if no response
```

## Setup steps (stateful tests)

Some tools need state to exist first. Use `setup` to run tool calls before the assertion:

```yaml
name: search finds created entity
server:
  command: npx
  args: ["@modelcontextprotocol/server-memory"]
setup:
  - tool: create_entities                   # Runs first
    args:
      entities:
        - name: "Alice"
          entityType: "person"
          observations: ["engineer"]
  - tool: add_observations                  # Runs second
    args:
      observations:
        - entityName: "Alice"
          contents: ["promoted to staff"]
assert:
  tool: search_nodes                        # Runs last: this is what we're testing
  args:
    query: "Alice"
  expect:
    not_error: true
    contains: ["Alice", "promoted to staff"]
```

Setup steps run in order. If any step fails, the whole assertion fails. Each assertion starts a fresh server process, so setup from one assertion never leaks into another.

## Fixture substitution

`{{fixture}}` in any argument value is replaced with the `--fixture` directory at runtime. This works in strings, arrays, and nested objects:

```yaml
assert:
  tool: read_multiple_files
  args:
    paths:                                  # {{fixture}} works inside arrays too
      - "{{fixture}}/file1.txt"
      - "{{fixture}}/file2.txt"
```

Use fixtures for test data your server needs: sample files, databases, config. Keep them in version control alongside your assertions.

**Fixture isolation:** Each assertion automatically receives its own copy of the fixture directory. The original fixture is never modified, even if the assertion writes files, applies edits, or commits changes. This prevents one assertion's side effects from shifting line numbers or altering state for subsequent assertions. Docker mode already isolates via fresh containers, so the copy is skipped when `--docker` is used.

## Negative tests (expecting errors)

Test that your server rejects bad input correctly:

```yaml
name: rejects path traversal
assert:
  tool: read_file
  args:
    path: "/etc/passwd"                     # Outside allowed directory
  expect:
    is_error: true                          # The tool SHOULD return an error
```

## Chaining outputs between steps (capture)

Some workflows need output from one step as input to the next. Use `capture` to extract values via jsonpath:

```yaml
name: session lifecycle: create, edit, evaluate
setup:
  - tool: create_simulation_session
    args:
      workspace_root: "{{fixture}}"
      language: go
    capture:
      session_id: "$.session_id"          # extract from JSON response
  - tool: simulate_edit
    args:
      session_id: "{{session_id}}"        # inject captured value
      file_path: "{{fixture}}/main.go"
      start_line: 13
      new_text: "return 42"
assert:
  tool: evaluate_session
  args:
    session_id: "{{session_id}}"          # same captured value
  expect:
    not_error: true
    contains: ["net_delta"]
```

Captured variables work anywhere `{{fixture}}` works: strings, arrays, nested objects. Use this for session IDs, auth tokens, created resource IDs, or any value returned by a setup step.

## Client Capabilities (Bidirectional MCP)

MCP is bidirectional: servers can request things from the client. Use `client_capabilities` in the server block to make mcp-assert respond to these server-initiated requests.

### Roots

Some servers call `roots/list` to discover the client's workspace. Provide the paths to return:

```yaml
name: roots tool returns client workspace roots
server:
  command: /path/to/roots-server
  client_capabilities:
    roots:
      - "{{fixture}}"
assert:
  tool: roots
  args: {}
  expect:
    not_error: true
    contains: ["Root list"]
timeout: 15s
```

`{{fixture}}` in roots paths is substituted at runtime, just like in args. Each path becomes a `file://` URI with the directory basename as the root name.

### Sampling

Servers that use `sampling/createMessage` to call an LLM via the client. Provide a mock response:

```yaml
name: ask_llm returns mock sampling response
server:
  command: /path/to/sampling-server
  client_capabilities:
    sampling:
      text: "The capital of France is Paris."
      model: mock-gpt          # Optional; defaults to "mock"
      stop_reason: end_turn    # Optional; defaults to "end_turn"
assert:
  tool: ask_llm
  args:
    question: "What is the capital of France?"
  expect:
    not_error: true
    contains: ["Paris"]
timeout: 15s
```

The server receives the mock response and incorporates it into its tool output. The assertion checks the final tool result, not the sampling exchange directly.

### Elicitation

Servers that use `elicitation/create` to prompt the user for structured input. Provide preset values:

```yaml
name: create_project accepts elicitation response
server:
  command: /path/to/elicitation-server
  client_capabilities:
    elicitation:
      content:
        projectName: "myapp"
        framework: "react"
        includeTests: true
assert:
  tool: create_project
  args: {}
  expect:
    not_error: true
    contains: ["myapp"]
timeout: 15s
```

mcp-assert responds to the elicitation request with `action: accept` and the provided content. If the `content` key is omitted, the entire elicitation map is used as the response content.

### Combining capabilities

All three capabilities can be set together:

```yaml
server:
  command: /path/to/server
  client_capabilities:
    roots:
      - "{{fixture}}/workspace"
    sampling:
      text: "Mock LLM answer"
    elicitation:
      content:
        confirmed: true
```

`client_capabilities` is a YAML-level feature. There is no CLI flag equivalent, and it applies to stdio transport only.

## Prompt Assertions

MCP servers can expose prompt templates via `prompts/list` and `prompts/get`. Use `assert_prompts:` instead of `assert:` to test them.

### List prompts

```yaml
name: server exposes expected prompt templates
server:
  command: /path/to/mcp-server
assert_prompts:
  list: {}
  expect:
    not_error: true
    not_empty: true
    contains: ["code_review", "summarize"]
timeout: 15s
```

The `list: {}` value calls `prompts/list` and marshals the result as JSON. All `expect` assertions work on the JSON response, including `json_path`:

```yaml
assert_prompts:
  list: {}
  expect:
    json_path:
      "$.prompts[0].name": "code_review"
```

### Get a prompt

Use `get:` to call `prompts/get` and retrieve a rendered prompt. The response text is built from the prompt's description and message content:

```yaml
name: simple prompt returns expected content
server:
  command: /path/to/mcp-server
assert_prompts:
  get:
    name: "simple_prompt"
  expect:
    not_error: true
    contains: ["simple prompt without arguments"]
timeout: 15s
```

### Get a template prompt with arguments

Template prompts require arguments. Pass them as a string map:

```yaml
name: code review prompt accepts language argument
server:
  command: /path/to/mcp-server
assert_prompts:
  get:
    name: "code_review"
    arguments:
      language: "go"
      style: "concise"
  expect:
    not_error: true
    contains: ["language=go", "style=concise"]
timeout: 15s
```

`{{fixture}}` and captured variables (from `setup:` steps) are substituted in `name` and all `arguments` values.

### Pagination

`prompts/list`, `resources/list`, and `tools/list` support cursor-based pagination. Assert on cursor values via `json_path`:

```yaml
assert_prompts:
  list: {}
  expect:
    json_path:
      "$.nextCursor": "page2-cursor"
```

To request a specific page, pass the cursor in the `list:` block:

```yaml
assert_prompts:
  list:
    cursor: "page2-cursor"
  expect:
    contains: ["page2-prompt"]
```

The same pattern applies to `assert_resources: list:` and works with `json_path` on any list response.

## Progress Notifications

Some tools send `notifications/progress` during execution. Use `capture_progress: true` on the `assert:` block to collect them, then assert the count with `min_progress:`:

```yaml
name: long operation sends progress updates
server:
  command: /path/to/mcp-server
assert:
  tool: long_running_operation
  args:
    duration: 5
    steps: 3
  capture_progress: true
  expect:
    not_error: true
    min_progress: 3
timeout: 30s
```

`capture_progress` registers a notification handler before the tool call. After the tool returns, `min_progress` asserts the handler received at least that many `notifications/progress` messages.

If `capture_progress: true` is set but `min_progress` is absent, progress notifications are collected but not checked — useful for ensuring the feature doesn't break existing assertions.

Note: progress capture requires the server to properly send `notifications/progress` notifications. The mcp-go `everything` server's `longRunningOperation` has a known stdio transport bug (`fmt.Printf` to stdout corrupts the JSON-RPC stream) that prevents reliable testing with that server. Test against servers that send progress correctly.

## Skipping assertions

Add `skip: true` at the top level of an assertion YAML to exclude it from execution:

```yaml
name: write tool that modifies external state
skip: true
server:
  command: my-mcp-server
assert:
  tool: delete_record
  args:
    id: "123"
  expect:
    not_error: true
```

Skipped assertions are reported as `SKIP` and do not affect pass/fail counts. Use this for destructive tools, tests that depend on external services, or temporarily flaky assertions.

The `generate` command sets `skip: true` automatically on tools detected as destructive. Remove the field (or set it to `false`) when the assertion is ready to run.

## HTTP/SSE Transport

By default, mcp-assert connects to servers over stdio (launching a subprocess). For servers that run over HTTP, set the `transport` and `url` fields:

### SSE transport (legacy)

```yaml
name: echo over SSE
server:
  transport: sse
  url: "http://localhost:8080/sse"
assert:
  tool: echo
  args:
    message: "hello"
  expect:
    not_error: true
    contains: ["hello"]
```

### Streamable HTTP transport

```yaml
name: echo over HTTP
server:
  transport: http
  url: "http://localhost:8080/mcp"
assert:
  tool: echo
  args:
    message: "hello"
  expect:
    not_error: true
    contains: ["hello"]
```

Transport values are case-insensitive. When `transport` is omitted or set to `stdio`, the `command`/`args`/`env` fields are used to launch a subprocess. When `transport` is `sse` or `http`, the `url` field is required and `command`/`args` are ignored.

Docker isolation (`--docker`) is only supported with stdio transport.

## Server environment variables

Pass environment variables to the server process:

```yaml
server:
  command: my-mcp-server
  args: ["--port", "0"]
  env:
    DATABASE_URL: "sqlite:///tmp/test.db"
    LOG_LEVEL: "debug"
```

## Resource Assertions

MCP servers can expose resources via `resources/list` and `resources/read`. Use `assert_resources:` instead of `assert:` to test them.

### List resources

```yaml
name: server exposes expected resources
server:
  command: /path/to/mcp-server
assert_resources:
  list: {}
  expect:
    not_error: true
    not_empty: true
    contains: ["test://static/resource"]
timeout: 15s
```

The `list: {}` value calls `resources/list` and marshals the result as JSON. All `expect` assertions work on the JSON response, including `json_path`. Pass `cursor:` in the `list:` block to request a specific page:

```yaml
assert_resources:
  list:
    cursor: "page2-cursor"
  expect:
    contains: ["page2-resource"]
```

### Read a resource

Use `read:` with the resource URI to call `resources/read` and assert on the content:

```yaml
name: static resource returns expected content
server:
  command: /path/to/mcp-server
assert_resources:
  read: "test://static/resource"
  expect:
    not_error: true
    not_empty: true
    contains: ["expected content"]
timeout: 15s
```

`{{fixture}}` and captured variables are substituted in the URI string.

### Parameterized (template) resources

Template resources have URIs with variable segments. Pass the filled-in URI to `read:`:

```yaml
assert_resources:
  read: "demo://greeting/Alice"
  expect:
    not_error: true
    contains: ["Alice"]
```

## Trajectory Assertions

Trajectory assertions validate that an agent (or a recorded tool call sequence) calls MCP tools in the correct order, with the correct arguments, without calling forbidden tools. They run without a live server.

### Inline trace

Use `trace:` to define the sequence inline in YAML:

```yaml
name: lsp-rename follows skill protocol
trace:
  - tool: prepare_rename
    args:
      file_path: "fixtures/go/main.go"
      line: 6
      column: 6
  - tool: rename_symbol
    args:
      file_path: "fixtures/go/main.go"
      line: 6
      column: 6
      new_name: "Entity"
      dry_run: false
  - tool: get_diagnostics
    args:
      file_path: "fixtures/go/main.go"
trajectory:
  - type: order
    tools: ["prepare_rename", "rename_symbol", "get_diagnostics"]
  - type: presence
    tools: ["prepare_rename", "rename_symbol", "get_diagnostics"]
  - type: absence
    tools: ["apply_edit"]
  - type: args_contain
    tool: rename_symbol
    args:
      new_name: "Entity"
```

### Audit log source

Replace `trace:` with `audit_log:` to validate real agent behavior from a recorded JSONL log:

```yaml
name: lsp-rename follows skill protocol (live agent)
audit_log: /path/to/agent-audit.jsonl
trajectory:
  - type: order
    tools: ["prepare_rename", "rename_symbol", "get_diagnostics"]
  - type: presence
    tools: ["prepare_rename"]
  - type: absence
    tools: ["apply_edit"]
```

The JSONL file should contain one JSON object per line with a `tool` field (and optionally an `args` object) for each tool call the agent made.

### Trajectory assertion types

| Type | What it checks | Required fields |
|------|----------------|----------------|
| `order` | Listed tools appear in this sequence (not necessarily adjacent) | `tools: [...]` |
| `presence` | All listed tools appear at least once | `tools: [...]` |
| `absence` | None of the listed tools appear | `tools: [...]` |
| `args_contain` | A specific tool was called with these argument values (partial match) | `tool: "..."`, `args: {...}` |

Multiple trajectory assertions can be combined in a single YAML file. All must pass for the assertion to pass.

Trajectory assertions run in 0ms (no server startup, no MCP protocol). They are designed for validating skill protocols: the required sequence of tools that an agent must call for a workflow to be correct.

## Assertion Types

Each assertion can combine multiple `expect` checks. All must pass for the assertion to pass.

### Content assertions

| Assertion | What it checks | Example |
|---|---|---|
| `contains` | Response includes all listed strings | `contains: ["Alice", "admin"]` |
| `not_contains` | Response excludes all listed strings | `not_contains: ["password", "secret"]` |
| `equals` | Response exactly matches (whitespace-trimmed) | `equals: "42"` |
| `matches_regex` | Response matches all regex patterns | `matches_regex: ["\\d+ rows"]` |
| `in_order` | Substrings appear in this order | `in_order: ["header", "body", "footer"]` |

### Error assertions

| Assertion | What it checks | Example |
|---|---|---|
| `not_error` | Tool returned `isError: false` | `not_error: true` |
| `is_error` | Tool returned `isError: true` | `is_error: true` |
| `not_empty` | Response is non-empty (not `null`/`[]`/`{}`) | `not_empty: true` |

### Structured assertions

| Assertion | What it checks | Example |
|---|---|---|
| `json_path` | JSON field at dot-path matches value | `json_path: {"$.name": "Alice"}` |
| `min_results` | Array has at least N items | `min_results: 3` |
| `max_results` | Array has at most N items | `max_results: 10` |

### File system assertions

| Assertion | What it checks | Example |
|---|---|---|
| `file_contains` | File on disk has expected text after tool runs | `file_contains: {"{{fixture}}/out.txt": "done"}` |
| `file_unchanged` | File was not modified by the tool | `file_unchanged: ["{{fixture}}/readonly.txt"]` |

### Advanced assertions

| Assertion | What it checks | Example |
|---|---|---|
| `net_delta` | Speculative edit diagnostic delta equals N | `net_delta: 0` |
| `min_progress` | At least N progress notifications received | `min_progress: 3` (requires `capture_progress: true`) |
