# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [Unreleased]

### Added

- **`coverage` command** — `mcp-assert coverage --suite <dir> --server <cmd>` queries the MCP server's `tools/list`, compares against assertion tool names in the suite, and reports coverage percentage with covered/uncovered tool lists. Useful for identifying which server tools lack test coverage.
- **Color output** — green `✓` pass, red `✗` fail, yellow `○` skip in interactive terminals. Falls back to plain `PASS`/`FAIL`/`SKIP` in pipes and CI. Respects the `NO_COLOR` environment variable.
- **Progress indicator** — `[1/21]`, `[2/21]`, etc. printed to stderr while assertions run. Summary line only shows non-zero counts.

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
- **Docker isolation** — `--docker <image>` wraps the MCP server in `docker run -i` with fixture volume mounts and environment forwarding. Each assertion runs in a fresh container.
- **Reliability metrics** — `--trials N` now prints a `pass@k` / `pass^k` table: pass@k (capability: passed at least once), pass^k (reliability: passed every time), per-assertion pass rate.
- **Regression detection** — `--baseline <path>` compares current results against a saved baseline. `--save-baseline <path>` persists results. `--fail-on-regression` exits 1 when a previously-passing assertion regresses.
- **49 unit tests** with race detection across checker (14), loader (8), report (27). All assertion types, report formats, reliability metrics, and baseline operations tested.
- **End-to-end verified in CI** — 37 assertions across 4 MCP servers in 3 languages (TypeScript, Python, Go), all passing.
- **Example suites for 4 MCP servers**:
  - `examples/filesystem/` — 5 assertions for `@modelcontextprotocol/server-filesystem` (TypeScript: read, list, info, search, path traversal rejection)
  - `examples/memory/` — 5 assertions for `@modelcontextprotocol/server-memory` (TypeScript: entities, relations, observations, graph, empty search)
  - `examples/sqlite/` — 6 assertions for `mcp-server-sqlite` (Python: list tables, SELECT, COUNT, JOIN, describe schema, invalid SQL error)
  - `examples/agent-lsp-go/` — 21 assertions for agent-lsp + gopls (Go: hover, definition, references, diagnostics, symbols, completions, speculative edit, type_definition, implementation, call_hierarchy, workspace_symbols, signature_help, semantic_tokens, highlights, prepare_rename, rename, format, code_actions, change_impact, build, inlay_hints)
- **CI pipeline** — 5 jobs: unit tests with `-race`, e2e-filesystem (5), e2e-memory (5), e2e-sqlite (6, Python via uv), e2e-agent-lsp (21). All e2e jobs upload JUnit XML artifacts.
