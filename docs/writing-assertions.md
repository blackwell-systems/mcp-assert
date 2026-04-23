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
  tool: search_nodes                        # Runs last — this is what we're testing
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
name: session lifecycle — create, edit, evaluate
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

Captured variables work anywhere `{{fixture}}` works — strings, arrays, nested objects. Use this for session IDs, auth tokens, created resource IDs, or any value returned by a setup step.

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
