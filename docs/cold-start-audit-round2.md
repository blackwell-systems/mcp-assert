# Second-Round New User Audit

**Date:** 2026-04-23
**Scope:** Full repository audit after first-round fixes
**First-round score:** 8.5/10 (4 critical gaps fixed)

## Summary

| Severity | Count |
|----------|-------|
| UX-critical | 2 |
| UX-improvement | 5 |
| UX-polish | 3 |

**Total issues:** 10

## Verification: First-Round Fixes

All four first-round critical gaps have been successfully fixed:

1. Decision tree at top of getting-started.md: Present
2. `--fixture` note in one-step generation section: Present (line 22)
3. `--fix` tip after first run example: Present (line 78)
4. Explicit git clone steps in examples.md: Present for agent-lsp (line 89), fastmcp (line 76), and mentioned for github-mcp (line 214)

---

## Critical Issues

### [Examples] Hardcoded absolute paths in github-mcp YAMLs
- **Severity**: UX-critical
- **What happens**: All 6 github-mcp example YAMLs use `/Users/dayna.blackwell/code/github-mcp-server/github-mcp-server` as the command. New users cannot run these without editing every file.
- **Expected**: Either use a portable binary name (`github-mcp-server` in PATH), or document the required setup in examples.md and include a comment in each YAML pointing to the setup instructions.
- **Repro**: 
  ```bash
  cat examples/github-mcp/get_me.yaml
  # Shows: command: /Users/dayna.blackwell/code/github-mcp-server/github-mcp-server
  ```
- **Impact**: Affects 6 files: `get_me.yaml`, `list_branches.yaml`, `get_file_contents.yaml`, `search_repositories.yaml`, `search_code.yaml`, `list_issues.yaml`

### [Examples] Hardcoded /tmp/ paths in mcp-go example YAMLs
- **Severity**: UX-critical
- **What happens**: 11 mcp-go example YAMLs reference `/tmp/mcp-go-*-server*` binaries that don't exist on a new user's machine. Examples fail silently with "command not found" or "no such file or directory".
- **Expected**: Document the build process in examples.md for the mcp-go examples, or provide a setup script.
- **Repro**: 
  ```bash
  cat examples/mcp-go-sampling/ask_llm.yaml
  # Shows: command: /tmp/mcp-go-sampling-server-v049
  ```
- **Impact**: Affects 11 files across 6 suites (elicitation, sampling, roots, everything-prompts, everything-resources, everything-http comments)

---

## Improvement Issues

### [CLI] --fix flag missing from main.go usage output
- **Severity**: UX-improvement
- **What happens**: User runs `mcp-assert run --help` and doesn't see `--fix` as an available flag. The flag IS implemented and documented in cli.md, but the CLI usage text doesn't mention it.
- **Expected**: Add `--fix` to the usage text in `cmd/mcp-assert/main.go` under the `run` and `ci` command lines.
- **Repro**: 
  ```bash
  mcp-assert run --help | grep -c "\\-\\-fix"
  # Returns 0
  ```

### [Docs] cold-start-audit.md and audit-results.md not in mkdocs nav
- **Severity**: UX-improvement
- **What happens**: Two valuable audit documents exist in `docs/` but aren't linked in the mkdocs navigation. Users browsing the docs site won't discover them.
- **Expected**: Add both to the mkdocs.yml nav, perhaps under a section called "Audit Results" or "Quality".
- **Repro**: 
  ```bash
  grep -E "(cold-start-audit|audit-results)" mkdocs.yml
  # Returns nothing
  ```

### [Docs] No CONTRIBUTING.md file
- **Severity**: UX-improvement
- **What happens**: New contributors don't have guidance on how to add assertion types, CLI commands, or where to start.
- **Expected**: Create a CONTRIBUTING.md that explains the package structure, how to add a new assertion type (modify types.go + checker.go + add tests), and how to run tests locally.
- **Repro**: 
  ```bash
  ls -la CONTRIBUTING.md
  # File not found
  ```

### [Code] Missing package-level documentation for internal/assertion
- **Severity**: UX-improvement
- **What happens**: The `internal/assertion` package has no package-level doc comment explaining what it does or how it's structured. Contributors reading the code have to infer the architecture.
- **Expected**: Add a package doc comment at the top of `types.go` (the canonical types file) explaining the package's role and structure.
- **Repro**: 
  ```bash
  grep "^// Package assertion" internal/assertion/*.go
  # Returns nothing
  ```

### [Code] Missing package-level documentation for internal/report
- **Severity**: UX-improvement
- **What happens**: The `internal/report` package has no package-level doc comment explaining its role in output formatting and reporting.
- **Expected**: Add a package doc comment at the top of `report.go` explaining the package's role (terminal output, JUnit, markdown, badges, etc.).
- **Repro**: 
  ```bash
  grep "^// Package report" internal/report/*.go
  # Returns nothing
  ```

---

## Polish Issues

### [Examples] /tmp/fastmcp path in 16 fastmcp-testing-demo YAMLs
- **Severity**: UX-polish
- **What happens**: The fastmcp-testing-demo suite references `/tmp/fastmcp/examples/testing_demo/server.py`. This path is documented in examples.md (line 76), but a user who just browses the YAML files won't see the setup requirement.
- **Expected**: Add a comment at the top of each fastmcp YAML pointing to the setup instructions, or add a README.md in the `examples/fastmcp-testing-demo/` directory.
- **Repro**: 
  ```bash
  cat examples/fastmcp-testing-demo/add.yaml | head -5
  # No comment about git clone requirement
  ```

### [Docs] Examples.md lists 17 suites but summary table shows 18
- **Severity**: UX-polish
- **What happens**: The opening sentence says "17 server assertion suites" but the summary table lists 18 rows (including the trajectory suite which is not a server suite). Minor inconsistency in counting.
- **Expected**: Either say "18 suites (17 server + 1 trajectory)" in the opening, or exclude the trajectory row from the table count.
- **Repro**: 
  ```bash
  grep "17 server" docs/examples.md
  grep "18 rows" docs/examples.md  # manually count the table
  ```

### [Docs] CHANGELOG.md says [Unreleased] but also has 0.1.3 section with same date
- **Severity**: UX-polish
- **What happens**: The CHANGELOG has both an [Unreleased] section and a [0.1.3] section dated 2026-04-23 with nearly identical content. It's unclear if 0.1.3 is released or not.
- **Expected**: If 0.1.3 is released, move all [Unreleased] content to the 0.1.3 section and leave [Unreleased] empty. If not released, remove the 0.1.3 section.
- **Repro**: 
  ```bash
  grep -A2 "## \[Unreleased\]" CHANGELOG.md
  grep -A2 "## \[0.1.3\]" CHANGELOG.md
  ```

---

## Positive Findings

- **Counts are accurate**: 174 assertions, 18 suites, 218 unit tests all match across README, FEATURES, docs.
- **go.mod is sensible**: Module name is correct, Go version is current.
- **LICENSE is present**: MIT license is in place.
- **Decision tree works well**: The getting-started.md decision tree at the top is clear and actionable.
- **First-round fixes all landed**: All four critical gaps from round 1 are resolved.
- **Documentation is comprehensive**: The docs site covers all major areas with good depth.
- **Package structure is clear**: Runner package has excellent package-level docs explaining the file organization.

---

## Recommendations

**Priority 1 (Critical):**
1. Fix all hardcoded paths in github-mcp YAMLs (use a portable command or document setup)
2. Fix all hardcoded paths in mcp-go YAMLs (use a portable command or document setup)

**Priority 2 (Improvement):**
3. Add --fix to main.go usage output
4. Add audit docs to mkdocs nav
5. Create CONTRIBUTING.md
6. Add package docs for assertion and report packages

**Priority 3 (Polish):**
7. Clarify CHANGELOG release status
8. Fix suite count in examples.md
9. Add setup comments or README to fastmcp-testing-demo directory
