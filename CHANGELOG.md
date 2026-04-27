# Changelog

All notable changes to this project will be documented in this file.
The format is based on Keep a Changelog, Semantic Versioning.

## [0.6.0] - 2026-04-26

### Added

- **`mcp-assert audit` command**: zero-config server quality audit. Point it at any MCP server: `mcp-assert audit --server "npx my-server"`. Connects, discovers all tools via `tools/list`, calls each one with schema-generated inputs, and reports a quality score. Classifies tools as healthy (proper error handling), crashed (internal error, panic, stack trace), or timed out. Pass `--output <dir>` to generate starter assertion YAML files for the CI workflow. Supports `--json` for structured output, `--docker <image>` for safe destructive tool testing in fresh containers, `--include-writes` for destructive tools without isolation, and all three transports (stdio, SSE, HTTP).
- **Per-assertion Docker isolation**: new `docker:` field in server YAML config. Each assertion runs in a fresh `docker run --rm -i` container, destroyed after completion. Enables safe testing of destructive/write tools without state leaking between assertions. Fixtures mounted via `-v`, env vars via `-e`. Per-assertion field takes precedence over CLI `--docker` flag.
- **Perplexity MCP server suite**: 4 assertions, clean. All tools return `isError:true` with 401 and API key guidance. Sixth AI company on the scorecard.
- **engram memory server suite**: expanded from 6 to 16 assertions (100% tool coverage). All 16 tools tested including writes (save, delete, update, merge, session lifecycle).
- **CodeGraphContext suite**: 16 assertions, clean. Code graph indexer with 21 tools.
- **u14app/deep-research suite**: 5 assertions via HTTP transport, clean.
- **steipete/Peekaboo suite**: 6 assertions, 1 bug found. First Swift MCP server. `image` returns internal error instead of `isError:true` for missing Screen Recording permission. Issue filed ([#108](https://github.com/steipete/Peekaboo/issues/108)).

- **`skip_unless_env` field**: conditional assertion skipping based on environment variables. Assertions that require credentials skip cleanly when the env var is not set, run normally when it is. Enables live-backend and no-auth assertions in the same suite.
- **Playwright suite expanded**: 10 to 14 assertions (67% coverage). Added tabs, navigate_back, press_key, wait_for.
- **Grafana live-backend assertions**: 3 assertions (search_dashboards, search_folders, list_datasources) that run when `GRAFANA_SERVICE_ACCOUNT_TOKEN` is set, skip otherwise.
- **SQLite suite expanded**: 6 to 9 assertions, 100% tool coverage (6/6 tools). Added CREATE TABLE, INSERT, write_query.
- **Memory suite expanded**: 5 to 9 assertions, 100% tool coverage (9/9 tools). Added open_nodes, delete_entities, delete_observations, delete_relations.
- **Anthropic git suite expanded**: 7 to 11 assertions, 92% tool coverage (11/12 tools). Added commit, add, reset, tag.
- **git-mcp (onmyway133) suite expanded**: 7 to 14 assertions, 39% coverage (14/36 tools). Added stash, tag, blame, grep, cherry-pick, remote, and more.
- **getsentry/XcodeBuildMCP suite**: 10 assertions, 27 tools, clean. Sentry-backed macOS server.
- **Anthropic Puppeteer suite**: 7 assertions, 1 bug found. `puppeteer_navigate` crashes on invalid URLs. Fix PR submitted ([modelcontextprotocol/servers#4051](https://github.com/modelcontextprotocol/servers/pull/4051)).
- **sammcj/mcp-devtools suite**: 5 assertions, 4 bugs found. Tool handler returns internal error instead of `isError:true` for all input validation failures. Fix PR submitted ([sammcj/mcp-devtools#258](https://github.com/sammcj/mcp-devtools/pull/258)).
- **Context7 (Upstash) suite**: 2 assertions, clean. Library resolution and docs search.
- **Chrome DevTools MCP suite**: 7 assertions, 29 tools, clean.
- **Mozilla Firefox DevTools suite**: 7 assertions, 29 tools, clean. Mozilla-backed.
- **Excalidraw Architect suite**: 4 assertions, clean. Diagram generation.
- **SEC EDGAR suite**: 5 assertions, clean. Free public financial data. Uses `skip_unless_env`.
- **mcp-math suite**: 4 assertions, 16 tools, clean. Math operations.
- **DuckDuckGo suite**: 2 assertions, clean. Search and fetch.
- **Kubernetes suite**: 2 assertions, clean. kubectl error handling.
- **mobile-next/mobile-mcp suite**: 6 assertions, 21 tools, clean. Mobile automation. 4.7K stars.
- **spec-workflow-mcp suite**: 1 assertion, clean.
- **Winget distribution**: mcp-assert now available via `winget install BlackwellSystems.mcp-assert`. Auto-updated on each release.
- **Download stats SVG**: daily auto-updated card on README showing pip, npm, pytest plugin, brew, and GitHub release downloads.
- **521 total assertions** across 50 servers, 55 suites, 6 languages, 3 transports. 20 bugs found across 9 servers, 6 fix PRs submitted.

### Fixed

- **Hardcoded paths in 97 YAML files**: replaced absolute paths (`/Users/.../uvx`, `/Users/.../npx`) with bare commands (`uvx`, `npx`). Fixed CI failures in e2e-sqlite and e2e-memory jobs.
- **arxiv isError assertion**: re-skipped after maintainer's fix (#95) didn't resolve the issue in published v0.5.0.

## [0.5.0] - 2026-04-25

### Added

- **grafana/mcp-grafana suite**: 10 assertions, 1 bug found. `get_assertions` returns internal error (-32603) instead of `isError:true` on invalid timestamps. Issue filed ([#792](https://github.com/grafana/mcp-grafana/issues/792)), fix PR submitted ([#793](https://github.com/grafana/mcp-grafana/pull/793)).
- **microsoft/playwright-mcp suite**: 10 assertions, 100% clean. Navigate, snapshot, screenshot, JS evaluate, console, network, resize, close, invalid URL rejection, empty page handling. 31K-star server.
- **openai/sample-deep-research-mcp suite**: 4 assertions, 100% clean. OpenAI's official sample MCP server.
- **@google-cloud/storage-mcp suite**: 6 assertions, 100% clean. Google Cloud's official Storage MCP server.
- **Anthropic time/fetch/git suites**: 15 assertions across 3 more Anthropic official servers. All clean. All 7 Anthropic official servers now tested.
- **blazickjp/arxiv-mcp-server suite**: 5 assertions, 1 bug found. `get_abstract` returns error content but `isError` flag not set. Issue filed ([#92](https://github.com/blazickjp/arxiv-mcp-server/issues/92)), fix PR submitted ([#93](https://github.com/blazickjp/arxiv-mcp-server/pull/93)).
- **Badge snippet on passing runs**: `mcp-assert run` and `mcp-assert ci` print a ready-to-paste badge markdown snippet when all assertions pass.
- **GitHub Action `badge_markdown` output**: set when all assertions pass, ready for PR comments or README update workflows.
- **Fix PRs submitted**: mark3labs/mcp-go [#828](https://github.com/mark3labs/mcp-go/pull/828) (stderr hooks), antvis/mcp-server-chart [#292](https://github.com/antvis/mcp-server-chart/pull/292) (isError on chart failures), grafana/mcp-grafana [#793](https://github.com/grafana/mcp-grafana/pull/793) (timestamp validation), blazickjp/arxiv-mcp-server [#93](https://github.com/blazickjp/arxiv-mcp-server/pull/93) (isError flag).
- **git-mcp suite** (onmyway133): 7 assertions, 100% clean. Status, log, branches, diff, show, reflog, invalid repo rejection.
- **pytest-mcp-assert plugin**: pytest plugin that runs YAML assertions as pytest test items. `pip install pytest-mcp-assert`, then `pytest --mcp-suite evals/`. Each YAML file becomes a pytest Item with pass/fail/skip semantics. 170 lines of Python, zero MCP logic; the Go binary does all the work.
- **386 total assertions** across 30 servers, 39 suites, 5 languages, 3 transports. 14 bugs found, 5 issues filed, 5 fix PRs (4 ours). All 4 major tech companies covered (Anthropic, Google, OpenAI, Microsoft).

### Fixed

- **SSE Start() test expectations**: updated createMCPClient tests to match SSE eager-connect vs HTTP lazy-connect behavior.
- **git-mcp false positive**: our assertions used `repo_path` but the server's schema uses `cwd`. The server silently ignored the unknown param. Closed false positive issue immediately.

## [0.4.0] - 2026-04-25

### Added

- **Spring AI MCP server suite (Kotlin)**: 3 assertions, 100% tool coverage (2/2). First JVM language in the suite collection. Uses HTTP transport. Clean scan on `jamesward/hello-spring-mcp-server`.
- **MongoDB MCP server suite**: 4 assertions for the official MongoDB MCP server (1K stars). Knowledge search, error handling. Clean scan. Exemplary error messages with LLM-aware guidance.
- **3 new assertion types**: `file_not_contains` (file must not contain string), `file_not_exists` (file must not exist), `contains_any` (response contains at least one of the listed strings). Total assertion types: 18 + 4 trajectory = 22.
- **Condition registry refactor**: Checker internals refactored from if/else chain to ordered registry pattern. Adding a new assertion type is now a one-line registration.
- **303 total assertions** across 27 suites, 19 servers, 5 languages (Go, TypeScript, Python, Rust, Kotlin/Java).
- **HTTP/SSE transport support in `generate` command**: `--transport http|sse` and `--headers` flags for generating assertions against remote MCP servers. Reuses `createMCPClient` for transport selection.
- **Release procedure documentation**: step-by-step guide in `docs/distribution.md` covering the full release pipeline (GoReleaser, npm, PyPI, Homebrew, Scoop), verification commands, required secrets, rollback procedure, and GitHub Action maintenance. A new maintainer can cut a release without tribal knowledge.
- **GitHub Action maintenance guide**: how the action repo stays in sync with main project, when to update it, floating `v1` tag convention.
- **FastMCP SSE transport suite**: 11 assertions against the official FastMCP testing_demo server over SSE transport. Same server that passes 16/16 over stdio now verified over SSE. First SSE transport coverage in the suite collection. All pass.

### Fixed

- **SSE and HTTP transports missing `Start()` call**: `createMCPClient` returned SSE and HTTP clients without calling `Start()`, causing "transport not started yet" on `Initialize()`. The mcp-go SDK requires `Start()` before `Initialize()` for all transport types; our code only did this for stdio. Found by dogfooding: created SSE assertions, immediately hit the bug.
- **Distribution doc date typos**: fixed "3036-04-24" to "2026-04-24" in content posting dates.

## [0.3.0] - 2026-04-24

### Added

- **Custom SVG badges**: Three branded badge variants (passing, score, failing) with checkbox icon matching the mcp-assert logo. [Badge guide](https://blackwell-systems.github.io/mcp-assert/badge/).
- **github-mcp-server suite expanded**: 6 to 20 assertions covering 17 read-only tools across 7 toolsets (context, repos, git, issues, pull requests, users, gists). 
- **rmcp (Rust MCP SDK) suite**: 14 assertions, 100% tool coverage (6/6 tools + resources + prompts). First Rust MCP server in the example suite collection. Found bug: `get_value` decrements counter instead of reading it (repo archived).
- **rust-mcp-filesystem suite**: 23 assertions, 92% coverage (22/24 tools). Read, list, search, write, edit, zip/unzip, head, tail, line ranges, path traversal rejection. Clean scan on `rust-mcp-stack/rust-mcp-filesystem` (145 stars).
- **excel-mcp-server suite**: 15 assertions covering workbook creation, data round-trip, formulas, charts, pivot tables, formatting, merge cells, validation. Clean scan on `haris-musa/excel-mcp-server` (3,750 stars).
- **VHS demo GIF**: Terminal recording of the generate-and-run flow (Catppuccin Mocha theme, colorful window bar). Embedded in README.
- **Dogfooding report: rmcp**: `docs/dogfooding-rmcp.md` documenting the Rust SDK testing experience and bug found.
- **antvis/mcp-server-chart suite**: 16 assertions for the AntV charting server (4K stars). Found 9 tools that crash with unhandled JavaScript exceptions on default input. Filed [antvis/mcp-server-chart#291](https://github.com/antvis/mcp-server-chart/issues/291).
- **Architecture doc rewrite**: Comprehensive 600-line guide covering MCP primer, assertion lifecycle, package structure, transport layer, fixture isolation, block types, reporting, and extension points.
- **Star CTA**: Dark/light mode star call-to-action image in README footer.
- **MCP everything server (TypeScript) suite**: 13 assertions (12 pass, 1 skip) for the official Anthropic reference server. Clean scan.
- **Terraform MCP server suite**: 5 assertions for HashiCorp's terraform-mcp-server (1.3K stars). Provider lookup, module search, policy search. Clean scan.
- **Notion MCP server suite**: 22 assertions, 100% tool coverage (22/22) for the official Notion MCP server (4.2K stars). Clean scan.
- **Custom HTTP/SSE headers**: New `headers` field on server config for authenticated remote MCP servers. Values support `${VAR}` expansion. Enables testing Cloudflare, Notion (with tokens), and any API-key-protected server.
- **pytest plugin on roadmap**: Framework integration layer planned. `pip install pytest-mcp-assert`, each YAML assertion becomes a pytest test case.
- **296 total assertions** across 25 suites, 17 servers, 4 languages (Go, TypeScript, Python, Rust). 12 bugs found, 3 upstream issues filed.

### Fixed

- **`skip` field now works.** The `skip: true` field was defined in `types.go` and set by the `generate` command on destructive tools, but the runner never checked it. Assertions marked `skip: true` were silently running instead of being skipped. Now they return `SKIP` immediately without starting a server.

## [0.2.4] - 2026-04-24

### Added

- **npm distribution**: `npx @blackwell-systems/mcp-assert`. Platform-specific optional dependencies (same pattern as esbuild). No Go toolchain required. 7 packages (1 root + 6 platform) auto-published on each release tag.
- **PyPI distribution**: `pip install mcp-assert`. Platform-specific wheels with Go binary embedded. Tagged with correct platform (macosx, manylinux, win). Auto-published via twine on each release tag.
- **Scoop distribution**: `scoop install mcp-assert` (Windows). Manifest auto-generated by GoReleaser and pushed to [blackwell-systems/scoop-bucket](https://github.com/blackwell-systems/scoop-bucket).
- **Homebrew distribution**: `brew install blackwell-systems/tap/mcp-assert`. Formula auto-generated by GoReleaser on each release.
- **GitHub Marketplace**: [mcp-assert-action](https://github.com/marketplace/actions/mcp-assert) listed on GitHub Marketplace.
- **"Works with mcp-assert" badge**: Static and dynamic (CI-verified) badge variants for MCP server READMEs. [Badge guide](https://blackwell-systems.github.io/mcp-assert/badge/). Static badge via shields.io, dynamic badge via `--badge` endpoint JSON + GitHub Pages.
- **Social preview**: GitHub social preview image for the repository.
- **GitHub topics**: mcp, testing, mcp-server, assertions, golang, cli, yaml, github-actions, mcp-testing, snapshot-testing, regression-testing.

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
