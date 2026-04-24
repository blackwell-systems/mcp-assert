# Dogfooding: GitHub MCP Server

**Date:** 2026-04-24
**Server:** [github/github-mcp-server](https://github.com/github/github-mcp-server) (28K+ stars, Go)
**Goal:** Test mcp-assert's onboarding experience against the most popular MCP server.
**Method:** Follow the "new server author" path end-to-end, documenting friction and findings.

## Results Summary

| Metric | Value |
|--------|-------|
| Tools discovered | 38 |
| Assertions written | 6 (targeted read-only) |
| Passing | 6/6 |
| Time to first pass | ~15 minutes |
| Friction points found | 4 |
| Server behavior findings | 1 |

## Timeline

### Step 1: Clone and build (2 min)

```bash
git clone --depth 1 https://github.com/github/github-mcp-server.git
cd github-mcp-server && go build -o github-mcp-server ./cmd/github-mcp-server
```

Zero friction. Standard Go project, builds clean.

### Step 2: Generate assertion stubs (5 sec)

```bash
mcp-assert generate \
  --server "./github-mcp-server stdio" \
  --output examples/github-mcp
```

**38 tools discovered, 38 stubs created.** The `generate` command worked perfectly. Clear "Next steps" output. This is the golden path working as designed.

### Step 3: Fill in values and run (10 min)

Wrote 6 targeted assertions for read-only tools: `get_me`, `search_repositories`, `get_file_contents`, `list_issues`, `search_code`, `list_branches`. All targeting `blackwell-systems/mcp-assert` and `blackwell-systems/agent-lsp`.

**6/6 passing** after fixing the env var issue (see friction #2 below).

### Step 4: Snapshot capture

```bash
mcp-assert snapshot --suite examples/github-mcp --update --timeout 30s
```

Captured baselines for all 6 assertions. Future runs will detect response changes.

## Friction Points

### F1: Auth token env var not documented in generate output

The generated stubs don't mention that the server needs `GITHUB_PERSONAL_ACCESS_TOKEN`. A user who runs `mcp-assert run` immediately after `generate` gets "transport closed" with no hint about auth. The `generate` command could detect env vars the server reads (from its `--help` or config) and include them in the stub comments.

**Severity:** Medium. Costs 5 minutes of debugging for every new authenticated server.

### F2: Shell variable expansion not supported in YAML `env:` blocks

Writing `GITHUB_PERSONAL_ACCESS_TOKEN: "${GITHUB_TOKEN}"` in the YAML `env:` block passed the literal string `${GITHUB_TOKEN}` to the subprocess.

**Status: Fixed.** `${VAR}` and `$VAR` expansion in YAML `env:` blocks is now supported. Variables are resolved from the host environment at runtime. If the variable is not set, the original string is preserved unchanged.

### F3: `--suite` doesn't accept a single file, only directories

`mcp-assert run --suite path/to/single.yaml` previously failed with "not a directory".

**Status: Fixed.** `--suite` now accepts both directories and single YAML files.

## Server Behavior Finding

### get_file_contents returns confirmation, not content

`get_file_contents` with `owner: blackwell-systems, repo: mcp-assert, path: README.md` returns:

```
successfully downloaded text file (SHA: 196789708ac4e498a94486d8d0dc98726b3baf9d)
```

The actual file content is not in the MCP text response. It may be in a separate content block (resource type). This means agents calling `get_file_contents` and reading the text response won't see the file. Worth investigating whether this is intentional (content in a different MCP content block) or a bug (content should be in text).

**Status:** Needs investigation. Not yet filed.

### F4: Generated stubs execute write operations with placeholder values

Running the full generated suite against the GitHub MCP server with a valid token caused `create_repository` to create a real repo called `TODO_NAME` on GitHub. Other write stubs (`delete_file`, `push_files`, `merge_pull_request`) failed only because their TODO args were invalid, not because anything prevented them from running.

**Severity:** Critical. A user following the `generate` then `run` path will execute destructive operations against real infrastructure. The `generate` command should either skip tools with `destructiveHint: true` in their annotations, mark write operations as `skip: true` in the generated YAML, or require a `--include-writes` flag to opt in.

**Workaround:** Manually delete stubs for write operations before running the suite.

## Onboarding Improvements Identified

1. ~~**`mcp-assert init --server`**: One command to generate + snapshot + CI template.~~ **Shipped.** `mcp-assert init evals --server "cmd" --fixture ./fixtures` collapses generate + snapshot into one step.
2. ~~**Env var expansion**: `${VAR}` in YAML env blocks should resolve from the shell environment.~~ **Shipped.** `${VAR}` and `$VAR` expansion now works in `env:` blocks.
3. ~~**Single-file `--suite`**: Accept `--suite path/to/file.yaml` for iterative development.~~ **Shipped.** `--suite` accepts both directories and single YAML files.
4. **Auth detection in `generate`**: When the server exits immediately (transport closed), suggest checking for required env vars. (Not yet implemented.)
5. ~~**Skip destructive stubs by default**: `generate` should check `destructiveHint` from tool annotations and set `skip: true` on write operations. Opt in with `--include-writes`.~~ **Shipped.** Destructive tools are generated with `skip: true` by default; use `--include-writes` to include them.
