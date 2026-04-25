# "Works with mcp-assert" Badge

Display a badge in your MCP server's README to signal that your tools are tested for protocol correctness.

## Quick Start

Add this to your README:

```markdown
[![Works with mcp-assert](https://img.shields.io/badge/works%20with-mcp--assert-green)](https://github.com/blackwell-systems/mcp-assert)
```

It renders as:

[![Works with mcp-assert](https://img.shields.io/badge/works%20with-mcp--assert-green)](https://github.com/blackwell-systems/mcp-assert)

## Badge with pass rate

If you run mcp-assert in CI and publish a `badge.json` endpoint, you can show live pass rate instead of a static label:

```bash
# In your CI workflow
mcp-assert ci --suite evals/ --badge badge.json
```

Then host `badge.json` via GitHub Pages, a CDN, or any static URL:

```markdown
![mcp-assert](https://img.shields.io/endpoint?url=https://your-site.com/badge.json)
```

The endpoint format is shields.io-compatible JSON:

```json
{
  "schemaVersion": 1,
  "label": "mcp-assert",
  "message": "21/21 passed",
  "color": "green"
}
```

## Full CI setup

The recommended setup uses the GitHub Action and publishes the badge via GitHub Pages:

```yaml
# .github/workflows/mcp-assert.yml
name: MCP Assert

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: blackwell-systems/mcp-assert-action@v1
        with:
          suite: evals/
          badge: badge.json

      - name: Publish badge
        if: github.ref == 'refs/heads/main'
        uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: .
          keep_files: true
          destination_dir: badges
```

Then reference the live badge:

```markdown
![mcp-assert](https://img.shields.io/endpoint?url=https://your-org.github.io/your-repo/badges/badge.json)
```

## Badge variants

| Badge | Markdown |
|-------|----------|
| Static (green) | `[![Works with mcp-assert](https://img.shields.io/badge/works%20with-mcp--assert-green)](https://github.com/blackwell-systems/mcp-assert)` |
| Static (with logo) | `[![Works with mcp-assert](https://img.shields.io/badge/works%20with-mcp--assert-green?logo=checkmarx&logoColor=white)](https://github.com/blackwell-systems/mcp-assert)` |
| Dynamic (live pass rate) | `![mcp-assert](https://img.shields.io/endpoint?url=<your-badge-url>)` |

## What the badge means

The badge says: "This MCP server has a test suite written with mcp-assert, and those tests pass." It is self-attested: the server author writes the assertions and runs them. The value comes from the tests themselves, not from a central authority.

For CI-verified badges (using the dynamic endpoint), the badge reflects the actual pass rate on the latest commit. Readers can click through to the CI logs for verification.
