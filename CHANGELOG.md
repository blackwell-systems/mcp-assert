# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [0.2.0] - 2026-04-24

### Added

- **`intercept` command**: Live stdio proxy between agent and MCP server. Captures every `tools/call` in real time and validates trajectory assertions on disconnect. `mcp-assert intercept --server <cmd> --trajectory <yaml>`.
- **`--fix` flag**: Position error correction for `run` and `ci` commands. When a position-sensitive assertion fails ("no identifier found", "column is beyond end of line"), scans nearby positions (line ±3, column ±5) and emits a suggested YAML patch.
- **Watch mode diff view**: When an assertion flips status between iterations in `--watch` mode, shows a unified diff of expected vs actual response.
- **`init --server`**: One-step suite generation. `mcp-assert init evals --server "my-server"` runs generate + snapshot in one command. Complete working suite with zero manual assertion writing.
- **Fixture isolation**: Each assertion automatically receives its own copy of the fixture directory. Write-tests can never contaminate read-tests. Automatic for stdio; Docker already isolates via containers.
- **`${VAR}` env expansion**: Shell variable patterns (`${VAR}` and `$VAR`) in YAML `env:` blocks resolve from the parent process environment at runtime.
- **Single-file `--suite`**: `--suite` accepts both directories and single YAML files for iterative development.
- **`generate` safety**: Destructive tools (based on MCP `destructiveHint` annotations) are generated with `skip: true` by default. Use `--include-writes` to opt in. Auth detection hints when the server exits immediately (transport closed).
- **`skip` field**: Assertions with `skip: true` are skipped during execution. Set automatically by `generate` for destructive tools.
- **`assert_completion` block**: Test `completion/complete` for argument autocompletion. `ref: {type, name}` + `argument: {name, value}`.
- **`assert_sampling` block**: First-class sampling test combining mock LLM config and tool call. `mock_text`, `mock_model`, tool, args, expect.
- **`assert_logging` block**: Test `logging/setLevel` and capture `notifications/message`. `set_level`, `expect: {min_messages, contains_level, contains_data}`.
- **Resource subscriptions**: `subscribe` and `unsubscribe` fields on `assert_resources` block for `resources/subscribe` and `resources/unsubscribe`.
- **Elicitation breadth**: `action: decline` and `action: cancel` in `client_capabilities.elicitation` for testing reject/cancel flows. 3 new example assertions.
- **GitHub MCP Server suite**: 6 assertions against github/github-mcp-server (28K+ stars). get_me, search_repositories, get_file_contents, list_issues, search_code, list_branches.
- **CONTRIBUTING.md**: Contributor guide covering package structure, adding assertion types, CLI commands, and block types.
- **Dogfooding report**: `docs/dogfooding-github-mcp.md` documenting the end-to-end onboarding experience against the most popular MCP server.
- **Protocol surface 10/12**: Resource subscriptions, completion, logging covered. Only cancellation and ping remain.

### Fixed

- **Fixture contamination**: apply_edit and commit_session tests no longer modify shared fixture files. Dedicated fixture files (`apply_edit_fixture.go`, `commit_fixture.go`) plus automatic per-test isolation.

## [0.1.3] - 2026-04-23

### Added

- See [Unreleased] entries above for full list. Key additions: 7 block types (assert, assert_prompts, assert_resources, assert_completion, assert_sampling, assert_logging, trajectory), intercept command, init --server, --fix mode, fixture isolation, env expansion, 18 example suites, 174 assertions, 218 unit tests.

## [0.1.2] - 2026-04-23

### Added

- **Trajectory assertions**: validate that agents call MCP tools in the correct order. Essential for skill protocol compliance. 4 assertion types: `order` (tools appear in sequence), `presence` (all listed tools called), `absence` (tools NOT called), `args_contain` (tool called with specific argument values). Inline trace (YAML) and audit log (JSONL) sources. Runs in 0ms (no server needed). **20 example assertions covering all agent-lsp skills**: rename, safe-edit, refactor, cross-repo, dead-code, docs, edit-export, edit-symbol, explore, extract-function, fix-all, format-code, generate, impact, implement, local-symbols, simulate, test-correlation, understand, verify. Each assertion captures the required tool call sequence, safety gates, and absence checks for its skill protocol.
- **fastmcp example suite.** 11 assertions for [PrefectHQ/fastmcp](https://github.com/PrefectHQ/fastmcp) (25K stars), the most popular Python MCP framework. Tests the `testing_demo` example server: `add` (4 assertions), `greet` (3), `async_multiply` (4). Covers happy paths, edge cases (negative numbers, zero, empty strings, fractional results), default vs custom parameters, and missing-argument error handling. 100% tool coverage. First Python framework (as opposed to Python server) in the example suite collection.
- **HTTP/SSE transport support.** Test MCP servers over HTTP, not just stdio. Set `transport: sse` or `transport: http` in assertion YAML with a `url` field to connect to remote/HTTP-based MCP servers. Uses mcp-go's `NewSSEMCPClient` (legacy SSE) and `NewStreamableHttpClient` (streamable HTTP). Client creation is unified via `createMCPClient` helper shared across `run`, `snapshot`, and `coverage` commands. Docker isolation remains stdio-only. 11 new unit tests for transport selection, URL validation, and error paths.
- **Snapshot testing.** `mcp-assert snapshot --suite <dir> [--update]` captures tool responses as `.snapshots.json`, compares on subsequent runs. Like `jest --updateSnapshot` for MCP servers. Eliminates manual assertion writing for initial coverage.
- **`--watch` mode.** `mcp-assert watch --suite <dir>` reruns assertions on YAML file change. Polls every 2s, clears terminal between runs. Assertion development loop.
- **`--coverage-json`.** Machine-readable coverage data for dashboards and badges on the `coverage` command.
- **pass@k / pass^k in structured reports.** Reliability metrics now included in JUnit XML (`<properties>`) and markdown (table section) when `--trials > 1`. Previously terminal-only.
- **`generate` command.** `mcp-assert generate --server <cmd> --output <dir>` queries `tools/list`, reads input schemas, creates one YAML stub per tool with sensible defaults (path params get `{{fixture}}/TODO`, integers default to 1). Combined with snapshot: `generate` + `snapshot --update` = zero-effort 100% coverage.
- **Setup output capture.** Setup steps can capture values from responses via jsonpath and inject them into subsequent steps using `{{variable}}` syntax. Unlocks real session lifecycle testing (create → edit → evaluate → commit/discard → destroy) instead of negative-only tests.
- **9 multi-step workflow assertions** using capture: session lifecycle (create→edit→evaluate→discard→destroy), session commit, session evaluate with net_delta, simulate_chain multi-edit, commit-verify-file, rename-verify-references, diagnostics-after-error, code-actions-for-error, open-close-reopen.
- **`mcp-assert init`.** Scaffolds a commented assertion template and fixture directory for new users.
- **Docs site.** mkdocs with Material theme. README slimmed from 553→103 lines. 8 pages: getting-started, writing-assertions, cli reference, examples, ci-integration, architecture, roadmap, dogfooding.
- **mcp-go SDK example suites.** 18 assertions across 3 servers from the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) SDK: everything (9, 100% coverage), typed_tools (3, 100%), structured_input_and_output (6, 100%). Found transport crash bug ([mark3labs/mcp-go#826](https://github.com/mark3labs/mcp-go/issues/826)).
- **HTTP transport conformance tests.** 5 assertions testing mcp-go everything server over HTTP, same tools and expectations as stdio suite. Proves transport-agnostic testing works end-to-end.
- **136 total assertions** across 9 server suites in 3 languages: filesystem (14), memory (5), sqlite (6), agent-lsp (60), mcp-go-everything (9), mcp-go-typed-tools (3), mcp-go-structured (6), mcp-go-everything-http (5), fastmcp-testing-demo (11). Trajectory suite added 20 skill protocol assertions.
- **111 unit tests** (up from 49). Runner: 53 tests (substitution, capture, extractJSONPath, overrides, error paths, timeout, Docker, generate, transport selection). Race-detector clean.
- **GitHub Pages docs site.** Dark mode, Material theme, auto-deployed on push.

## [0.1.1] - 2026-04-23

### Fixed

- **51/51 agent-lsp assertions passing.** Fixed 15 assertion failures caused by `apply_edit` modifying the fixture file and shifting line numbers for subsequent tests. Added `get_diagnostics` warmup steps, corrected column positions, converted `restart_lsp_server` and `cross_repo_references` to negative tests.
- **Recursive `{{fixture}}` substitution.** Template replacement now recurses into arrays and nested maps, not just top-level string values. Fixes assertions using list parameters (e.g. `read_multiple_files`).

### Added

- **Filesystem 92% coverage.** 14 assertions for `@modelcontextprotocol/server-filesystem` (up from 5). 13/14 tools covered. `read_media_file` excluded due to upstream spec violation ([modelcontextprotocol/servers#4029](https://github.com/modelcontextprotocol/servers/issues/4029)).
- **agent-lsp 100% coverage.** 51 assertions covering all 50 tools. Navigation, refactoring, analysis, session lifecycle, workspace management, build, and config tools.
- **SQLite example suite.** 6 assertions for `mcp-server-sqlite` (Python): list tables, SELECT, COUNT, JOIN, describe schema, invalid SQL error.

## [0.1.0] - 2026-04-23

### Added

- **Core assertion engine** with 14 deterministic assertion types: `contains`, `not_contains`, `equals`, `json_path`, `min_results`, `max_results`, `not_empty`, `not_error`, `is_error`, `matches_regex`, `file_contains`, `file_unchanged`, `net_delta`, `in_order`.
- **CLI commands**: `run`, `matrix`, `ci`, `coverage`.
- **`coverage` command.** Queries the MCP server's `tools/list`, compares against assertion tool names, reports coverage percentage with covered/uncovered tool lists.
- **`--server` flag** on `run` and `ci` overrides per-YAML server config from CLI.
- **Color output.** Green `✓` pass, red `✗` fail, yellow `○` skip. Respects `NO_COLOR`. Progress counter on stderr.
- **Structured reporting.** `--junit` (JUnit XML), `--markdown` (GitHub Step Summary), `--badge` (shields.io JSON).
- **Docker isolation.** `--docker <image>` wraps MCP server in `docker run -i`.
- **Reliability metrics.** `--trials N` prints `pass@k` / `pass^k` table.
- **Regression detection.** `--baseline`, `--save-baseline`, `--fail-on-regression`.
- **GoReleaser.** Cross-compiled binaries for linux/darwin/windows × amd64/arm64.
- **GitHub Action.** [`blackwell-systems/mcp-assert-action@v1`](https://github.com/blackwell-systems/mcp-assert-action) for one-line CI adoption.
- **Example suites for 4 MCP servers** in 3 languages (TypeScript, Python, Go): filesystem (14), memory (5), sqlite (6), agent-lsp (51).
- **CI pipeline.** 5 jobs: unit tests with `-race` (49 tests), e2e-filesystem, e2e-memory, e2e-sqlite, e2e-agent-lsp. All upload JUnit XML.
- **Documentation.** README, FEATURES.md, architecture.md, roadmap.md, distribution.md, dogfooding-findings.md.
