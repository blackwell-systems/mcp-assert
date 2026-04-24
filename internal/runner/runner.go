// Package runner implements the mcp-assert test runner.
//
// The package is split across focused files:
//   - commands.go  — Run, CI, Matrix (CLI entry points)
//   - client.go    — MCP client creation (stdio, SSE, HTTP, capabilities)
//   - execute.go   — runAssertion and its sub-dispatchers (resource, prompt, trajectory)
//   - substitute.go — template substitution ({{fixture}}, captured variables, JSON path)
//   - fixture.go   — per-assertion fixture isolation (copyDir, isolateFixture)
//   - util.go      — writeReports, applyServerOverride, countFails, countPasses, extractText
//   - snapshot.go  — Snapshot command and runAndCapture
//   - coverage.go  — Coverage command
//   - generate.go  — Generate command
//   - watch.go     — Watch command
package runner
