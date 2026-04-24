#!/bin/bash
# Build and publish platform-specific wheels to PyPI.
#
# Usage: ./scripts/pypi-publish.sh [TAG]
# TAG defaults to the latest git tag (e.g. v0.2.0)
#
# Requires: TWINE_USERNAME and TWINE_PASSWORD env vars (or __token__ + API token)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYPI_DIR="${SCRIPT_DIR}/../pypi"
DIST_DIR="${PYPI_DIR}/dist"
TAG="${1:-$(git describe --tags --abbrev=0)}"

# Build wheels
"${SCRIPT_DIR}/pypi-build-wheels.sh" "$TAG"

# Publish
echo "Publishing to PyPI..."
python3 -m twine upload "${DIST_DIR}"/*.whl

echo ""
echo "Done. Install with: pip install mcp-assert"
