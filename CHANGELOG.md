# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [Unreleased]

### Added

- **fastmcp testing_demo suite expanded**: added resource and prompt assertions (3 resources + 2 prompts), now covers all three MCP server feature categories. Total: 16 assertions for fastmcp. Resources: `resources/list` (static resource + template), `resources/read demo://info`, `resources/read demo://greeting/Alice`. Prompts: `prompts/list` (hello + explain), `prompts/get hello` with name argument.

- **Prompts assertions**: `assert_prompts:` block tests `prompts/list` and `prompts/get` — prompt-template servers are now fully testable. Use `list: {}` to assert which prompts the server exposes, or `get: {name: "...", arguments: {...}}` to retrieve a rendered prompt and assert on its content. `{{fixture}}` and captured variables are substituted in prompt names and arguments. Verified against the mcp-go `everything` server (`simple_prompt` and `complex_prompt`). 4 example assertions in `examples/mcp-go-everything-prompts/`.
- **Progress capture** (`capture_progress` + `min_progress`): set `capture_progress: true` on an `assert:` block to collect `notifications/progress` messages sent by the server during tool execution. Use `min_progress: N` in `expect:` to assert the server sent at least N progress notifications. Requires a server that properly sends progress notifications (note: the mcp-go `everything` server's `longRunningOperation` has a known stdio transport bug that prevents reliable progress capture with that server).
- **Pagination documentation**: cursor-based pagination on `prompts/list`, `resources/list`, and `tools/list` is covered via `json_path` assertions on the marshaled response. Use `json_path: {"$.nextCursor": "page2-cursor"}` to assert on cursor values, or pass `cursor: "value"` in the `list:` block to request a specific page. Example in `examples/mcp-go-everything-prompts/pagination_pattern.yaml`.
- **13 new unit tests**: `TestRunAssertion_Prompts_RequiresListOrGet`, `TestRunAssertion_Prompts_BadServer`, `TestRunAssertion_Prompts_FixtureSubstitutionInGetArgs`, `TestRunAssertion_Progress_CounterInitializedToZero`, `TestRunAssertion_Progress_DoesNotBreakExistingAssertions`, `TestCheckProgress_PassesWhenCountMeetsMinimum`, `TestCheckProgress_PassesWhenCountExceedsMinimum`, `TestCheckProgress_FailsWhenCountBelowMinimum`, `TestCheckProgress_NoMinProgress_AlwaysPasses`. Total unit tests: 157.
- **Resources assertions**: `assert_resources:` block tests `resources/list` and `resources/read` — whole category of MCP servers (document stores, knowledge bases, file systems) now fully testable. Use `list: {}` to assert what resources the server exposes, `read: "uri"` to assert resource content. `{{fixture}}` substitution works in URIs. Verified against mcp-go `everything` server.
- **Client capabilities (roots, sampling, elicitation)**: servers that make requests back to the client can now be fully tested via `client_capabilities` in server YAML config. Set `roots: [paths]` to respond to `roots/list`, `sampling: {text, model, stop_reason}` to respond to `sampling/createMessage` with a mock LLM response, or `elicitation: {content: {...}}` to respond to `elicitation/create` with preset values. Verified against mcp-go `roots_server`, `sampling_server`, and `elicitation` example servers. Makes mcp-assert the only MCP testing tool that can fully simulate a bidirectional MCP client environment.
- **mcp-go sampling_server suite**: 3 assertions for `mark3labs/mcp-go` sampling_server — `ask_llm` with and without a custom system prompt, and `greet` (verifying that non-sampling tools work normally when `client_capabilities.sampling` is set). 100% tool coverage.
- **mcp-go elicitation suite**: 1 assertion for `mark3labs/mcp-go` elicitation server — `create_project` responding to a form-based elicitation request with preset field values (`projectName`, `framework`, `includeTests`).
- **19 new unit tests** for client capabilities: `TestStaticRootsHandler_ListRoots`, `TestStaticRootsHandler_EmptyRoots`, `TestStaticSamplingHandler_CreateMessage`, `TestStaticSamplingHandler_DefaultModelAndStopReason`, `TestStaticElicitationHandler_ReturnsContentAndAccept`, `TestStaticElicitationHandler_FallsBackToWholeMap`, `TestCreateStdioClientWithCapabilities_FixtureSubstitution`, `TestCreateMCPClient_WithRoots_UsesCapabilityPath`, `TestCreateMCPClient_WithSampling_UsesCapabilityPath`, `TestCreateMCPClient_WithElicitation_UsesCapabilityPath`, `TestCreateMCPClient_EmptyClientCapabilities_UsesSimplePath`, `TestRunAssertion_ClientCapabilities_BadServer`. Total unit tests: 144.

## [0.1.3] - 2026-04-23

### Added

- See [Unreleased] entries above — trajectory assertions, prompts, resources, progress capture, client capabilities, fastmcp expansion, 157 unit tests.

## [0.1.2] - 2026-04-23

### Added

- **Trajectory assertions** — validate that agents call MCP tools in the correct order. Essential for skill protocol compliance. 4 assertion types: `order` (tools appear in sequence), `presence` (all listed tools called), `absence` (tools NOT called), `args_contain` (tool called with specific argument values). Inline trace (YAML) and audit log (JSONL) sources. Runs in 0ms (no server needed). **20 example assertions covering all agent-lsp skills**: rename, safe-edit, refactor, cross-repo, dead-code, docs, edit-export, edit-symbol, explore, extract-function, fix-all, format-code, generate, impact, implement, local-symbols, simulate, test-correlation, understand, verify. Each assertion captures the required tool call sequence, safety gates, and absence checks for its skill protocol.
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
- **136 total assertions** across 9 server suites in 3 languages: filesystem (14), memory (5), sqlite (6), agent-lsp (60), mcp-go-everything (9), mcp-go-typed-tools (3), mcp-go-structured (6), mcp-go-everything-http (5), fastmcp-testing-demo (11). Trajectory suite adds 20 skill protocol assertions.
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
