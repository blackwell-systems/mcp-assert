#!/bin/bash
# Publish @blackwell-systems/mcp-assert packages to npm.
# Downloads binaries from the GitHub release for TAG, injects them into
# platform packages, and publishes all packages.
#
# Usage: ./scripts/npm-publish.sh [TAG]
# TAG defaults to the latest git tag (e.g. v0.2.0)
#
# Requires: npm (authenticated), curl, tar, unzip, node

set -euo pipefail

TAG="${1:-$(git describe --tags --abbrev=0)}"
VERSION="${TAG#v}"
REPO="blackwell-systems/mcp-assert"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NPM_DIR="${SCRIPT_DIR}/../npm"

echo "Publishing @blackwell-systems/mcp-assert@${VERSION} (tag: ${TAG})"

# Mapping: goreleaser archive suffix -> "npm_suffix:binary_name:archive_ext"
# GoReleaser names: mcp-assert_VERSION_OS_ARCH.tar.gz
declare -A PLATFORMS=(
  ["darwin_arm64"]="darwin-arm64:mcp-assert:tar.gz"
  ["darwin_amd64"]="darwin-x64:mcp-assert:tar.gz"
  ["linux_arm64"]="linux-arm64:mcp-assert:tar.gz"
  ["linux_amd64"]="linux-x64:mcp-assert:tar.gz"
  ["windows_amd64"]="win32-x64:mcp-assert.exe:zip"
  ["windows_arm64"]="win32-arm64:mcp-assert.exe:zip"
)

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

for GOKEY in "${!PLATFORMS[@]}"; do
  IFS=: read -r NPM_SUFFIX BINARY_NAME ARCHIVE_EXT <<< "${PLATFORMS[$GOKEY]}"

  ARCHIVE="mcp-assert_${VERSION}_${GOKEY}.${ARCHIVE_EXT}"
  URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"
  PKG_DIR="${NPM_DIR}/mcp-assert-${NPM_SUFFIX}"
  BIN_DIR="${PKG_DIR}/bin"

  echo "  [${NPM_SUFFIX}] Downloading ${ARCHIVE}..."
  curl -fsSL "$URL" -o "${TMP_DIR}/${ARCHIVE}"

  mkdir -p "$BIN_DIR"

  if [ "$ARCHIVE_EXT" = "tar.gz" ]; then
    tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR" "$BINARY_NAME"
  else
    unzip -o "${TMP_DIR}/${ARCHIVE}" "$BINARY_NAME" -d "$TMP_DIR"
  fi

  cp "${TMP_DIR}/${BINARY_NAME}" "${BIN_DIR}/${BINARY_NAME}"
  chmod +x "${BIN_DIR}/${BINARY_NAME}"
  rm -f "${TMP_DIR}/${BINARY_NAME}"
  # npm pack honors nested .gitignore files even when "files" is set in
  # package.json. A bin/.gitignore would exclude the binary from the tarball.
  rm -f "${BIN_DIR}/.gitignore"

  # Update version
  node -e "
    const fs = require('fs');
    const p = '${PKG_DIR}/package.json';
    const pkg = JSON.parse(fs.readFileSync(p));
    pkg.version = '${VERSION}';
    fs.writeFileSync(p, JSON.stringify(pkg, null, 2) + '\n');
  "

  # Skip if this exact version is already on the registry (idempotent reruns).
  PKG_NAME="@blackwell-systems/mcp-assert-${NPM_SUFFIX}"
  if npm view "${PKG_NAME}@${VERSION}" version &>/dev/null 2>&1; then
    echo "  [${NPM_SUFFIX}] Already published at ${VERSION}, skipping."
  else
    echo "  [${NPM_SUFFIX}] Publishing ${PKG_NAME}@${VERSION}..."
    npm publish "${PKG_DIR}" --access public
  fi
done

# Sync the root (meta) package version and pin every optionalDependency to the
# same release. The root package has no binary itself; it lists per-platform
# packages as optionalDependencies so npm installs only the matching one.
ROOT_PKG="${NPM_DIR}/mcp-assert/package.json"
node -e "
  const fs = require('fs');
  const pkg = JSON.parse(fs.readFileSync('${ROOT_PKG}'));
  pkg.version = '${VERSION}';
  for (const dep of Object.keys(pkg.optionalDependencies)) {
    pkg.optionalDependencies[dep] = '${VERSION}';
  }
  fs.writeFileSync('${ROOT_PKG}', JSON.stringify(pkg, null, 2) + '\n');
"

ROOT_PKG_NAME="@blackwell-systems/mcp-assert"
if npm view "${ROOT_PKG_NAME}@${VERSION}" version &>/dev/null 2>&1; then
  echo "  [root] Already published at ${VERSION}, skipping."
else
  echo "  [root] Publishing ${ROOT_PKG_NAME}@${VERSION}..."
  npm publish "${NPM_DIR}/mcp-assert" --access public
fi

echo ""
echo "Done. Install with: npx @blackwell-systems/mcp-assert"
