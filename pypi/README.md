# mcp-assert

The testing standard for MCP servers. Lint, test, and fuzz. No SDK required. No LLM required.

This is the PyPI distribution of [mcp-assert](https://github.com/blackwell-systems/mcp-assert), a single Go binary that connects to your MCP server over stdio, SSE, or HTTP, runs 24 static analysis rules, calls tools with real arguments, and asserts results against YAML expectations.

## Install

```bash
pip install mcp-assert
```

## Usage

```bash
# Zero-config audit (finds crashes)
mcp-assert audit --server "python my_server.py"

# Static analysis (finds schema issues)
mcp-assert lint --server "python my_server.py"

# Auto-fix schema issues
mcp-assert lint --server "python my_server.py" --fix

# Adversarial fuzzing
mcp-assert fuzz --server "python my_server.py"

# Run YAML assertions
mcp-assert run --suite evals/

# CI with threshold
mcp-assert ci --suite evals/ --threshold 95
```

Or use as a Python module:

```bash
python -m mcp_assert lint --server "python my_server.py"
```

## pytest Integration

```bash
pip install pytest-mcp-assert
pytest --mcp-suite evals/
```

Each YAML file becomes a pytest Item. Configure via `pyproject.toml`:

```toml
[tool.pytest.ini_options]
mcp_suite = "evals/"
mcp_fixture = "fixtures/"
```

## Lint (24 Rules)

```bash
mcp-assert lint --server "python my_server.py"

  E  E103   create_user    Required parameter "email" has no description
  W  W109   search         User-facing parameter "query" has no examples
  W  W108   delete_user    Tool name implies mutation but description doesn't acknowledge

# Auto-generate fixes
mcp-assert lint --server "python my_server.py" --fix

# Strict mode for CI
mcp-assert lint --server "python my_server.py" --strict
```

## Links

- [Documentation](https://blackwell-systems.github.io/mcp-assert/)
- [GitHub](https://github.com/blackwell-systems/mcp-assert)
- [Error Reference](https://github.com/blackwell-systems/mcp-assert/blob/main/docs/ERROR_REFERENCE.md) (24 error codes)
- [Scorecard](https://blackwell-systems.github.io/mcp-assert/scorecard/) (102 servers scanned, 4,794 issues found)
- [pytest plugin](https://pypi.org/project/pytest-mcp-assert/)

## License

MIT
