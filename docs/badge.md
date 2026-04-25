# "Works with mcp-assert" Badge

Display a badge in your MCP server's README to signal that your tools are tested for protocol correctness.

## Badge variants

Three custom badges are available, hosted from the mcp-assert repository:

| Preview | Variant | Use when |
|---------|---------|----------|
| ![passing](https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/badge-passing.svg) | Passing | All assertions pass |
| ![score](https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/badge-score.svg) | Score | Show pass count (customize the SVG with your numbers) |
| ![failing](https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/badge-failing.svg) | Failing | Some assertions fail (for honesty in dashboards) |

## Quick Start

Add the passing badge to your README:

```markdown
[![mcp-assert: passing](https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/badge-passing.svg)](https://github.com/blackwell-systems/mcp-assert)
```

Or use the score badge (download, edit the `20/20` text in the SVG to match your count, commit to your repo):

```markdown
[![mcp-assert: 20/20](https://raw.githubusercontent.com/your-org/your-repo/main/assets/mcp-assert-badge.svg)](https://github.com/blackwell-systems/mcp-assert)
```

## Shields.io alternative

If you prefer shields.io's CDN and caching, a generic badge is also available:

```markdown
[![Works with mcp-assert](https://img.shields.io/badge/works%20with-mcp--assert-green)](https://github.com/blackwell-systems/mcp-assert)
```

## Dynamic badge with CI

If you run mcp-assert in CI, you can publish a live pass rate badge:

```bash
# In your CI workflow
mcp-assert ci --suite evals/ --badge badge.json
```

The `--badge` flag generates shields.io-compatible endpoint JSON:

```json
{
  "schemaVersion": 1,
  "label": "mcp-assert",
  "message": "21/21 passed",
  "color": "green"
}
```

Host `badge.json` via GitHub Pages, a CDN, or any static URL, then reference it:

```markdown
![mcp-assert](https://img.shields.io/endpoint?url=https://your-site.com/badge.json)
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

## What the badge means

The badge says: "This MCP server has a test suite written with mcp-assert, and those tests pass." It is self-attested: the server author writes the assertions and runs them. The value comes from the tests themselves, not from a central authority.

For CI-verified badges (using the dynamic endpoint), the badge reflects the actual pass rate on the latest commit. Readers can click through to the CI logs for verification.
