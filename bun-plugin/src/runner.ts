import { findBinary } from "./binary";
import type { McpAssertResult, RunOptions } from "./types";

/**
 * Run mcp-assert on a single YAML assertion file.
 * Throws on failure. Returns the result on success.
 *
 * Usage:
 *   import { test } from "bun:test"
 *   import { runMcpAssert } from "bun-mcp-assert"
 *
 *   test("echo returns message", () => {
 *     runMcpAssert("evals/echo.yaml")
 *   })
 */
export function runMcpAssert(yamlPath: string, options?: RunOptions): McpAssertResult {
  const binary = findBinary(options?.binary);
  if (!binary) {
    throw new Error(
      "mcp-assert binary not found. Install via: " +
        "brew install blackwell-systems/tap/mcp-assert, " +
        "bun add @blackwell-systems/mcp-assert, or " +
        "go install github.com/blackwell-systems/mcp-assert@latest"
    );
  }

  const cmd = [
    binary,
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

  const proc = Bun.spawnSync(cmd, {
    stdout: "pipe",
    stderr: "pipe",
    timeout: 120_000,
  });

  const stdout = proc.stdout.toString();
  const stderr = proc.stderr.toString();

  if (!stdout && proc.exitCode !== 0) {
    throw new Error(
      `mcp-assert failed (exit ${proc.exitCode}): ${stderr.trim() || "unknown error"}`
    );
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
    return r;
  }

  if (r.status === "FAIL") {
    throw new McpAssertError(r.detail ?? "assertion failed", r);
  }

  return r;
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
