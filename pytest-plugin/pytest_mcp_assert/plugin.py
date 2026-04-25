"""pytest plugin that runs mcp-assert YAML assertions as test items.

Each .yaml file in the configured suite directory becomes a pytest test item.
The plugin calls the mcp-assert binary with --json output and maps the result
to pytest pass/fail/skip.

Usage:
    pytest --mcp-suite evals/
    pytest --mcp-suite evals/ --mcp-fixture /path/to/fixtures
    pytest --mcp-suite evals/ --mcp-server "agent-lsp go:gopls"
    pytest --mcp-suite evals/ --mcp-timeout 60s

Configuration via pyproject.toml:
    [tool.pytest.ini_options]
    mcp_suite = "evals/"
    mcp_fixture = "fixtures/"
    mcp_timeout = "30s"
"""

import json
import os
import shutil
import subprocess

import pytest


def pytest_addoption(parser):
    """Register mcp-assert CLI options."""
    group = parser.getgroup("mcp-assert", "MCP server assertion testing")
    group.addoption(
        "--mcp-suite",
        action="store",
        default=None,
        help="Directory containing mcp-assert YAML assertion files",
    )
    group.addoption(
        "--mcp-fixture",
        action="store",
        default=None,
        help="Fixture directory (substituted for {{fixture}} in assertions)",
    )
    group.addoption(
        "--mcp-server",
        action="store",
        default=None,
        help="Override server command for all assertions",
    )
    group.addoption(
        "--mcp-timeout",
        action="store",
        default="30s",
        help="Per-assertion timeout (default: 30s)",
    )
    group.addoption(
        "--mcp-binary",
        action="store",
        default=None,
        help="Path to mcp-assert binary (default: auto-detect from PATH or pip package)",
    )
    parser.addini(
        "mcp_suite", "Directory containing mcp-assert YAML assertion files"
    )
    parser.addini("mcp_fixture", "Fixture directory for assertions")
    parser.addini("mcp_timeout", "Per-assertion timeout")


def _find_binary(config):
    """Find the mcp-assert binary."""
    # 1. Explicit --mcp-binary flag
    explicit = config.getoption("--mcp-binary")
    if explicit:
        return explicit

    # 2. PATH lookup
    found = shutil.which("mcp-assert")
    if found:
        return found

    # 3. PyPI package (mcp_assert/bin/mcp-assert)
    try:
        import mcp_assert  # noqa: F811

        pkg_dir = os.path.dirname(os.path.abspath(mcp_assert.__file__))
        bin_path = os.path.join(pkg_dir, "bin", "mcp-assert")
        if os.path.isfile(bin_path):
            return bin_path
    except ImportError:
        pass

    return None


def pytest_configure(config):
    """Add the mcp-suite directory to pytest's collection paths."""
    suite_dir = _get_suite_dir(config)
    if suite_dir and os.path.isdir(suite_dir):
        # Only add if no explicit test paths were given (avoid overriding user intent)
        if not config.args or config.args == ["."]:
            config.args = [suite_dir]
        elif suite_dir not in config.args:
            config.args.append(suite_dir)


def pytest_collect_file(parent, file_path):
    """Collect .yaml files from the mcp-suite directory as test items."""
    suite_dir = _get_suite_dir(parent.config)
    if suite_dir is None:
        return None

    suite_path = os.path.abspath(suite_dir)
    file_abs = str(file_path)

    if file_abs.startswith(suite_path) and file_path.suffix == ".yaml":
        return McpAssertFile.from_parent(parent, path=file_path)

    return None


def _get_suite_dir(config):
    """Get suite directory from CLI option or ini."""
    suite = config.getoption("--mcp-suite")
    if suite:
        return suite
    ini = config.getini("mcp_suite")
    if ini:
        return ini
    return None


class McpAssertFile(pytest.File):
    """Collector for a single mcp-assert YAML file."""

    def collect(self):
        """Yield a single test item for this YAML assertion file."""
        # Extract assertion name from YAML (first "name:" line)
        name = self.path.stem
        try:
            with open(self.path) as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("name:"):
                        name = line[5:].strip().strip('"').strip("'")
                        break
        except OSError:
            pass

        yield McpAssertItem.from_parent(self, name=name, yaml_path=self.path)


class McpAssertItem(pytest.Item):
    """A single mcp-assert YAML assertion as a pytest test item."""

    def __init__(self, name, parent, yaml_path):
        super().__init__(name, parent)
        self.yaml_path = yaml_path
        self._result = None

    def runtest(self):
        """Run the mcp-assert binary on this YAML file."""
        config = self.config
        binary = _find_binary(config)
        if binary is None:
            pytest.skip(
                "mcp-assert binary not found. Install via: "
                "brew install blackwell-systems/tap/mcp-assert, "
                "pip install mcp-assert, or "
                "npm install -g @blackwell-systems/mcp-assert"
            )

        cmd = [
            binary,
            "run",
            "--suite",
            str(self.yaml_path),
            "--json",
            "--timeout",
            config.getoption("--mcp-timeout"),
        ]

        fixture = config.getoption("--mcp-fixture") or config.getini("mcp_fixture")
        if fixture:
            cmd.extend(["--fixture", fixture])

        server = config.getoption("--mcp-server")
        if server:
            cmd.extend(["--server", server])

        result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)

        # Parse JSON output
        try:
            results = json.loads(result.stdout)
        except json.JSONDecodeError:
            if result.returncode != 0:
                raise McpAssertFailure(
                    f"mcp-assert failed (exit {result.returncode}): {result.stderr.strip()}"
                )
            raise McpAssertFailure(f"Could not parse mcp-assert output: {result.stdout[:500]}")

        if not results:
            raise McpAssertFailure("mcp-assert returned no results")

        # Each YAML file produces one result
        r = results[0] if isinstance(results, list) else results

        self._result = r
        status = r.get("status", "").lower()

        if status == "skip":
            pytest.skip(r.get("detail", "skipped"))
        elif status == "fail":
            raise McpAssertFailure(r.get("detail", "assertion failed"))

    def repr_failure(self, excinfo):
        """Format assertion failure for pytest output."""
        if isinstance(excinfo.value, McpAssertFailure):
            return str(excinfo.value)
        return super().repr_failure(excinfo)

    def reportinfo(self):
        return self.path, None, f"mcp-assert: {self.name}"


class McpAssertFailure(Exception):
    """Raised when an mcp-assert assertion fails."""
