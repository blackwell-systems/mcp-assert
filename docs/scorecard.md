# Scan-and-Contribute Scorecard

Servers tested by mcp-assert, bugs found, issues filed.

## Summary

| Metric | Count |
|--------|-------|
| Servers scanned | 6 |
| Server suites | 9 (including HTTP transport variant) |
| Languages tested | 3 (Go, TypeScript, Python) |
| Transports tested | 3 (stdio, SSE, HTTP) |
| Total assertions | 119 |
| Upstream bugs found | 2 |
| Upstream issues filed | 2 |
| Clean scans (no bugs) | 3 |
| Internal bugs fixed | 5 |

## Server Results

### Anthropic Official Servers

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `@modelcontextprotocol/server-filesystem` | TypeScript | stdio | 14 | 92% (13/14) | 1 | [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029). `read_media_file` returns `type: "blob"`, violating MCP 2025-11-25 spec |
| `@modelcontextprotocol/server-memory` | TypeScript | stdio | 5 | - | 0 | Clean |
| `mcp-server-sqlite` | Python | stdio | 6 | - | 0 | Clean |

### Community Framework SDKs

| Server | Language | Transport | Assertions | Coverage | Bugs | Issue |
|--------|----------|-----------|------------|----------|------|-------|
| `mark3labs/mcp-go` everything | Go | stdio | 9 | 100% | 1 | [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826). `longRunningOperation` crashes stdio transport (fmt.Printf to stdout corrupts JSON-RPC) |
| `mark3labs/mcp-go` everything | Go | HTTP | 5 | 100% | 0 | Transport conformance: same tools pass over HTTP |
| `mark3labs/mcp-go` typed_tools | Go | stdio | 3 | 100% | 0 | Clean |
| `mark3labs/mcp-go` structured | Go | stdio | 6 | 100% | 0 | Clean |
| `PrefectHQ/fastmcp` testing_demo | Python | stdio | 11 | 100% | 0 | Clean. Pydantic validation handles edge cases correctly |

### Internal (agent-lsp)

| Server | Language | Transport | Assertions | Coverage | Bugs fixed |
|--------|----------|-----------|------------|----------|------------|
| agent-lsp + gopls | Go | stdio | 60 | 100% (50/50 tools) | 5: `character`→`column` param rename, `format_range` 0-indexed docs, undocumented `simulate_edit_atomic` params, missing warmup pattern, shared fixture mutation |

## Bug Details

### Bug #1: Anthropic filesystem: invalid MCP content type

- **Severity:** Spec violation
- **Tool:** `read_media_file`
- **What:** Returns `type: "blob"` which is not a valid MCP content type. The spec allows `text`, `image`, `audio`, `resource_link`, `resource`.
- **Impact:** Any MCP client receiving this response crashes at the transport layer.
- **Issue:** [modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029)
- **Status:** Open

### Bug #2: mcp-go SDK: stdio transport crash on slow tools

- **Severity:** Transport crash
- **Tool:** `longRunningOperation` in `examples/everything`
- **What:** Tool calls `time.Sleep()` which creates a timing window for `fmt.Printf` hooks to corrupt the stdio JSON-RPC stream.
- **Impact:** Any mcp-go server with debug hooks and slow tool handlers will crash over stdio.
- **Issue:** [mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826)
- **Status:** Open, linked to project board

## Observations

**Bug rate:** 2 bugs in 6 servers scanned (33%). Both were transport/protocol-level issues, not logic bugs.

**Clean scans are valuable too.** fastmcp's clean result (25K-star framework, zero bugs) validates the Python MCP ecosystem's foundations. We document clean scans as positive signals, not wasted effort.

**The flywheel works.** Each issue filed links back to mcp-assert. Maintainers discovering the tool through bug reports is organic adoption: no marketing required.

**Transport bugs are invisible to unit tests.** Both bugs were in the transport layer: they'd never show up in the server's own unit tests because those test tool logic, not MCP protocol compliance. This is mcp-assert's core value proposition.
