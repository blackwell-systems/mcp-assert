# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [Unreleased]

### Added

- **Core assertion engine** with 11 deterministic assertion types: `contains`, `equals`, `json_path`, `min_results`, `max_results`, `not_empty`, `not_error`, `file_contains`, `file_unchanged`, `net_delta`, `in_order`.
- **YAML assertion format** with server config, setup steps, and expected outputs. Supports `{{fixture}}` template substitution.
- **CLI commands**: `run` (execute assertions), `matrix` (cross-language), `ci` (threshold + regression detection).
- **MCP client** via mcp-go stdio transport. Starts the MCP server under test, sends tool calls, captures responses.
- **Cross-language matrix mode** runs the same assertions across multiple language server configurations.
- **Report output** with pass/fail/skip per assertion, duration, and failure details.
- **Example assertions** for agent-lsp Go fixtures: hover, definition, references, diagnostics, symbols, completions, speculative execution.
- **CI workflow** for build, vet, and test.
