/**
 * jest-mcp-assert: Run mcp-assert YAML assertion files as Jest tests.
 *
 * Two usage modes:
 *
 * 1. Suite mode (auto-discover YAML files):
 *
 *    ```ts
 *    // mcp.test.ts
 *    import { describeMcpSuite } from 'jest-mcp-assert'
 *    describeMcpSuite('mcp server', 'evals/')
 *    ```
 *
 * 2. Single assertion mode:
 *
 *    ```ts
 *    // mcp.test.ts
 *    import { runMcpAssert } from 'jest-mcp-assert'
 *
 *    test('echo tool', () => { runMcpAssert('evals/echo.yaml') })
 *    test('greet tool', () => { runMcpAssert('evals/greet.yaml', { timeout: '60s' }) })
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
