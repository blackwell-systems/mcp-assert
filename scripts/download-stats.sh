#!/usr/bin/env bash
# Queries PyPI, npm, GitHub releases for cumulative download stats.
# Generates assets/download-stats.svg with a styled card.
set -euo pipefail

REPO="blackwell-systems/mcp-assert"
PYPI_PKG="mcp-assert"
PYTEST_PKG="pytest-mcp-assert"
NPM_PKG="@blackwell-systems/mcp-assert"
VITEST_PKG="vitest-mcp-assert"
JEST_PKG="jest-mcp-assert"
OUT="${1:-assets/download-stats.svg}"
CACHE="${OUT%.svg}.cache"

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

jest_total=$(curl -sf --max-time 10 "https://api.npmjs.org/downloads/point/2000-01-01:2030-01-01/${JEST_PKG}" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['downloads'])" 2>/dev/null || echo "--")

gh_total=$(gh api "repos/${REPO}/releases" --jq '[.[].assets[].download_count] | add // 0' 2>/dev/null || echo "?")

docker_total=$(curl -sf --max-time 10 "https://hub.docker.com/v2/repositories/blackwellsystems/mcp-assert/" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['pull_count'])" 2>/dev/null || echo "--")

# Snap doesn't expose public download counts via API.
# Stats are only visible on the Snapcraft dashboard.

# ── High-water mark: never regress displayed totals ────────────────
# Cache stores the last known good value per channel. Downloads are
# monotonic, so if the API returns a lower number or fails, we keep
# the previous value.
read_cache() {
  local key="$1"
  if [[ -f "$CACHE" ]]; then
    grep "^${key}=" "$CACHE" 2>/dev/null | cut -d= -f2
  fi
}

use_or_cache() {
  local key="$1" val="$2"
  local prev
  prev=$(read_cache "$key")
  prev="${prev:-0}"
  # If we got a valid number from the API:
  if [[ "$val" != "?" && "$val" != "--" ]]; then
    # If the cached value is also a valid number, keep the higher one.
    if [[ "$prev" != "?" && "$prev" != "--" && "$prev" != "0" ]]; then
      if (( val >= prev )); then
        echo "$val"
      else
        echo "$prev"
      fi
    else
      # Cache was empty/invalid, use the fresh value.
      echo "$val"
    fi
    return
  fi
  # API failed: fall back to cache if it has a valid value.
  if [[ "$prev" != "0" && "$prev" != "--" && "$prev" != "?" ]]; then
    echo "$prev"
  else
    echo "$val"
  fi
}

pypi_total=$(use_or_cache pip "$pypi_total")
pytest_total=$(use_or_cache pytest "$pytest_total")
npm_total=$(use_or_cache npm "$npm_total")
vitest_total=$(use_or_cache vitest "$vitest_total")
jest_total=$(use_or_cache jest "$jest_total")
gh_total=$(use_or_cache gh "$gh_total")
docker_total=$(use_or_cache docker "$docker_total")

# Write cache with current best values.
cat > "$CACHE" << CACHEEOF
pip=${pypi_total}
pytest=${pytest_total}
npm=${npm_total}
vitest=${vitest_total}
jest=${jest_total}
gh=${gh_total}
docker=${docker_total}
CACHEEOF

# ── Calculate cumulative total ──────────────────────────────────────
cumulative=0
for v in "$pypi_total" "$pytest_total" "$npm_total" "$vitest_total" "$jest_total" "$gh_total" "$docker_total"; do
  if [[ "$v" != "?" && "$v" != "--" ]]; then
    cumulative=$((cumulative + v))
  fi
done
if (( cumulative == 0 )); then cumulative="?"; fi

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
jest_fmt="$jest_total"
if [[ "$jest_total" != "--" && "$jest_total" != "?" ]]; then
  jest_fmt=$(fmt "$jest_total" 2>/dev/null || echo "$jest_total")
fi
docker_fmt="$docker_total"
if [[ "$docker_total" != "--" && "$docker_total" != "?" ]]; then
  docker_fmt=$(fmt "$docker_total" 2>/dev/null || echo "$docker_total")
fi
cumulative_fmt=$(fmt "$cumulative" 2>/dev/null || echo "$cumulative")

date_str=$(date +"%Y-%m-%d")

# ── Generate SVG ─────────────────────────────────────────────────────
cat > "$OUT" << SVGEOF
<svg xmlns="http://www.w3.org/2000/svg" width="320" height="270" viewBox="0 0 320 270">
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
  <rect width="320" height="270" rx="12" fill="url(#bg)"/>
  <rect x="1" y="1" width="318" height="268" rx="11" fill="none" stroke="#334155" stroke-width="1"/>

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

  <text x="24" y="162" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">npm (jest plugin)</text>
  <text x="296" y="162" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${jest_fmt}</text>

  <text x="24" y="184" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">github releases</text>
  <text x="296" y="184" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${gh_fmt}</text>

  <text x="24" y="206" fill="#94a3b8" font-family="ui-monospace,monospace" font-size="12">docker pulls</text>
  <text x="296" y="206" fill="#e2e8f0" font-family="ui-monospace,monospace" font-size="13" font-weight="600" text-anchor="end">${docker_fmt}</text>

  <!-- Divider -->
  <line x1="24" y1="222" x2="296" y2="222" stroke="#334155" stroke-width="1"/>

  <!-- Total -->
  <text x="24" y="250" fill="url(#accent)" font-family="system-ui,-apple-system,sans-serif" font-size="16" font-weight="700">${cumulative_fmt} total</text>
  <text x="296" y="250" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">cumulative downloads</text>
</svg>
SVGEOF

# Generate downloads badge JSON for shields.io endpoint badge.
BADGE_OUT="$(dirname "$OUT")/downloads-badge.json"
if [[ "$cumulative" != "?" ]]; then
  badge_label="downloads"
  badge_msg="${cumulative_fmt}"
  badge_color="brightgreen"
  if (( cumulative < 1000 )); then badge_color="green"; fi
  cat > "$BADGE_OUT" << BADGEEOF
{
  "schemaVersion": 1,
  "label": "${badge_label}",
  "message": "${badge_msg}",
  "color": "${badge_color}"
}
BADGEEOF
fi

echo "Generated ${OUT}"
echo "  pip: ${pypi_total}  pytest: ${pytest_total}  npm: ${npm_total}  vitest: ${vitest_total}  jest: ${jest_total}  gh: ${gh_total}  docker: ${docker_total}  total: ${cumulative}"
