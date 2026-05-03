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

packagist_total=$(curl -sf --max-time 10 "https://packagist.org/packages/blackwell-systems/phpunit-mcp-assert.json" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['package']['downloads']['total'])" 2>/dev/null || echo "--")

# Snap doesn't expose public download counts via API.
# Stats are only visible on the Snapcraft dashboard.

# ── High-water mark: never regress displayed totals ────────────────
# Cache stores the last known good value per channel. Downloads are
# monotonic, so if the API returns a lower number or fails, we keep
# the previous value. This prevents badge flicker from API outages
# or rolling-window quirks in PyPI stats.
read_cache() {
  local key="$1"
  if [[ -f "$CACHE" ]]; then
    grep "^${key}=" "$CACHE" 2>/dev/null | cut -d= -f2
  fi
}

# Returns max(fresh_value, cached_value), falling back to cache on API failure.
# Guarantees displayed totals never decrease between runs.
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
packagist_total=$(use_or_cache packagist "$packagist_total")
docker_total=$(use_or_cache docker "$docker_total")

# Write cache with current best values.
cat > "$CACHE" << CACHEEOF
pip=${pypi_total}
pytest=${pytest_total}
npm=${npm_total}
vitest=${vitest_total}
jest=${jest_total}
gh=${gh_total}
packagist=${packagist_total}
docker=${docker_total}
CACHEEOF

# ── Calculate cumulative total ──────────────────────────────────────
# Sum all channels that returned a numeric value; skip unknowns so they
# don't pollute the total. If every channel failed, display "?" instead of 0.
cumulative=0
for v in "$pypi_total" "$pytest_total" "$npm_total" "$vitest_total" "$jest_total" "$gh_total" "$packagist_total" "$docker_total"; do
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
packagist_fmt="$packagist_total"
if [[ "$packagist_total" != "--" && "$packagist_total" != "?" ]]; then
  packagist_fmt=$(fmt "$packagist_total" 2>/dev/null || echo "$packagist_total")
fi
docker_fmt="$docker_total"
if [[ "$docker_total" != "--" && "$docker_total" != "?" ]]; then
  docker_fmt=$(fmt "$docker_total" 2>/dev/null || echo "$docker_total")
fi
cumulative_fmt=$(fmt "$cumulative" 2>/dev/null || echo "$cumulative")

date_str=$(date +"%Y-%m-%d")

# ── Build rows dynamically (hide channels with zero downloads) ──────
# Each visible row is 22px tall, starting at y=74.
has_downloads() {
  local val="$1"
  # "--" = package not yet published (known zero); "0" = confirmed zero.
  # "?" = API failed but package exists, so we still render the row.
  [[ "$val" != "--" && "$val" != "0" ]]
}

rows=""
row_count=0
add_row() {
  local label="$1" value="$2"
  local y=$(( 74 + row_count * 22 ))
  rows+="
  <text x=\"24\" y=\"${y}\" fill=\"#94a3b8\" font-family=\"ui-monospace,monospace\" font-size=\"12\">${label}</text>
  <text x=\"296\" y=\"${y}\" fill=\"#e2e8f0\" font-family=\"ui-monospace,monospace\" font-size=\"13\" font-weight=\"600\" text-anchor=\"end\">${value}</text>"
  row_count=$((row_count + 1))
}

has_downloads "$pypi_total"     && add_row "pip (mcp-assert)"   "$pypi_fmt"
has_downloads "$pytest_total"   && add_row "pip (pytest plugin)" "$pytest_fmt"
has_downloads "$npm_total"      && add_row "npm (cli)"          "$npm_fmt"
has_downloads "$vitest_total"   && add_row "npm (vitest plugin)" "$vitest_fmt"
has_downloads "$jest_total"     && add_row "npm (jest plugin)"  "$jest_fmt"
has_downloads "$gh_total"       && add_row "github releases"    "$gh_fmt"
has_downloads "$packagist_total" && add_row "packagist (phpunit)" "$packagist_fmt"
has_downloads "$docker_total"   && add_row "docker pulls"       "$docker_fmt"

# Calculate SVG height: header(48) + rows(22 each) + divider gap(16) + total row(28) + bottom padding(20)
svg_height=$(( 48 + row_count * 22 + 16 + 28 + 20 ))
divider_y=$(( 48 + row_count * 22 + 8 ))
total_y=$(( divider_y + 28 ))
stroke_height=$(( svg_height - 2 ))

# ── Generate SVG ─────────────────────────────────────────────────────
cat > "$OUT" << SVGEOF
<svg xmlns="http://www.w3.org/2000/svg" width="320" height="${svg_height}" viewBox="0 0 320 ${svg_height}">
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
  <rect width="320" height="${svg_height}" rx="12" fill="url(#bg)"/>
  <rect x="1" y="1" width="318" height="${stroke_height}" rx="11" fill="none" stroke="#334155" stroke-width="1"/>

  <!-- Title -->
  <text x="24" y="36" fill="#e2e8f0" font-family="system-ui,-apple-system,sans-serif" font-size="14" font-weight="600">mcp-assert downloads</text>
  <text x="296" y="36" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">${date_str}</text>

  <!-- Divider -->
  <line x1="24" y1="48" x2="296" y2="48" stroke="#334155" stroke-width="1"/>

  <!-- Stats rows -->${rows}

  <!-- Divider -->
  <line x1="24" y1="${divider_y}" x2="296" y2="${divider_y}" stroke="#334155" stroke-width="1"/>

  <!-- Total -->
  <text x="24" y="${total_y}" fill="url(#accent)" font-family="system-ui,-apple-system,sans-serif" font-size="16" font-weight="700">${cumulative_fmt} total</text>
  <text x="296" y="${total_y}" fill="#64748b" font-family="system-ui,-apple-system,sans-serif" font-size="10" text-anchor="end">cumulative downloads</text>
</svg>
SVGEOF

# Generate downloads badge JSON for shields.io endpoint badge.
BADGE_OUT="$(dirname "$OUT")/downloads-badge.json"
if [[ "$cumulative" != "?" ]]; then
  badge_label="downloads"
  badge_msg="${cumulative_fmt}"
  badge_color="ff69b4"
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
echo "  pip: ${pypi_total}  pytest: ${pytest_total}  npm: ${npm_total}  vitest: ${vitest_total}  jest: ${jest_total}  gh: ${gh_total}  packagist: ${packagist_total}  docker: ${docker_total}  total: ${cumulative}"
