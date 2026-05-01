/**
 * bun-mcp-assert: Run mcp-assert YAML assertion files as Bun tests.
 *
 * Two usage modes:
 *
 * 1. Suite mode (auto-discover YAML files):
 *
 *    ```ts
 *    // mcp.test.ts
 *    import { describeMcpSuite } from "bun-mcp-assert"
 *    describeMcpSuite("mcp server", "evals/")
 *    ```
 *
 * 2. Single assertion mode:
 *
 *    ```ts
 *    // mcp.test.ts
 *    import { test } from "bun:test"
 *    import { runMcpAssert } from "bun-mcp-assert"
 *
 *    test("echo tool", () => { runMcpAssert("evals/echo.yaml") })
 *    ```
 */

export { runMcpAssert, McpAssertError } from "./runner";
export { describeMcpSuite } from "./suite";
export { findBinary } from "./binary";
export type {
  McpAssertResult,
  McpSuiteOptions,
  RunOptions,
} from "./types";
