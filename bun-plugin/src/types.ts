/** Result from a single mcp-assert assertion run. */
export interface McpAssertResult {
  name: string;
  status: "PASS" | "FAIL" | "SKIP";
  detail?: string;
  duration_ms?: number;
  trial?: number;
}

/** Options for running a single assertion. */
export interface RunOptions {
  /** Path to the mcp-assert binary. Auto-detected if omitted. */
  binary?: string;
  /** Per-assertion timeout (e.g. "30s", "1m"). Default: "30s". */
  timeout?: string;
  /** Fixture directory for {{fixture}} substitution. */
  fixture?: string;
  /** Server override (e.g. "npx my-server"). */
  server?: string;
}

/** Options for describeMcpSuite. */
export interface McpSuiteOptions extends RunOptions {
  /** Glob pattern for YAML files within the suite directory. Default: "*.yaml" */
  pattern?: string;
}
