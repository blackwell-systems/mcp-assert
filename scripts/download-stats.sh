#!/usr/bin/env bash
# Queries PyPI, npm, GitHub releases, and Homebrew for download stats.
# Generates assets/download-stats.svg with a styled card.
set -euo pipefail

REPO="blackwell-systems/mcp-assert"
PYPI_PKG="mcp-assert"
PYTEST_PKG="pytest-mcp-assert"
NPM_PKG="@blackwell-systems/mcp-assert"
VITEST_PKG="vitest-mcp-assert"
OUT="${1:-assets/download-stats.svg}"

# ── Extract last known values from existing SVG (fallback for blocked APIs) ──
extract_last() {
  local label="$1"
  grep "$label" "$OUT" 2>/dev/null | grep -o '[0-9][0-9,]*' | tail -1 || echo "?"
}
last_pypi=$(extract_last "pip (mcp-assert)")
last_pytest=$(extract_last "pip (pytest plugin)")

# ── Fetch stats ──────────────────────────────────────────────────────
UA="mcp-assert-stats/1.0 (https://github.com/blackwell-systems/mcp-assert)"

pypi_week=$(curl -sf -A "$UA" --max-time 10 "https://pypistats.org/api/packages/${PYPI_PKG}/recent" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['last_week'])" 2>/dev/null || echo "$last_pypi")

pytest_week=$(curl -sf -A "$UA" --max-time 10 "https://pypistats.org/api/packages/${PYTEST_PKG}/recent" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['last_week'])" 2>/dev/null || echo "$last_pytest")

npm_week=$(curl -sf --max-time 10 "https://api.npmjs.org/downloads/point/last-week/${NPM_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "?")

vitest_week=$(curl -sf --max-time 10 "https://api.npmjs.org/downloads/point/last-week/${VITEST_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "--")

gh_total=$(gh api "repos/${REPO}/releases" --jq '[.[].assets[].download_count] | add // 0' 2>/dev/null || echo "?")

brew_30d=$(curl -sf "https://formulae.brew.sh/api/formula/${PYPI_PKG}.json" \
  | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['analytics']['install']['30d'].get('${PYPI_PKG}',0))" 2>/dev/null || echo "--")

# ── Fetch all-time totals ───────────────────────────────────────────
pypi_alltime=$(curl -sf -A "$UA" --max-time 10 "https://pypistats.org/api/packages/${PYPI_PKG}/overall" \
  | python3 -c "import json,sys; print(sum(r['downloads'] for r in json.load(sys.stdin).get('data',[])))" 2>/dev/null || echo "0")

pytest_alltime=$(curl -sf -A "$UA" --max-time 10 "https://pypistats.org/api/packages/${PYTEST_PKG}/overall" \
  | python3 -c "import json,sys; print(sum(r['downloads'] for r in json.load(sys.stdin).get('data',[])))" 2>/dev/null || echo "0")

npm_alltime=$(curl -sf --max-time 10 "https://api.npmjs.org/downloads/point/2000-01-01:2030-01-01/${NPM_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "0")

vitest_alltime=$(curl -sf --max-time 10 "https://api.npmjs.org/downloads/point/2000-01-01:2030-01-01/${VITEST_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "0")

# Calculate cumulative total
cumulative="?"
if [[ "$pypi_alltime" != "0" || "$npm_alltime" != "0" ]]; then
  cumulative=$((pypi_alltime + pytest_alltime + npm_alltime + vitest_alltime))
  if [[ "$gh_total" != "?" ]]; then
    cumulative=$((cumulative + gh_total))
  fi
fi

# Calculate total weekly
total="?"
if [[ "$pypi_week" != "?" && "$npm_week" != "?" ]]; then
  total=$((pypi_week + npm_week))
  if [[ "$pytest_week" != "?" && "$pytest_week" != "0" ]]; then
    total=$((total + pytest_week))
  fi
  if [[ "$vitest_week" != "--" && "$vitest_week" != "?" && "$vitest_week" != "0" ]]; then
    total=$((total + vitest_week))
  fi
  if [[ "$brew_30d" != "--" && "$brew_30d" != "?" && "$brew_30d" != "0" ]]; then
    brew_week=$((brew_30d / 4))
    total=$((total + brew_week))
  fi
fi

date_str=$(date +"%Y-%m-%d")

# ── Generate SVG ─────────────────────────────────────────────────────
cat > "$OUT" << SVGEOF
<svg xmlns="http://www.w3.org/2000/svg" width="320" height="268" viewBox="0 0 320 268">
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
  <rect width="320" height="268" rx="12" fill="url(#bg)"/>
  <rect x="1" y="1" width="318" height="266" rx="11" fill="none" stroke="#334155" stroke-width="1"/>

  <!-- Title -->
  <text x="24" y="36" fill="#e2e8f0" font-family="system-ui,-apple-system,sans-serif" font-size="14" font-weight="600">mcp-assert downloads</text>
  <text x="296" y="36" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">${date_str}</text>

  <!-- Divider -->
  <line x1="24" y1="48" x2="296" y2="48" stroke="#334155" stroke-width="1"/>

  <!-- Stats rows -->
  <text x="24" y="74" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">pip (mcp-assert)</text>
  <text x="296" y="74" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${pypi_week}/wk</text>

  <text x="24" y="96" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">pip (pytest plugin)</text>
  <text x="296" y="96" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${pytest_week}/wk</text>

  <text x="24" y="118" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">npm (cli)</text>
  <text x="296" y="118" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${npm_week}/wk</text>

  <text x="24" y="140" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">npm (vitest plugin)</text>
  <text x="296" y="140" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${vitest_week}/wk</text>

  <text x="24" y="162" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">brew</text>
  <text x="296" y="162" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${brew_30d}/30d</text>

  <text x="24" y="184" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">github releases</text>
  <text x="296" y="184" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${gh_total} total</text>

  <!-- Divider -->
  <line x1="24" y1="200" x2="296" y2="200" stroke="#334155" stroke-width="1"/>

  <!-- Totals -->
  <text x="24" y="224" fill="url(#accent)" font-family="system-ui,-apple-system,sans-serif" font-size="14" font-weight="700">~${total}/week</text>
  <text x="296" y="224" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">across all channels</text>

  <text x="24" y="250" fill="url(#accent)" font-family="system-ui,-apple-system,sans-serif" font-size="14" font-weight="700">${cumulative} total</text>
  <text x="296" y="250" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">cumulative downloads</text>
</svg>
SVGEOF

echo "Generated ${OUT}"
echo "  pip: ${pypi_week}/wk  pytest: ${pytest_week}/wk  npm: ${npm_week}/wk  vitest: ${vitest_week}/wk  brew: ${brew_30d}/30d  gh: ${gh_total} total  ~${total}/wk  ${cumulative} cumulative"
