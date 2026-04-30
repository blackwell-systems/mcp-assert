# mcpassert

Go test helpers for [mcp-assert](https://github.com/blackwell-systems/mcp-assert). Run MCP server assertion YAML files as Go subtests.

Same YAML files work across Go test, Jest, Vitest, pytest, and the CLI.

## Install

```bash
go get github.com/blackwell-systems/mcp-assert/go-plugin
```

Also install the mcp-assert binary:

```bash
go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest
```

## Usage

### Single assertion

```go
func TestEchoTool(t *testing.T) {
    mcpassert.Run(t, "evals/echo.yaml")
}
```

### Suite (auto-discover all YAML files in a directory)

```go
func TestMCPServer(t *testing.T) {
    mcpassert.Suite(t, "evals/")
}
```

Each YAML file becomes a `t.Run` subtest. Run a single one with `go test -run TestMCPServer/echo`.

### Options

```go
func TestWithOptions(t *testing.T) {
    mcpassert.Run(t, "evals/echo.yaml", mcpassert.Options{
        Timeout: "60s",
        Fixture: "testdata/fixtures",
        Server:  "go run ./cmd/server",
    })
}
```

## How it works

mcpassert shells out to the mcp-assert binary with `--json` and maps results to `t.Error` (FAIL), `t.Skip` (SKIP), or pass (PASS). The Go binary handles all MCP protocol logic.

Binary resolution (in order):
1. Explicit `Binary` option
2. `mcp-assert` on PATH
3. `$HOME/go/bin/mcp-assert`

## Writing assertions

Assertions are YAML files. See the [mcp-assert docs](https://blackwell-systems.github.io/mcp-assert) for the full reference.

```yaml
name: echo returns the input message
server:
  command: go
  args: ["run", "./cmd/server"]
assert:
  tool: echo
  args:
    message: "hello"
  expect:
    not_error: true
    contains: ["hello"]
```

## License

MIT
