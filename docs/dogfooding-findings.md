# Dogfooding Findings

What we found by using mcp-assert to test agent-lsp (our own MCP server) and the official Anthropic MCP servers. Each finding is a real bug or inconsistency that would silently hurt agents in production.

---

## agent-lsp findings

### 1. `character` vs `column` parameter inconsistency

**What:** `get_symbol_source` used `character` as its position parameter. All 49 other tools use `column`. An agent switching between `get_references` (column) and `get_symbol_source` (character) would silently pass the wrong parameter and get no results.

**How mcp-assert found it:** Writing assertion YAML for `get_symbol_source`: the assertion author instinctively used `column` (consistent with every other tool), the assertion failed, and the error message revealed the parameter was called `character`.

**Fix:** Renamed to `column` in the schema. Implementation accepts both for backward compatibility.

### 2. `format_range` schema said "0-indexed" but implementation was 1-indexed

**What:** The JSON Schema description for `format_range` said "Start line of the range to format (0-indexed)" but the implementation validated `>= 1`. Agents following the schema would send `start_line: 0` and get a cryptic error.

**How mcp-assert found it:** The `format_range` assertion used 1-indexed positions (matching the convention of every other tool). It passed. Reading the schema while writing the assertion revealed the "0-indexed" claim, which contradicted both the implementation and all other tools.

**Fix:** Corrected schema descriptions to "1-indexed".

### 3. `simulate_edit_atomic` requires undocumented parameters

**What:** When used without a session, `simulate_edit_atomic` requires `workspace_root` and `language` parameters. The error messages were just "workspace_root is required" and "language is required": clear enough, but not discoverable from the schema alone.

**How mcp-assert found it:** The first speculative assertion used only `file_path` and position params (matching the schema). Failed with `isError: true`. Adding server response surfacing to mcp-assert revealed the missing params one at a time.

**Fix:** Added server error text to mcp-assert failure output. The schema already listed the params as optional: the issue was that they're conditionally required (required when no session_id is provided).

### 4. Cross-file tools need warmup that isn't documented

**What:** `get_references`, `call_hierarchy`, `rename_symbol`, and other cross-file tools return empty or error results when called immediately after `start_lsp`. gopls needs time to index the workspace.

**How mcp-assert found it:** The `get_references` assertion passed locally (gopls indexed fast enough) but failed in CI (slower runner, gopls not ready). Adding `get_diagnostics` as a warmup step: which blocks until gopls returns diagnostics: fixed it.

**Fix:** Documented the warmup pattern in ci-notes.md. All cross-file assertions now include `get_diagnostics` setup steps.

### 5. `{{fixture}}` substitution didn't recurse into arrays

**What:** mcp-assert's `{{fixture}}` template substitution only replaced strings in top-level map values. Array elements (like `paths: ["{{fixture}}/a.txt", "{{fixture}}/b.txt"]`) were not substituted.

**How it was found:** The `read_multiple_files` assertion for the filesystem server passed an array of paths. The `{{fixture}}` placeholders were sent literally to the server, which rejected them as "path outside allowed directories."

**Fix:** Made substitution recursive: handles strings, arrays, and nested maps.

---

## @modelcontextprotocol/server-filesystem findings

### 6. `read_media_file` returns invalid MCP response for non-media files

**What:** Calling `read_media_file` on a `.txt` file causes the server to return a malformed MCP response that violates the protocol schema. The MCP client can't parse it, resulting in a transport-level error rather than a clean tool error.

**How mcp-assert found it:** Writing a comprehensive assertion suite for all 14 filesystem tools. The `read_media_file` assertion crashed the MCP transport layer instead of returning `isError: true`.

**Status:** Filed as [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029). The `type: "blob"` content type violates the MCP 2025-11-25 specification which only allows `text`, `image`, `audio`, `resource_link`, and `resource`.

### 7. Shared fixture mutation causes cascading position failures

**What:** `apply_edit.yaml` runs first (alphabetical order) and inserts a comment line into `main.go`, shifting all line numbers by +1. Every subsequent assertion pointing at `main.go` was hitting the wrong line: landing on comments instead of identifiers.

**How mcp-assert found it:** 12 of 15 assertion failures had the same root cause: "no identifier found" or empty responses at positions that should have worked. The pattern: all failures in the same file, all off by exactly one line: pointed to a shared-state mutation.

**Fix:** Adjusted all position-dependent assertions to account for the inserted line.

**Permanent fix shipped:** Fixture isolation. Each stdio assertion now automatically receives its own temporary copy of the fixture directory. The original fixture is never modified, so write-heavy assertions cannot shift line numbers or alter state for subsequent assertions. This eliminated the root cause entirely.

---

## External server findings

### 8. mcp-go SDK: longRunningOperation crashes stdio transport

**What:** The `longRunningOperation` tool in `mark3labs/mcp-go`'s `examples/everything` server causes `transport error: transport closed` when called over stdio, even with minimal params (`duration: 0.1`, `steps: 1`). The tool handler calls `time.Sleep()` which races with stdio transport teardown.

**How mcp-assert found it:** Building a comprehensive assertion suite for the mcp-go SDK's 3 example servers. 17/18 assertions passed. The `longRunningOperation` crash was the only failure.

**Status:** Filed as [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826). This is a bug in the Go MCP SDK that mcp-assert itself depends on.

**Impact:** This is the first external community server we scanned. Finding a bug in the SDK we depend on validates the scan-and-contribute flywheel strategy.

---

## PrefectHQ/fastmcp findings

### 9. fastmcp testing_demo: clean results: no bugs found

**What:** 11 assertions across 3 tools (add, greet, async_multiply) all passed on the first run. Edge cases tested: negative numbers, zero, empty strings, fractional multiplication, missing required arguments, default vs custom optional parameters.

**What this tells us:** fastmcp's tool registration and Pydantic-based input validation handle edge cases well. Missing arguments correctly return `isError: true` with a validation error. Default parameter values work as expected. Async tools behave identically to sync tools from the MCP protocol perspective.

**Significance:** This is the first Python MCP framework (as opposed to a standalone Python MCP server like sqlite) we've tested. A clean bill of health for a 25K-star framework validates that the Python MCP ecosystem has solid foundations. Not every scan finds bugs: and that's a useful signal too.

---

## Patterns observed

**Writing assertions forces you to read your own schema.** You discover parameter naming inconsistencies, misleading descriptions, and undocumented requirements that you'd never notice from the implementation side.

**The first assertion for a tool is the hardest.** Once you know the exact parameter names, required setup steps, and expected response shape, subsequent assertions are trivial. The difficulty of writing the first assertion is a proxy for how hard the tool is for agents to use.

**Timing is the #1 source of flakiness.** Language servers need indexing time. Assertions that work locally fail in CI. The `get_diagnostics` warmup pattern is the universal fix.

**Negative tests catch real bugs.** The `read_media_file` finding came from a negative test (testing error handling). The `is_error` assertion type exists specifically for this: and it found an upstream protocol violation.
