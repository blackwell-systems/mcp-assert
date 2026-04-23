# CI Integration

## GitHub Action

```yaml
- name: Assert MCP server correctness
  run: |
    go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest
    mcp-assert ci --suite evals/ --threshold 95 --junit results.xml

- name: Upload test results
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: mcp-assert-results
    path: results.xml
```

A dedicated GitHub Action is also available: [`blackwell-systems/mcp-assert-action@v1`](https://github.com/blackwell-systems/mcp-assert-action): one line in any workflow. Downloads binary, runs assertions, uploads JUnit XML + badge.

## CI commands

```bash
# Fail the build if any assertion regresses (requires --baseline)
mcp-assert ci --suite evals/ --baseline baseline.json --fail-on-regression

# Set a minimum pass threshold
mcp-assert ci --suite evals/ --threshold 95

# Override server from CLI
mcp-assert ci --suite evals/ --server "my-mcp-server" --threshold 100
```

## JUnit XML

Standard JUnit format for CI test result tabs (GitHub Actions, Jenkins, CircleCI). Includes pass@k/pass^k properties when `--trials > 1`.

```bash
mcp-assert run --suite evals/ --junit results.xml
```

## Markdown Summary

GitHub Step Summary table. Auto-detects `$GITHUB_STEP_SUMMARY` in ci mode. Includes reliability section when `--trials > 1`.

```bash
mcp-assert ci --suite evals/ --markdown summary.md
```

## Badge

shields.io endpoint JSON for README badges:

```bash
mcp-assert run --suite evals/ --badge badge.json
# Then use: ![mcp-assert](https://img.shields.io/endpoint?url=<badge-url>)
```

## Baseline and Regression Detection

Save a baseline, then detect regressions on future runs:

```bash
# Save current results as baseline
mcp-assert run --suite evals/ --save-baseline baseline.json

# Later: compare against baseline
mcp-assert ci --suite evals/ --baseline baseline.json --fail-on-regression
```

Only flags transitions from PASS to FAIL. Previously-failing tests that still fail are not regressions. New tests that fail are not regressions.
