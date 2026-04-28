/** Result from a single mcp-assert assertion run. */
export interface McpAssertResult {
  name: string;
  status: "PASS" | "FAIL" | "SKIP";
  detail?: string;
  duration_ms: number;
  language?: string;
  trial?: number;
}

/** Options for the vitest-mcp-assert plugin. */
export interface McpAssertPluginOptions {
  /** Directory containing mcp-assert YAML assertion files. */
  suite: string;
  /** Fixture directory (substituted for {{fixture}} in assertions). */
  fixture?: string;
  /** Override server command for all assertions. */
  server?: string;
  /** Per-assertion timeout (default: "30s"). */
  timeout?: string;
  /** Path to mcp-assert binary (default: auto-detect). */
  binary?: string;
}

/** Options for a single runMcpAssert() call. */
export interface RunOptions {
  /** Fixture directory. */
  fixture?: string;
  /** Override server command. */
  server?: string;
  /** Per-assertion timeout (default: "30s"). */
  timeout?: string;
  /** Path to mcp-assert binary (default: auto-detect). */
  binary?: string;
}
