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
- **27 unit tests** across checker (14), loader (8), report (5). All assertion types tested including edge cases.
- **End-to-end verified** against real agent-lsp + gopls. All 7 example assertions pass: hover, definition, references, diagnostics, symbols, completions, speculative execution.
- **CI workflow** for build, vet, and test.
