# New User Comprehensibility Audit

**Date:** 2026-04-23
**Auditor:** Simulated new user (developer, TypeScript/Go, MCP server author)

## Overall Score: 8.5/10

The documentation is comprehensive, well-structured, and technically accurate. A new user with MCP experience can go from zero to running assertions in under 5 minutes. However, there are several gaps that would cause confusion, especially around the recommended onboarding path and the relationship between commands.

---

## Path 1: README -> Getting Started

### Strengths

1. **Immediate value proposition**: The opening line ("The testing standard for deterministic MCP tools") establishes positioning in 3 seconds. The scope statement ("Works with any language, any transport") addresses the first question a polyglot MCP developer would ask.

2. **Accurate Quick Start**: The three-command quick start works as advertised:
   ```bash
   go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest
   mcp-assert init evals
   mcp-assert run --suite evals/ --fixture evals/fixtures
   ```
   All commands exist, flags are correct, and the path creates a runnable assertion.

3. **Clear positioning vs LLM-as-judge**: The "When to use what" table immediately answers a question every MCP server author would have. This is excellent. The table format makes it scannable.

4. **Zero-Effort Coverage is compelling**: The generate/snapshot/run flow is a legitimate 30-second path from "I want tests" to "I have tests." This should be the hero feature.

5. **Getting Started doc structure**: The progression from template scaffold → manual assertion → server-based generation → zero-effort coverage is logical. Each step builds on the previous one.

6. **Accurate examples**: The YAML examples in `getting-started.md` match the actual schema and work when copied. The `{{fixture}}` substitution is explained early and consistently.

### Gaps

1. **Quick Start in README omits fixtures**: The Quick Start shows:
   ```bash
   mcp-assert init evals
   mcp-assert run --suite evals/ --fixture evals/fixtures
   ```
   But `init` (without `--server`) creates `evals/fixtures/hello.txt` automatically. A new user might think they need to create `evals/fixtures` themselves. This isn't a blocker (the command will succeed), but it's a minor source of confusion.

2. **Two "recommended" paths with no guidance**: The README Quick Start shows the template path (`init evals`), while Getting Started §1 ("One-step suite generation (recommended)") shows the server-based path (`init evals --server`). Both are labeled as the default or recommended path. A new user doesn't know which to choose. The README should say "If you have a running server, use `--server`; otherwise, start with the template."

3. **Missing context on when to write assertions manually**: Getting Started §3 ("Write an assertion by hand") doesn't explain when you'd choose this over `init` or `generate`. The answer: when you know your server config and want more control than the template provides, but don't want to run `generate`. This is a valid path but needs a one-sentence "when to use" callout.

4. **"Zero-Effort Coverage" appears twice**: It's in the README (lines 67-78) and Getting Started (lines 121-147). The duplication isn't harmful, but the README version is more polished (includes expected output). The Getting Started version adds the individual-step breakdown, which is useful. Consider: README gets the one-liner version, Getting Started gets the detailed version with rationale.

5. **No mention of `--fix` in Quick Start paths**: The `--fix` flag is documented in cli.md and writing-assertions.md, but never surfaces in the initial "your first run" flow. A new user will hit position-sensitive failures (especially in LSP servers) and won't know `--fix` exists. Add a one-line callout in Getting Started §2 after the first `run` example: "If a position-sensitive assertion fails, pass `--fix` to get a suggested correction."

6. **Fixture directory ambiguity**: Getting Started line 14 says `init` with `--server` accepts `--fixture ./fixtures`, but the example output (lines 19-39) doesn't show where fixtures are created or expected. Does `--server` mode create a fixture directory? Does it require one to exist? The answer (from code review): it doesn't create fixtures; it uses `--fixture` only for `{{fixture}}` substitution in generated YAMLs. If `--fixture` is omitted, `{{fixture}}` appears literally in the generated files. This should be clarified.

### Suggestions

1. Add a one-sentence "which path should I use?" decision tree at the top of Getting Started:
   > **Which path should I use?**
   > - Have a running MCP server? → `init evals --server "my-server"` (generates stubs + baselines)
   > - Want to start from a template? → `init evals` (creates one commented YAML)
   > - Know your server config already? → Write the YAML directly (§3)

2. In the README Quick Start, add one line clarifying fixture creation:
   ```bash
   mcp-assert init evals  # Creates evals/read_file.yaml and evals/fixtures/hello.txt
   ```

3. Consolidate "Zero-Effort Coverage": README keeps the polished one-liner version with output, Getting Started keeps the detailed breakdown. Add a cross-reference in the README: "See [Getting Started](link) for step-by-step details."

4. Add `--fix` to the "Next steps" output of `init`. Currently it prints:
   ```
   Next steps:
     Run the suite:   mcp-assert run --suite evals --server "my-server"
   ```
   Should also print:
   ```
     Fix position failures: mcp-assert run --suite evals --server "my-server" --fix
   ```

---

## Path 2: CI Integration

### Strengths

1. **One-liner GitHub Action**: The example at the top (lines 6-9) is copy-paste ready and works. The dedicated action (`blackwell-systems/mcp-assert-action@v1`) is mentioned but not required, which is correct.

2. **All flags documented**: `--threshold`, `--fail-on-regression`, `--baseline`, `--save-baseline`, `--junit`, `--markdown`, `--badge` are all explained with examples.

3. **Regression detection is clear**: The baseline/compare flow (lines 61-71) explains exactly what counts as a regression ("PASS to FAIL") and what doesn't ("new tests that fail"). This is critical for CI adoption.

### Gaps

1. **Auto-detection of `$GITHUB_STEP_SUMMARY` is mentioned but not explained**: Line 44 says "`ci` mode auto-detects `$GITHUB_STEP_SUMMARY` for markdown output," but doesn't say what happens (markdown is written there automatically) or how to override it. A new user might wonder if they need `--markdown` in GitHub Actions. The answer: no, it's automatic in `ci` mode. Add one sentence: "In `ci` mode on GitHub Actions, markdown is written to `$GITHUB_STEP_SUMMARY` automatically (no `--markdown` flag needed)."

2. **No example of `--fail-on-regression` with `--threshold`**: What happens if you set both? Does a regression fail the build even if the threshold is met? The answer (from cli.md and typical CI semantics): both are checked independently; either can fail the build. Clarify: "You can combine `--threshold` and `--fail-on-regression`. Both are checked; either can cause a failure."

3. **Badge example is incomplete**: Line 56 shows the badge output format but doesn't explain where to host `badge.json` or how to use the URL. A new user would need to look up shields.io endpoint syntax. Add one sentence: "Host `badge.json` at a public URL (GitHub Pages, GitHub releases, or a CDN), then use `https://img.shields.io/endpoint?url=<badge-url>`."

### Suggestions

1. Add a "Typical CI workflow" section that combines all the pieces:
   ```yaml
   - name: Run assertions
     run: |
       go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest
       mcp-assert ci --suite evals/ --threshold 95 --baseline baseline.json --fail-on-regression --junit results.xml
   
   - name: Upload results
     if: always()
     uses: actions/upload-artifact@v4
     with:
       name: test-results
       path: results.xml
   ```
   This shows how all the flags work together in a realistic workflow.

2. Clarify the `--markdown` behavior in GitHub Actions (see Gap #1 above).

---

## Path 3: CLI Reference

### Strengths

1. **Command table is accurate**: All 9 commands (init, run, ci, matrix, coverage, generate, snapshot, watch, intercept, version) are listed. I verified these against `main.go` (lines 18-66) and all exist.

2. **Flag tables are complete**: Checked `run` flags against `commands.go` and all documented flags exist. The `--fix` flag is documented in both `run` and `ci` (lines 61, 78), which is correct.

3. **`intercept` is documented**: Lines 178-192 explain the proxy behavior, `--trajectory` flag, and use case. This is a non-obvious command and the explanation is clear.

4. **"YAML-level feature" callouts**: Lines 245-339 explain which features (client_capabilities, assert_resources, assert_prompts, trajectory, progress, transport) are YAML-only and have no CLI equivalents. This prevents a new user from searching for a `--client-capabilities` flag that doesn't exist.

5. **Docker isolation caveats**: Line 241 notes Docker only works with stdio transport, not HTTP/SSE. This is critical for HTTP server authors to know.

### Gaps

1. **`--fix` is not in the command usage summary**: Lines 6-14 show the usage for each command, but `--fix` is never listed. It appears in the flag table (line 61) and the description (line 78), but not in the `run` or `ci` usage lines. The usage should show:
   ```
   mcp-assert run --suite <path> [--fix] [flags]
   mcp-assert ci  --suite <path> [--fix] [flags]
   ```
   (This is consistent with the main.go printUsage output, which also omits `--fix`. It should be added there too.)

2. **`intercept --trajectory` is listed as required but description says "optional"**: Line 14 shows `--trajectory <path>` with no brackets (implying required), but line 188 in the description says "YAML file containing trajectory assertions to validate on disconnect" (no "required" marker). The code (intercept.go line 22) shows it IS required. The description should say "required" explicitly.

3. **No explanation of what happens when `--server` CLI flag overrides YAML `server:` block**: Line 52 says "`--server` overrides server command from CLI instead of per-YAML" but doesn't explain what happens to `server.args` or `server.env` from the YAML. Does the CLI override everything, or just `command`? The answer (from code review): CLI `--server` replaces the entire server block (command + args). The YAML `env` is ignored. This should be clarified with an example.

4. **`--interval` for watch mode is listed but not explained**: Line 171 shows `--interval <duration>` with default `2s`, but doesn't say what it does (polling interval for file changes). This is mentioned in the intro (line 173: "Polls for changes"), but the flag description should be explicit: "How often to check for YAML file changes (default: 2s)."

5. **Server Override section (lines 206-211) duplicates `--server` flag description**: This section repeats information from the flag table. It adds a concrete example, which is useful, but the duplication is noticeable. Consider: move the example up to the flag table, or frame this section as "Example: Override server config for an entire suite."

6. **Reliability metrics example is excellent but misplaced**: Lines 342-366 show a multi-trial run with pass@k/pass^k output. This is one of the best examples in the entire doc. However, it's at the end of a long reference doc. Consider: move this to a "Key Features" section near the top, or add a callout in the `--trials` flag description (line 53) pointing to the example.

### Suggestions

1. Add `[--fix]` to the `run` and `ci` usage lines (lines 7, 8) and the main.go printUsage output.

2. Mark `--trajectory` as required in the flag table (line 188): "Path to YAML file with trajectory assertions (required)."

3. Expand the `--server` flag description (line 52) with a behavioral note:
   > Override server command for all assertions. Replaces the entire `server:` block in each YAML (command + args). The YAML's `env:` block is preserved. Example: `--server "agent-lsp go:gopls"`.

4. Expand `--interval` description (line 171):
   > How often to poll the `--suite` directory for YAML file changes (default: 2s).

5. Add a "Key Features & Examples" section at the top of cli.md, before the command table, highlighting reliability metrics, Docker isolation, and fix mode with their example outputs. Then link to them from the flag descriptions.

---

## Path 4: Examples

### Strengths

1. **Coverage is excellent**: 18 suites (17 server + 1 trajectory), 12 different servers, 3 languages, 174 assertions. This is the best examples section I've seen in any testing tool.

2. **Each example has clear setup instructions**: Every suite shows the install command (npm, uvx, git clone) and the `mcp-assert run` invocation. A new user can copy-paste and run any example in under 60 seconds.

3. **Annotations explain what's being tested**: Each suite description lists the tools/features covered and notes coverage percentage ("92% tool coverage (13/14 tools)"). This is extremely helpful for understanding what a comprehensive test suite looks like.

4. **Trajectory suite is well-explained**: Lines 215-236 explain that trajectory assertions validate skill protocols, run without a server, and can use inline traces or audit logs. The table showing constraints per skill is a great reference.

5. **Real-world servers**: filesystem, memory, sqlite, fastmcp, agent-lsp, mcp-go, github-mcp. These are the servers people actually use. The examples aren't toy demos.

6. **Transport diversity**: stdio (most suites), HTTP (mcp-go-everything-http). This shows the tool works with both transports.

### Gaps

1. **fastmcp note about `/tmp/fastmcp` is easy to miss**: Lines 75-79 have a note saying you need to clone the fastmcp repo to `/tmp/fastmcp` before running. This is critical (the assertions will fail without it), but it's in a note block that could be skipped. The "install" line should be:
   ```bash
   git clone --depth 1 https://github.com/PrefectHQ/fastmcp.git /tmp/fastmcp
   mcp-assert run --suite examples/fastmcp-testing-demo
   ```
   (This is shown in lines 75-76, so not wrong, just easy to miss.)

2. **agent-lsp example requires fixtures but doesn't show where to get them**: Line 88 shows `--fixture /path/to/go/fixtures` but doesn't say where these fixtures are or how to create them. Are they in the agent-lsp repo? Should the user create them? The answer (from code review): the agent-lsp repo includes test fixtures in `test/fixtures/go`. This should be stated: "Use the fixtures from the agent-lsp repo: `--fixture /path/to/agent-lsp/test/fixtures/go`."

3. **No example showing how to adapt an example to your own server**: A new user might think they need to copy the entire `examples/filesystem/` directory and modify every YAML. They don't—they can just run `generate` or `init --server`. But the examples doc never says this. Add a "Adapting Examples" section at the end: "To test your own server, use `mcp-assert generate --server "your-server" --output evals/` instead of copying these examples."

4. **Summary table doesn't show which examples use advanced features**: The table (lines 7-26) shows server, language, transport, and assertion count, but not which examples demonstrate setup steps, capture, client_capabilities, trajectory, etc. A new user looking for "an example with setup steps" would need to read every description. Add columns for "Key Features" showing: setup, capture, client_capabilities, negative tests, stateful, trajectory.

5. **mcp-go longRunningOperation is skipped but reason isn't in the main examples table**: Line 95 mentions the known bug (stdio transport issue) but this isn't visible in the summary table. If a new user runs `mcp-assert run --suite examples/mcp-go-everything`, they'll see a SKIP and wonder why. Add a note in the suite description: "One assertion (longRunningOperation) is skipped due to a known mcp-go stdio bug."

6. **GitHub MCP Server example requires a token but doesn't explain how to get one**: Line 207 shows `GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_TOKEN mcp-assert run ...` but doesn't explain what scopes the token needs or where to create it. Add one sentence: "Create a token at https://github.com/settings/tokens with `repo` and `read:user` scopes."

### Suggestions

1. Add a "Key Features" column to the summary table showing which suites demonstrate: setup, capture, client_capabilities, negative tests, stateful, trajectory, multi-file, auth.

2. Expand agent-lsp fixture note (line 88):
   ```bash
   # Clone agent-lsp if not already present
   git clone https://github.com/blackwell-systems/agent-lsp.git /tmp/agent-lsp
   mcp-assert run --suite examples/agent-lsp-go --fixture /tmp/agent-lsp/test/fixtures/go
   ```

3. Add an "Adapting Examples to Your Server" section at the end:
   > **Adapting Examples to Your Server**
   >
   > These examples are for reference. To test your own MCP server, use:
   > ```bash
   > mcp-assert generate --server "your-mcp-server" --output evals/
   > ```
   > This queries your server's `tools/list`, generates stub YAMLs, and captures baselines automatically.

4. Add GitHub token scope note (line 207):
   ```bash
   # Create a token at https://github.com/settings/tokens with repo + read:user scopes
   GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_TOKEN mcp-assert run --suite examples/github-mcp
   ```

---

## Critical Gaps (would block a new user)

### 1. No guidance on choosing between `init` (template) and `init --server` (generate)
**Impact:** A new user reading the README will use `init evals`, then read Getting Started and see `init evals --server` described as "recommended." They'll wonder if they took the wrong path.

**Fix:** Add a decision tree at the top of Getting Started (see Path 1 suggestions above).

### 2. `--fixture` behavior with `init --server` is unclear
**Impact:** A new user running `init evals --server "my-server"` without `--fixture` will get YAMLs with literal `{{fixture}}` strings. They won't know this is wrong until they run the assertions and get cryptic errors.

**Fix:** Document in Getting Started §1 that `--fixture` is optional but recommended: "If omitted, `{{fixture}}` appears literally in generated YAMLs. Pass `--fixture ./fixtures` to substitute real paths."

### 3. No mention of `--fix` in initial onboarding
**Impact:** A new user writing LSP server assertions will hit "no identifier found at line X, column Y" errors and won't know how to fix them. They'll manually adjust line/column values, which is tedious and error-prone.

**Fix:** Add `--fix` to the "Next steps" output of `init` and mention it in Getting Started §2 after the first `run` command.

### 4. agent-lsp and fastmcp examples require external fixtures/repos but this isn't obvious
**Impact:** A new user running `mcp-assert run --suite examples/agent-lsp-go --fixture /path/to/go/fixtures` will get a "directory not found" error if they don't have the agent-lsp repo cloned.

**Fix:** Expand the setup instructions for agent-lsp and fastmcp to show the `git clone` step explicitly (see Path 4 suggestions above).

---

## Nice-to-haves (would improve experience)

### 1. Consolidate "Zero-Effort Coverage" duplication
The README and Getting Started both explain this flow. The duplication isn't harmful, but it's noticeable. Consolidate as suggested in Path 1.

### 2. Add "Key Features" to examples summary table
Helps new users find relevant examples faster. See Path 4, Gap #4.

### 3. Move reliability metrics example higher in cli.md
The pass@k/pass^k example (lines 342-366) is one of the best features but it's buried at the end. Surface it earlier.

### 4. Explain `--server` CLI override behavior more clearly
What happens to `args` and `env` when you override from CLI? See Path 3, Gap #3.

### 5. Add a "Common Workflows" section to cli.md
Show how flags combine in realistic scenarios:
- One-shot local run: `run --suite evals/ --fixture ./fixtures`
- CI with threshold: `ci --suite evals/ --threshold 95 --junit results.xml`
- Regression detection: `ci --suite evals/ --baseline baseline.json --fail-on-regression`
- Multi-trial reliability: `run --suite evals/ --trials 5 --json`
- Position fix: `run --suite evals/ --fix`

### 6. Add shields.io badge hosting example
The `--badge` flag is documented but doesn't explain where to host the JSON or how to use the URL. See Path 2, Gap #3.

### 7. Clarify `$GITHUB_STEP_SUMMARY` auto-detection
Mention it in ci-integration.md: "In `ci` mode on GitHub Actions, markdown is written to `$GITHUB_STEP_SUMMARY` automatically."

### 8. Add a "Troubleshooting" section
Common issues a new user might hit:
- "no identifier found at line X" → use `--fix`
- "{{fixture}} not substituted" → pass `--fixture` flag
- "SKIP: destructive tool" → edit the YAML and remove `skip: true`
- "timeout: context deadline exceeded" → increase `timeout:` in YAML

---

## Summary

The documentation is **excellent** for a new testing tool. A motivated developer can go from "never heard of mcp-assert" to "running assertions on my server" in under 10 minutes. The writing is clear, examples are realistic, and technical accuracy is high.

The main friction points are:

1. **Two "recommended" onboarding paths with no decision guidance** (template vs generate)
2. **`--fix` is never surfaced in initial onboarding** (will cause pain for LSP server authors)
3. **Fixture behavior with `init --server` is ambiguous** (what happens if you omit `--fixture`?)
4. **Examples requiring external repos don't show the full setup** (agent-lsp, fastmcp)

Fixing these four issues would eliminate all blocking friction. The nice-to-haves are polish: consolidating duplication, surfacing advanced features earlier, and adding troubleshooting guidance.

**Recommendation:** Fix the four critical gaps before the next release. The nice-to-haves can wait. Even without them, this is already better than 90% of OSS testing tools.
