# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [Unreleased]

### Added

- **`assert_completion` block type**: test `completion/complete` for prompt argument and resource URI autocompletion. Specify `ref` (type + name) and `argument` (name + partial value), assert on the completion results. 3 example assertions in `examples/mcp-go-everything-completion/`.
- **`assert_sampling` block type**: first-class sampling test block that combines mock LLM config and tool call in one YAML block. Set `mock_text` and optional `mock_model`; the runner configures `client_capabilities.sampling` automatically, calls the specified tool, and asserts on the result.
- **`assert_logging` block type**: test `logging/setLevel` and `notifications/message` capture. Set `set_level` to configure log verbosity, optionally call a tool to trigger log messages, then assert on captured messages with `min_messages`, `contains_level`, and `contains_data`. 2 example assertions in `examples/mcp-go-everything-logging/`.
- **`init --server` one-step suite generation**: `mcp-assert init evals --server "cmd" --fixture ./fixtures` queries `tools/list`, generates one stub per tool, and captures response snapshots in a single command. Recommended path for new users.
- **`--fix` flag**: pass `--fix` to `run` or `ci` to scan nearby positions when position-sensitive assertions fail. Emits a suggested YAML patch showing the corrected line/column values.
- **Watch mode diff view**: when an assertion's status flips between iterations, watch mode displays a unified diff of expected vs actual to help diagnose changes.
- **`intercept` command**: proxy stdio between an agent and MCP server, capturing every `tools/call` invocation in real time. Validates the captured sequence against trajectory assertions on disconnect.
- **Fixture isolation**: each stdio assertion automatically receives its own copy of the fixture directory via a temporary directory. The original fixture is never modified, preventing side effects from one assertion affecting subsequent ones.
- **`${VAR}` environment variable expansion**: `env:` blocks in YAML now support `${VAR}` and `$VAR` syntax to reference host environment variables. If the variable is not set, the original string is preserved unchanged.
- **`--suite` accepts single files**: `--suite path/to/file.yaml` now works in addition to directories, for iterating on one assertion at a time.
- **`generate` skips destructive tools by default**: tools annotated as destructive or not read-only are generated with `skip: true`. Use `--include-writes` to include all tools.
- **`skip: true` field**: any assertion YAML can set `skip: true` to exclude it from execution. Skipped assertions appear as SKIP in output.
- **Elicitation decline/cancel flows**: `client_capabilities.elicitation` now supports decline and cancel flows in addition to accept. 4 assertions in `examples/mcp-go-elicitation/` cover all three.
- **Resource subscription support**: `assert_resources:` block supports `subscribe` and `unsubscribe` fields. 2 assertions in `examples/mcp-go-everything-resources/` cover subscribe and unsubscribe.
- **GitHub MCP Server suite**: 6 read-only assertions for `github/github-mcp-server` (28K+ stars) in `examples/github-mcp/`: `get_me`, `search_repositories`, `get_file_contents`, `list_issues`, `search_code`, `list_branches`.
- **fastmcp testing_demo suite expanded**: added resource and prompt assertions (3 resources + 2 prompts), now covers all three MCP server feature categories. Total: 16 assertions.
- **Prompts assertions**: `assert_prompts:` block tests `prompts/list` and `prompts/get`. 4 example assertions in `examples/mcp-go-everything-prompts/`.
- **Progress capture** (`capture_progress` + `min_progress`): collect `notifications/progress` messages during tool execution.
- **Resources assertions**: `assert_resources:` block tests `resources/list` and `resources/read`. 4 example assertions in `examples/mcp-go-everything-resources/`.
- **Client capabilities (roots, sampling, elicitation)**: `client_capabilities` in server YAML config supports `roots`, `sampling`, and `elicitation` for bidirectional MCP testing.
- **mcp-go sampling_server suite**: 3 assertions for `mark3labs/mcp-go` sampling_server. 100% tool coverage.
- **mcp-go elicitation suite**: 4 assertions for `mark3labs/mcp-go` elicitation server.
- **218 unit tests** across 3 packages (assertion: 53, report: 42, runner: 123). Race-detector clean.
- **174 total assertions** across 18 suites in 3 languages.

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
