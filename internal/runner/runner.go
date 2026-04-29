// Package runner implements the mcp-assert test runner.
//
// Execution flow:
//  1. CLI (main.go) dispatches to Run, CI, or Matrix in commands.go.
//  2. commands.go loads the Suite, iterates assertions, and calls runAssertion
//     (execute.go) for each one.
//  3. runAssertion creates a fresh MCP client (client.go), performs the
//     initialize handshake, runs setup steps, calls the tool, and evaluates
//     the Expect block via assertion.Check (checker.go).
//  4. Results are collected and passed to the report package for output.
//
// The package is split across focused files:
//   - commands.go   — Run, CI, Matrix (CLI entry points that own flag parsing and the run loop)
//   - client.go     — MCP client creation (stdio, SSE, HTTP, Docker wrapping, mock capabilities)
//   - execute.go    — runAssertion and its sub-dispatchers (resource, prompt, completion, trajectory)
//   - substitute.go — template substitution ({{fixture}}, captured variables, JSON path)
//   - fixture.go    — per-assertion fixture isolation (copyDir, isolateFixture)
//   - util.go       — writeReports, applyServerOverride, countFails, countPasses, extractText
//   - snapshot.go   — Snapshot command and runAndCapture
//   - coverage.go   — Coverage command
//   - generate.go   — Generate command
//   - watch.go      — Watch command
//   - audit.go      — Zero-config quality audit
//   - intercept.go  — Stdio proxy for trajectory capture
package runner
