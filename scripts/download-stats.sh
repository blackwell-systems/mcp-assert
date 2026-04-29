#!/usr/bin/env bash
# Queries PyPI, npm, GitHub releases for cumulative download stats.
# Generates assets/download-stats.svg with a styled card.
set -euo pipefail

REPO="blackwell-systems/mcp-assert"
PYPI_PKG="mcp-assert"
PYTEST_PKG="pytest-mcp-assert"
NPM_PKG="@blackwell-systems/mcp-assert"
VITEST_PKG="vitest-mcp-assert"
OUT="${1:-assets/download-stats.svg}"

# ── Fetch all-time totals ───────────────────────────────────────────
UA="mcp-assert-stats/1.0 (https://github.com/blackwell-systems/mcp-assert)"

pypi_total=$(curl -sf -A "$UA" --max-time 10 "https://pypistats.org/api/packages/${PYPI_PKG}/overall" \
  | python3 -c "import json,sys; print(sum(r['downloads'] for r in json.load(sys.stdin).get('data',[])))" 2>/dev/null || echo "?")

pytest_total=$(curl -sf -A "$UA" --max-time 10 "https://pypistats.org/api/packages/${PYTEST_PKG}/overall" \
  | python3 -c "import json,sys; print(sum(r['downloads'] for r in json.load(sys.stdin).get('data',[])))" 2>/dev/null || echo "?")

npm_total=$(curl -sf --max-time 10 "https://api.npmjs.org/downloads/point/2000-01-01:2030-01-01/${NPM_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "?")

vitest_total=$(curl -sf --max-time 10 "https://api.npmjs.org/downloads/point/2000-01-01:2030-01-01/${VITEST_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "--")

gh_total=$(gh api "repos/${REPO}/releases" --jq '[.[].assets[].download_count] | add // 0' 2>/dev/null || echo "?")

# ── Calculate cumulative total ──────────────────────────────────────
cumulative="?"
if [[ "$pypi_total" != "?" && "$npm_total" != "?" ]]; then
  cumulative=$((pypi_total))
  if [[ "$pytest_total" != "?" ]]; then
    cumulative=$((cumulative + pytest_total))
  fi
  if [[ "$npm_total" != "?" ]]; then
    cumulative=$((cumulative + npm_total))
  fi
  if [[ "$vitest_total" != "--" && "$vitest_total" != "?" ]]; then
    cumulative=$((cumulative + vitest_total))
  fi
  if [[ "$gh_total" != "?" ]]; then
    cumulative=$((cumulative + gh_total))
  fi
fi

# Format numbers with commas
fmt() {
  printf "%'d" "$1" 2>/dev/null || echo "$1"
}

pypi_fmt=$(fmt "$pypi_total" 2>/dev/null || echo "$pypi_total")
pytest_fmt=$(fmt "$pytest_total" 2>/dev/null || echo "$pytest_total")
npm_fmt=$(fmt "$npm_total" 2>/dev/null || echo "$npm_total")
vitest_fmt="$vitest_total"
if [[ "$vitest_total" != "--" && "$vitest_total" != "?" ]]; then
  vitest_fmt=$(fmt "$vitest_total" 2>/dev/null || echo "$vitest_total")
fi
gh_fmt=$(fmt "$gh_total" 2>/dev/null || echo "$gh_total")
cumulative_fmt=$(fmt "$cumulative" 2>/dev/null || echo "$cumulative")

date_str=$(date +"%Y-%m-%d")

# ── Generate SVG ─────────────────────────────────────────────────────
cat > "$OUT" << SVGEOF
<svg xmlns="http://www.w3.org/2000/svg" width="320" height="224" viewBox="0 0 320 224">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#1a1a2e"/>
      <stop offset="100%" stop-color="#16213e"/>
    </linearGradient>
    <linearGradient id="accent" x1="0" y1="0" x2="1" y2="0">
      <stop offset="0%" stop-color="#4ade80"/>
      <stop offset="100%" stop-color="#22d3ee"/>
    </linearGradient>
  </defs>

  <!-- Card background -->
  <rect width="320" height="224" rx="12" fill="url(#bg)"/>
  <rect x="1" y="1" width="318" height="222" rx="11" fill="none" stroke="#334155" stroke-width="1"/>

  <!-- Title -->
  <text x="24" y="36" fill="#e2e8f0" font-family="system-ui,-apple-system,sans-serif" font-size="14" font-weight="600">mcp-assert downloads</text>
  <text x="296" y="36" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">${date_str}</text>

  <!-- Divider -->
  <line x1="24" y1="48" x2="296" y2="48" stroke="#334155" stroke-width="1"/>

  <!-- Stats rows -->
  <text x="24" y="74" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">pip (mcp-assert)</text>
  <text x="296" y="74" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${pypi_fmt}</text>

  <text x="24" y="96" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">pip (pytest plugin)</text>
  <text x="296" y="96" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${pytest_fmt}</text>

  <text x="24" y="118" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">npm (cli)</text>
  <text x="296" y="118" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${npm_fmt}</text>

  <text x="24" y="140" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">npm (vitest plugin)</text>
  <text x="296" y="140" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${vitest_fmt}</text>

  <text x="24" y="162" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">github releases</text>
  <text x="296" y="162" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${gh_fmt}</text>

  <!-- Divider -->
  <line x1="24" y1="178" x2="296" y2="178" stroke="#334155" stroke-width="1"/>

  <!-- Total -->
  <text x="24" y="206" fill="url(#accent)" font-family="system-ui,-apple-system,sans-serif" font-size="16" font-weight="700">${cumulative_fmt} total</text>
  <text x="296" y="206" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">cumulative downloads</text>
</svg>
SVGEOF

echo "Generated ${OUT}"
echo "  pip: ${pypi_total}  pytest: ${pytest_total}  npm: ${npm_total}  vitest: ${vitest_total}  gh: ${gh_total}  total: ${cumulative}"
