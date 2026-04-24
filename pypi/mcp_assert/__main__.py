"""Entry point for `python -m mcp_assert`."""

import os
import sys
import subprocess


def _find_binary():
    """Locate the mcp-assert binary bundled in this package."""
    pkg_dir = os.path.dirname(os.path.abspath(__file__))
    # Binary is placed in mcp_assert/bin/ by the wheel build
    names = ["mcp-assert.exe", "mcp-assert"] if sys.platform == "win32" else ["mcp-assert"]
    for name in names:
        path = os.path.join(pkg_dir, "bin", name)
        if os.path.isfile(path):
            return path
    return None


def main():
    binary = _find_binary()
    if binary is None:
        print(
            "mcp-assert: binary not found. This platform may not be supported.\n"
            "Install from https://github.com/blackwell-systems/mcp-assert/releases",
            file=sys.stderr,
        )
        sys.exit(1)

    result = subprocess.run([binary] + sys.argv[1:])
    sys.exit(result.returncode)


if __name__ == "__main__":
    main()
