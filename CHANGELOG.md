# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [Unreleased]

### Added

- **HTTP/SSE transport support** — test MCP servers over HTTP, not just stdio. Set `transport: sse` or `transport: http` in assertion YAML with a `url` field to connect to remote/HTTP-based MCP servers. Uses mcp-go's `NewSSEMCPClient` (legacy SSE) and `NewStreamableHttpClient` (streamable HTTP). Client creation is unified via `createMCPClient` helper shared across `run`, `snapshot`, and `coverage` commands. Docker isolation remains stdio-only. 11 new unit tests for transport selection, URL validation, and error paths.
- **Snapshot testing** — `mcp-assert snapshot --suite <dir> [--update]` captures tool responses as `.snapshots.json`, compares on subsequent runs. Like `jest --updateSnapshot` for MCP servers. Eliminates manual assertion writing for initial coverage.
- **`--watch` mode** — `mcp-assert watch --suite <dir>` reruns assertions on YAML file change. Polls every 2s, clears terminal between runs. Assertion development loop.
- **`--coverage-json`** — machine-readable coverage data for dashboards and badges on the `coverage` command.
- **pass@k / pass^k in structured reports** — reliability metrics now included in JUnit XML (`<properties>`) and markdown (table section) when `--trials > 1`. Previously terminal-only.
- **`generate` command** — `mcp-assert generate --server <cmd> --output <dir>` queries `tools/list`, reads input schemas, creates one YAML stub per tool with sensible defaults (path params get `{{fixture}}/TODO`, integers default to 1). Combined with snapshot: `generate` + `snapshot --update` = zero-effort 100% coverage.
- **Setup output capture** — setup steps can capture values from responses via jsonpath and inject them into subsequent steps using `{{variable}}` syntax. Unlocks real session lifecycle testing (create → edit → evaluate → commit/discard → destroy) instead of negative-only tests.
- **9 multi-step workflow assertions** using capture: session lifecycle (create→edit→evaluate→discard→destroy), session commit, session evaluate with net_delta, simulate_chain multi-edit, commit-verify-file, rename-verify-references, diagnostics-after-error, code-actions-for-error, open-close-reopen.
- **`mcp-assert init`** — scaffolds a commented assertion template and fixture directory for new users.
- **Docs site** — mkdocs with Material theme. README slimmed from 553→103 lines. 8 pages: getting-started, writing-assertions, cli reference, examples, ci-integration, architecture, roadmap, dogfooding.
- **mcp-go SDK example suites** — 18 assertions across 3 servers from the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) SDK: everything (9, 100% coverage), typed_tools (3, 100%), structured_input_and_output (6, 100%). Found transport crash bug ([mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826)).
- **HTTP transport conformance tests** — 5 assertions testing mcp-go everything server over HTTP, same tools and expectations as stdio suite. Proves transport-agnostic testing works end-to-end.
- **108 total assertions** across 8 server suites in 3 languages: filesystem (14), memory (5), sqlite (6), agent-lsp (60), mcp-go-everything (9), mcp-go-typed-tools (3), mcp-go-structured (6), mcp-go-everything-http (5).
- **111 unit tests** (up from 49). Runner: 53 tests (substitution, capture, extractJSONPath, overrides, error paths, timeout, Docker, generate, transport selection). Race-detector clean.
- **GitHub Pages docs site** — dark mode, Material theme, auto-deployed on push.

## [0.1.1] - 2026-04-23

### Fixed

- **51/51 agent-lsp assertions passing** — fixed 15 assertion failures caused by `apply_edit` modifying the fixture file and shifting line numbers for subsequent tests. Added `get_diagnostics` warmup steps, corrected column positions, converted `restart_lsp_server` and `cross_repo_references` to negative tests.
- **Recursive `{{fixture}}` substitution** — template replacement now recurses into arrays and nested maps, not just top-level string values. Fixes assertions using list parameters (e.g. `read_multiple_files`).

### Added

- **Filesystem 92% coverage** — 14 assertions for `@modelcontextprotocol/server-filesystem` (up from 5). 13/14 tools covered. `read_media_file` excluded due to upstream spec violation ([modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029)).
- **agent-lsp 100% coverage** — 51 assertions covering all 50 tools. Navigation, refactoring, analysis, session lifecycle, workspace management, build, and config tools.
- **SQLite example suite** — 6 assertions for `mcp-server-sqlite` (Python): list tables, SELECT, COUNT, JOIN, describe schema, invalid SQL error.

## [0.1.0] - 2026-04-23

### Added

- **Core assertion engine** with 14 deterministic assertion types: `contains`, `not_contains`, `equals`, `json_path`, `min_results`, `max_results`, `not_empty`, `not_error`, `is_error`, `matches_regex`, `file_contains`, `file_unchanged`, `net_delta`, `in_order`.
- **CLI commands**: `run`, `matrix`, `ci`, `coverage`.
- **`coverage` command** — queries the MCP server's `tools/list`, compares against assertion tool names, reports coverage percentage with covered/uncovered tool lists.
- **`--server` flag** on `run` and `ci` overrides per-YAML server config from CLI.
- **Color output** — green `✓` pass, red `✗` fail, yellow `○` skip. Respects `NO_COLOR`. Progress counter on stderr.
- **Structured reporting** — `--junit` (JUnit XML), `--markdown` (GitHub Step Summary), `--badge` (shields.io JSON).
- **Docker isolation** — `--docker <image>` wraps MCP server in `docker run -i`.
- **Reliability metrics** — `--trials N` prints `pass@k` / `pass^k` table.
- **Regression detection** — `--baseline`, `--save-baseline`, `--fail-on-regression`.
- **GoReleaser** — cross-compiled binaries for linux/darwin/windows × amd64/arm64.
- **GitHub Action** — [`blackwell-systems/mcp-assert-action@v1`](https://github.com/blackwell-systems/mcp-assert-action) for one-line CI adoption.
- **Example suites for 4 MCP servers** in 3 languages (TypeScript, Python, Go): filesystem (14), memory (5), sqlite (6), agent-lsp (51).
- **CI pipeline** — 5 jobs: unit tests with `-race` (49 tests), e2e-filesystem, e2e-memory, e2e-sqlite, e2e-agent-lsp. All upload JUnit XML.
- **Documentation** — README, FEATURES.md, architecture.md, roadmap.md, distribution.md, dogfooding-findings.md.
