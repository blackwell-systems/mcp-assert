# pytest-mcp-assert

pytest plugin for [mcp-assert](https://github.com/blackwell-systems/mcp-assert). Run MCP server assertions as pytest test items.

## Install

```bash
pip install pytest-mcp-assert
```

You also need the mcp-assert binary. Any of these:

```bash
brew install blackwell-systems/tap/mcp-assert
pip install mcp-assert
npm install -g @blackwell-systems/mcp-assert
```

## Usage

Write assertion YAML files (see [mcp-assert docs](https://blackwell-systems.github.io/mcp-assert)):

```yaml
# evals/echo.yaml
name: echo returns the message
server:
  command: my-mcp-server
assert:
  tool: echo
  args:
    message: hello
  expect:
    not_error: true
    contains: ["hello"]
timeout: 30s
```

Run with pytest:

```bash
pytest --mcp-suite evals/
```

Each YAML file becomes a pytest test item with pass/fail/skip semantics.

## Options

| Option | pyproject.toml | Description |
|--------|---------------|-------------|
| `--mcp-suite DIR` | `mcp_suite` | Directory containing assertion YAML files |
| `--mcp-fixture DIR` | `mcp_fixture` | Fixture directory (substituted for `{{fixture}}`) |
| `--mcp-server CMD` | | Override server command for all assertions |
| `--mcp-timeout DUR` | `mcp_timeout` | Per-assertion timeout (default: 30s) |
| `--mcp-binary PATH` | | Path to mcp-assert binary |

## Configuration

```toml
# pyproject.toml
[tool.pytest.ini_options]
mcp_suite = "evals/"
mcp_fixture = "fixtures/"
mcp_timeout = "30s"
```

Then just run `pytest` with no extra flags.

## How it works

The plugin discovers `.yaml` files in the suite directory, creates a pytest Item for each one, and calls `mcp-assert run --suite <file> --json` to execute it. The JSON result is mapped to pytest outcomes:

- `status: "pass"` becomes a pytest pass
- `status: "fail"` becomes a pytest failure with the detail as the message
- `status: "skip"` becomes a pytest skip
- `skip: true` in the YAML skips the test (for known bugs)

The Go binary handles all MCP protocol interaction. The plugin is a thin bridge.
