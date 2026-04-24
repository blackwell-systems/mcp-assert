# mcp-go Example Servers Setup

The mcp-go example suites test various servers from the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) SDK. These assertions reference pre-built binaries in `/tmp/`. You have two options for setup:

## Option 1: Build from source (recommended)

Clone the mcp-go repository and build the example servers:

```bash
git clone --depth 1 https://github.com/mark3labs/mcp-go.git /tmp/mcp-go
cd /tmp/mcp-go

# Build the servers you need
go build -o /tmp/mcp-go-everything-server ./examples/everything
go build -o /tmp/mcp-go-sampling-server-v049 ./examples/sampling_server
go build -o /tmp/mcp-go-roots-server-v049 ./examples/roots_server
go build -o /tmp/mcp-go-elicitation-server-v049 ./examples/elicitation
```

Then run the assertions:

```bash
mcp-assert run --suite examples/mcp-go-everything
mcp-assert run --suite examples/mcp-go-sampling
mcp-assert run --suite examples/mcp-go-roots --fixture examples/mcp-go-roots/fixtures
mcp-assert run --suite examples/mcp-go-elicitation
# ... and so on for other mcp-go suites
```

## Option 2: Override with --server

If you built the servers elsewhere or want to use different names, override the command:

```bash
mcp-assert run --suite examples/mcp-go-sampling \
  --server "/path/to/your/sampling-server"
```

## Affected Suites

The following example suites use mcp-go servers with `/tmp/` paths:

- `examples/mcp-go-everything/`
- `examples/mcp-go-everything-http/`
- `examples/mcp-go-everything-prompts/`
- `examples/mcp-go-everything-resources/`
- `examples/mcp-go-everything-completion/`
- `examples/mcp-go-everything-logging/`
- `examples/mcp-go-sampling/`
- `examples/mcp-go-roots/`
- `examples/mcp-go-elicitation/`
