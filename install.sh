#!/bin/sh
# mcp-assert installer
# Usage: curl -fsSL https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/install.sh | sh
set -e

REPO="blackwell-systems/mcp-assert"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="mcp-assert"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)      echo "error: unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)              echo "error: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Fetch latest release metadata from GitHub
echo "Fetching latest release..."
RELEASE_JSON=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest")
TAG=$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [ -z "$TAG" ]; then
  echo "error: could not determine latest release tag" >&2
  exit 1
fi

# Find the matching asset URL from the release
ASSET_URL=$(echo "$RELEASE_JSON" | grep '"browser_download_url"' | grep "${OS}_${ARCH}.tar.gz" | head -1 | cut -d'"' -f4)

if [ -z "$ASSET_URL" ]; then
  echo "error: no release asset found for ${OS}/${ARCH}" >&2
  echo "Available releases: https://github.com/${REPO}/releases/tag/${TAG}" >&2
  exit 1
fi

echo "Installing mcp-assert ${TAG} for ${OS}/${ARCH}..."

# Download and extract
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "$ASSET_URL" -o "${TMP_DIR}/mcp-assert.tar.gz"
tar -xzf "${TMP_DIR}/mcp-assert.tar.gz" -C "$TMP_DIR"

# Locate binary (GoReleaser may nest it in a subdirectory)
BINARY_PATH=$(find "$TMP_DIR" -name "$BINARY" -type f | head -1)
if [ -z "$BINARY_PATH" ]; then
  echo "error: could not find ${BINARY} in downloaded archive" >&2
  exit 1
fi

# Install binary
if [ -w "$INSTALL_DIR" ]; then
  mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
  sudo chmod +x "${INSTALL_DIR}/${BINARY}"
fi

# Verify
VERSION=$("${INSTALL_DIR}/${BINARY}" version 2>/dev/null || echo "unknown")
echo ""
echo "Installed mcp-assert ${VERSION} to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Quick start:"
echo "  mcp-assert init evals --server \"your-mcp-server\""
echo "  mcp-assert run --suite evals/"
