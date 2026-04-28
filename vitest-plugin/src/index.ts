/**
 * vitest-mcp-assert: Run mcp-assert YAML assertion files as Vitest tests.
 *
 * Two usage modes:
 *
 * 1. Suite mode (auto-discover YAML files):
 *
 *    ```ts
 *    // mcp.test.ts
 *    import { describeMcpSuite } from 'vitest-mcp-assert'
 *    describeMcpSuite('mcp server', 'evals/')
 *    ```
 *
 * 2. Single assertion mode:
 *
 *    ```ts
 *    // mcp.test.ts
 *    import { test } from 'vitest'
 *    import { runMcpAssert } from 'vitest-mcp-assert'
 *
 *    test('echo tool', () => runMcpAssert('evals/echo.yaml'))
 *    test('greet tool', () => runMcpAssert('evals/greet.yaml', { timeout: '60s' }))
 *    ```
 */

export { runMcpAssert, McpAssertError } from "./runner.js";
export { describeMcpSuite } from "./suite.js";
export { findBinary } from "./binary.js";
export type {
  McpAssertResult,
  McpAssertPluginOptions,
  RunOptions,
} from "./types.js";
