#!/usr/bin/env bash
# Queries PyPI, npm, GitHub releases, and Homebrew for download stats.
# Generates assets/download-stats.svg with a styled card.
set -euo pipefail

REPO="blackwell-systems/mcp-assert"
PYPI_PKG="mcp-assert"
PYTEST_PKG="pytest-mcp-assert"
NPM_PKG="@blackwell-systems/mcp-assert"
OUT="${1:-assets/download-stats.svg}"

# ── Fetch stats ──────────────────────────────────────────────────────
pypi_week=$(curl -sf "https://pypistats.org/api/packages/${PYPI_PKG}/recent" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['last_week'])" 2>/dev/null || echo "?")

pytest_week=$(curl -sf "https://pypistats.org/api/packages/${PYTEST_PKG}/recent" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['last_week'])" 2>/dev/null || echo "?")

npm_week=$(curl -sf "https://api.npmjs.org/downloads/point/last-week/${NPM_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "?")

gh_total=$(gh api "repos/${REPO}/releases" --jq '[.[].assets[].download_count] | add // 0' 2>/dev/null || echo "?")

brew_30d=$(curl -sf "https://formulae.brew.sh/api/formula/${PYPI_PKG}.json" \
  | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['analytics']['install']['30d'].get('${PYPI_PKG}',0))" 2>/dev/null || echo "--")

# Calculate total weekly
total="?"
if [[ "$pypi_week" != "?" && "$npm_week" != "?" ]]; then
  total=$((pypi_week + npm_week))
  if [[ "$pytest_week" != "?" && "$pytest_week" != "0" ]]; then
    total=$((total + pytest_week))
  fi
  if [[ "$brew_30d" != "--" && "$brew_30d" != "?" && "$brew_30d" != "0" ]]; then
    brew_week=$((brew_30d / 4))
    total=$((total + brew_week))
  fi
fi

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
  <text x="296" y="74" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${pypi_week}/wk</text>

  <text x="24" y="96" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">pip (pytest plugin)</text>
  <text x="296" y="96" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${pytest_week}/wk</text>

  <text x="24" y="118" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">npm</text>
  <text x="296" y="118" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${npm_week}/wk</text>

  <text x="24" y="140" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">brew</text>
  <text x="296" y="140" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${brew_30d}/30d</text>

  <text x="24" y="162" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">github releases</text>
  <text x="296" y="162" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${gh_total} total</text>

  <!-- Divider -->
  <line x1="24" y1="178" x2="296" y2="178" stroke="#334155" stroke-width="1"/>

  <!-- Total -->
  <text x="24" y="206" fill="url(#accent)" font-family="system-ui,-apple-system,sans-serif" font-size="14" font-weight="700">~${total}/week</text>
  <text x="296" y="206" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">across all channels</text>
</svg>
SVGEOF

echo "Generated ${OUT}"
echo "  pip: ${pypi_week}/wk  pytest: ${pytest_week}/wk  npm: ${npm_week}/wk  brew: ${brew_30d}/30d  gh: ${gh_total} total  ~${total}/wk"
