# Contributing to mcp-assert

Thank you for your interest in contributing! This guide will help you understand the codebase structure and how to add new features.

## Getting Started

```bash
git clone https://github.com/blackwell-systems/mcp-assert.git
cd mcp-assert
go test ./... -race  # Run all tests
```

## Project Structure

```
cmd/mcp-assert/
  main.go           CLI entry point, command dispatch

internal/assertion/
  types.go          Core types (Suite, Assertion, Expect, Result, all block types)
  loader.go         YAML file loading with subdirectory recursion
  checker.go        18 assertion type implementations
  trajectory.go     4 trajectory assertion types
  logging*.go       Logging assertion block and checker
  sampling*.go      Sampling assertion block type

internal/runner/
  commands.go       Run, CI, Matrix (CLI entry points)
  client.go         MCP client creation (stdio, SSE, HTTP, capabilities)
  execute.go        Assertion execution and routing
  substitute.go     Template substitution ({{fixture}}, ${VAR}, captured variables)
  fixture.go        Per-assertion fixture isolation
  generate.go       Auto-generate stubs from tools/list
  snapshot.go       Snapshot capture/compare
  coverage.go       Coverage command
  watch.go          Watch mode
  intercept.go      Intercept command (stdio proxy, trajectory validation)
  fix.go            --fix mode (position scanning, YAML patch generation)
  util.go           Shared utilities

internal/report/
  report.go         Terminal output (color-aware)
  junit.go          JUnit XML generation
  markdown.go       GitHub Step Summary
  badge.go          shields.io endpoint JSON
  reliability.go    pass@k / pass^k computation
  baseline.go       Baseline write/load, regression detection
  coverage.go       Coverage JSON serialization
  snapshot.go       Snapshot file read/write/compare
  diff.go           Unified diff for watch mode
```

## Adding a New Assertion Type

1. **Add to `internal/assertion/types.go`:**
   
   Add a field to the `Expect` struct:
   ```go
   type Expect struct {
       // existing fields...
       MyNewCheck string `yaml:"my_new_check,omitempty"`
   }
   ```

2. **Implement checker in `internal/assertion/checker.go`:**
   
   Add a case to the `Check` function:
   ```go
   if expect.MyNewCheck != "" {
       if !strings.Contains(result, expect.MyNewCheck) {
           return fmt.Errorf("expected to find '%s' but didn't", expect.MyNewCheck)
       }
   }
   ```

3. **Add tests in `internal/assertion/checker_test.go`:**
   
   ```go
   func TestMyNewCheck(t *testing.T) {
       result := assertion.Result{Content: "hello world"}
       expect := assertion.Expect{MyNewCheck: "hello"}
       err := assertion.Check(expect, result)
       if err != nil {
           t.Errorf("expected pass, got: %v", err)
       }
   }
   ```

4. **Document in `docs/writing-assertions.md`:**
   
   Add an entry to the assertion types table and a usage example.

5. **Update `FEATURES.md`:**
   
   Increment the assertion type count and add a row to the table.

## Adding a New CLI Command

1. **Add command dispatch in `cmd/mcp-assert/main.go`:**
   
   ```go
   case "mycommand":
       if err := runner.MyCommand(os.Args[2:]); err != nil {
           fmt.Fprintf(os.Stderr, "error: %v\n", err)
           os.Exit(1)
       }
   ```

2. **Implement in `internal/runner/commands.go`:**
   
   ```go
   func MyCommand(args []string) error {
       // Parse flags with flag.NewFlagSet
       // Implement command logic
       return nil
   }
   ```

3. **Add to usage in `cmd/mcp-assert/main.go`:**
   
   Update `printUsage()` to include the new command.

4. **Document in `docs/cli.md`:**
   
   Add a section with usage, flags, and examples.

5. **Add tests in `internal/runner/`:**
   
   Create `mycommand_test.go` if the logic is complex enough to warrant dedicated tests.

## Adding a New Block Type

Block types are alternatives to `assert:` (like `assert_prompts:`, `assert_resources:`, etc.).

1. **Add to `internal/assertion/types.go`:**
   
   ```go
   type MyNewBlock struct {
       Field1 string `yaml:"field1"`
       Field2 int    `yaml:"field2"`
   }
   
   // Add to Assertion struct:
   type Assertion struct {
       // existing fields...
       AssertMyNew *MyNewBlock `yaml:"assert_my_new,omitempty"`
   }
   ```

2. **Add routing in `internal/runner/execute.go`:**
   
   Add a case to `runAssertion` that detects the new block and calls a dedicated handler:
   ```go
   if a.AssertMyNew != nil {
       return runMyNewAssertion(ctx, client, a, capturedVars)
   }
   ```

3. **Implement handler:**
   
   Create `internal/runner/mynew.go` with the handler function.

4. **Add tests:**
   
   Add tests in `internal/runner/mynew_test.go` (or `runner_test.go` if small).

5. **Document:**
   
   Add to `docs/writing-assertions.md` with YAML examples.

## Running Tests

```bash
# All tests with race detector
GOWORK=off go test ./... -race

# Specific package
GOWORK=off go test ./internal/assertion -v

# Run with coverage
GOWORK=off go test ./... -cover
```

## Code Style

- Follow standard Go conventions (gofmt, golint)
- Package-level documentation at the top of the canonical file (types.go, report.go, etc.)
- Keep functions focused (single responsibility)
- Write tests for new assertion types, CLI commands, and core logic

## Documentation

When adding features visible to users:

1. Update relevant doc pages in `docs/`
2. Update `FEATURES.md` if adding a command, assertion type, or output format
3. Update `CHANGELOG.md` under `[Unreleased]`
4. Add example YAMLs in `examples/` if demonstrating a new assertion pattern

## Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-new-feature`)
3. Make your changes with tests
4. Run the full test suite (`GOWORK=off go test ./... -race`)
5. Commit with clear messages
6. Push to your fork and open a PR

We prefer small, focused PRs over large rewrites. If you're planning a major change, open an issue first to discuss the design.

## Questions?

Open an issue or start a discussion on GitHub. We're happy to help!
