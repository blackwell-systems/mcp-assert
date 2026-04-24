# mcp-assert

Deterministic correctness testing for MCP servers. No SDK required. No LLM required.

This is the PyPI distribution of [mcp-assert](https://github.com/blackwell-systems/mcp-assert), a single Go binary that connects to your MCP server over stdio, SSE, or HTTP, calls your tools, and asserts the results.

## Install

```bash
pip install mcp-assert
```

## Usage

```bash
# Scaffold your first assertion
mcp-assert init evals

# Run it
mcp-assert run --suite evals/ --fixture evals/fixtures
```

Or use as a Python module:

```bash
python -m mcp_assert run --suite evals/
```

## Documentation

Full documentation: [blackwell-systems.github.io/mcp-assert](https://blackwell-systems.github.io/mcp-assert)

## License

MIT
