import { execFileSync } from "node:child_process";
import { findBinary } from "./binary.js";
import type { McpAssertResult, RunOptions } from "./types.js";

/**
 * Run mcp-assert on a single YAML assertion file and return the result.
 * Throws on failure, calls Vitest's skip on SKIP status.
 *
 * Usage:
 *   import { test } from 'vitest'
 *   import { runMcpAssert } from 'vitest-mcp-assert'
 *
 *   test('echo returns message', () => runMcpAssert('evals/echo.yaml'))
 */
export function runMcpAssert(yamlPath: string, options?: RunOptions): void {
  const binary = findBinary(options?.binary);
  if (!binary) {
    // Dynamic import to avoid pulling vitest into the module graph at parse time.
    // This module may be imported outside vitest (e.g., for type checking).
    throw new Error(
      "mcp-assert binary not found. Install via: " +
        "brew install blackwell-systems/tap/mcp-assert, " +
        "npm install @blackwell-systems/mcp-assert, or " +
        "go install github.com/blackwell-systems/mcp-assert@latest"
    );
  }

  const cmd = [
    "run",
    "--suite",
    yamlPath,
    "--json",
    "--timeout",
    options?.timeout ?? "30s",
  ];

  if (options?.fixture) {
    cmd.push("--fixture", options.fixture);
  }
  if (options?.server) {
    cmd.push("--server", options.server);
  }

  let stdout: string;
  try {
    stdout = execFileSync(binary, cmd, {
      encoding: "utf-8",
      timeout: 120_000,
      stdio: ["pipe", "pipe", "pipe"],
    });
  } catch (err: unknown) {
    const execErr = err as { stdout?: string; stderr?: string; status?: number };
    // mcp-assert exits non-zero on assertion failure but still writes JSON to stdout
    if (execErr.stdout) {
      stdout = execErr.stdout;
    } else {
      throw new Error(
        `mcp-assert failed (exit ${execErr.status}): ${execErr.stderr?.trim() ?? "unknown error"}`
      );
    }
  }

  let results: McpAssertResult[];
  try {
    results = JSON.parse(stdout);
  } catch {
    throw new Error(`Could not parse mcp-assert output: ${stdout.slice(0, 500)}`);
  }

  if (!results || results.length === 0) {
    throw new Error("mcp-assert returned no results");
  }

  const r = results[0];

  if (r.status === "SKIP") {
    // Use a special error that vitest-mcp-assert plugin or the caller can catch.
    // In direct usage, the caller should handle this with test.skip() or
    // wrap with describeMcpSuite which handles it automatically.
    const err = new Error(r.detail ?? "skipped");
    err.name = "McpAssertSkip";
    throw err;
  }

  if (r.status === "FAIL") {
    throw new McpAssertError(r.detail ?? "assertion failed", r);
  }

  // PASS: do nothing, test succeeds
}

/** Error thrown when an mcp-assert assertion fails. */
export class McpAssertError extends Error {
  public result: McpAssertResult;

  constructor(message: string, result: McpAssertResult) {
    super(message);
    this.name = "McpAssertError";
    this.result = result;
  }
}
