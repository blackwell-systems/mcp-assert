# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [Unreleased]

### Added

- **Core assertion engine** with 13 deterministic assertion types: `contains`, `not_contains`, `equals`, `json_path`, `min_results`, `max_results`, `not_empty`, `not_error`, `is_error`, `matches_regex`, `file_contains`, `file_unchanged`, `net_delta`, `in_order`.
- **YAML assertion format** with server config, setup steps, and expected outputs. Supports `{{fixture}}` template substitution.
- **CLI commands**: `run` (execute assertions), `matrix` (cross-language), `ci` (threshold + regression detection).
- **`--server` flag** on `run` and `ci` commands overrides the per-YAML server config from CLI, so assertions can be reused across different servers.
- **MCP client** via mcp-go stdio transport. Starts the MCP server under test, sends tool calls, captures responses.
- **Cross-language matrix mode** runs the same assertions across multiple language server configurations.
- **Report output** with pass/fail/skip per assertion, duration, and failure details. Server error text surfaced in failure output for debugging.
- **Structured reporting** — three output formats for CI integration:
  - `--junit <path>`: JUnit XML for GitHub Actions test result tabs, Jenkins, CircleCI
  - `--markdown <path>`: GitHub Step Summary table (auto-detects `$GITHUB_STEP_SUMMARY` in `ci` mode)
  - `--badge <path>`: shields.io endpoint JSON for README badges
- **36 unit tests** across checker (14), loader (8), report (14). All assertion types and report formats tested.
- **End-to-end verified** against real agent-lsp + gopls. All 7 example assertions pass: hover, definition, references, diagnostics, symbols, completions, speculative execution.
- **Example suites for 3 MCP servers** — not just agent-lsp:
  - `examples/filesystem/` — 5 assertions for `@modelcontextprotocol/server-filesystem` (read, list, info, search, path traversal rejection)
  - `examples/memory/` — 5 assertions for `@modelcontextprotocol/server-memory` (entities, relations, observations, graph, empty search)
  - `examples/agent-lsp-go/` — 7 assertions for agent-lsp + gopls
- **CI workflow** for build, vet, and test.
