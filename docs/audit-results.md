# Documentation Audit Results

Audit performed 2026-04-23. Reviewed all 13 documentation files from the perspective of a new user discovering mcp-assert for the first time.

---

## Issues Found and Fixed

### 1. Stale assertion counts (FEATURES.md, examples.md)

**Files:** `FEATURES.md`, `docs/examples.md`

agent-lsp example suite has 60 assertions (verified by file count), but documentation said 51 in multiple places. Total assertion count said 76 instead of 85 (14 + 5 + 6 + 60 = 85).

**Fixed:**
- `FEATURES.md`: Updated "76 assertions" to "85 assertions" in section header
- `FEATURES.md`: Updated agent-lsp row from 51 to 60
- `FEATURES.md`: Updated CI pipeline table from "51 assertions" to "60 assertions"
- `docs/examples.md`: Updated agent-lsp description from "51 assertions" to "60 assertions"

### 2. Stale unit test breakdown (FEATURES.md)

**File:** `FEATURES.md`

Unit test breakdown listed `internal/runner` as 31 tests. Actual count is 42 (verified via `go test -v`). The total of 100 was correct but the per-package breakdown was wrong (22 + 36 + 31 = 89, not 100).

**Fixed:** Updated runner row from 31 to 42 and added "capture/extractJSONPath" to the test description to reflect the added tests.

### 3. Contradictory roadmap for capture feature (roadmap.md)

**File:** `docs/roadmap.md`

The "Setup output capture" table row correctly says "Shipped", but the detail section below still reads as if it were unimplemented: "The root limitation: mcp-assert can't chain outputs...", "Current workaround:", "Proposed syntax:", "Why it's high priority: 7 of agent-lsp's 50 tools (14%) are currently tested only as negative tests because of this limitation."

A new user reading the roadmap would be confused about whether capture actually works.

**Fixed:** Rewrote the detail section to describe capture as a shipped feature, removed "Current workaround" / "Proposed syntax" / "Implementation" framing, and added a cross-link to the writing-assertions docs.

### 4. Missing CLI flags in cli.md

**File:** `docs/cli.md`

Three items present in the actual CLI (`main.go` printUsage) but absent from the docs:

- `--timeout <duration>` on the `run` command (default 30s)
- `--interval <duration>` on the `watch` command (default 2s)
- `version` / `--version` command

**Fixed:** Added `--timeout` to the `run` flags table, added `--interval` to the `watch` section with its own flags table, and added a `mcp-assert version` command section.

### 5. Missing doc links in README.md

**File:** `README.md`

The Documentation section links to 6 of 8 docs site pages. Roadmap and Dogfooding are in the mkdocs nav but not linked from the README.

**Fixed:** Added Roadmap and Dogfooding links to the Documentation section.

### 6. Wrong assertion type count in CHANGELOG (CHANGELOG.md)

**File:** `CHANGELOG.md`

v0.1.0 entry says "13 deterministic assertion types" but then lists 14 types (the list includes `in_order`).

**Fixed:** Changed "13" to "14".

### 7. Corrupted Unicode in architecture.md

**File:** `docs/architecture.md`

The package dependency graph tree-drawing characters had corrupted UTF-8 bytes (Unicode replacement characters U+FFFD) on the line referencing `mark3labs/mcp-go/client`. Rendered as `+-~~` instead of the intended box-drawing characters.

**Fixed:** Replaced corrupted byte sequence with correct UTF-8 box-drawing characters.

### 8. Misleading --fail-on-regression example (ci-integration.md)

**File:** `docs/ci-integration.md`

The "CI commands" section showed `--fail-on-regression` without `--baseline`, but the flag requires `--baseline` to function (as documented in cli.md).

**Fixed:** Added `--baseline baseline.json` to the example command and a parenthetical note.

### 9. Capture not shown in FEATURES.md YAML format example

**File:** `FEATURES.md`

The YAML Assertion Format section showed the full schema but omitted `capture:` on setup steps, even though capture is a shipped feature documented elsewhere.

**Fixed:** Added `capture:` field to the setup step in the YAML example.

---

## Verified Correct (No Changes Needed)

- **All 8 CLI commands documented in cli.md:** init, run, ci, matrix, coverage, generate, snapshot, watch (plus version, now added)
- **`capture` feature documented in writing-assertions.md:** Full section with example at lines 126-154
- **14 assertion types documented in writing-assertions.md:** All present with examples
- **Unit test count of 100:** Verified via `go test -v` (22 + 36 + 42 = 100)
- **GitHub Action README matches action.yml:** All 8 inputs and 3 outputs match between README table and action.yml definitions
- **No dead links to nonexistent docs pages:** All mkdocs nav entries correspond to existing files
- **`--coverage-json`, `--baseline`, `--save-baseline`, `--fail-on-regression`, `--docker` all documented in cli.md**
- **Code examples are copy-pasteable:** YAML examples use correct syntax, bash commands use correct flag names
- **docs/index.md and README.md content is consistent:** Both tell the same story with appropriate detail levels
