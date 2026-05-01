import { readdirSync } from "node:fs";
import { join } from "node:path";
import { describe, test, expect } from "bun:test";
import { runMcpAssert } from "./runner";
import type { McpSuiteOptions } from "./types";

/**
 * Auto-discover YAML assertion files in a directory and register each
 * as a Bun test case.
 *
 * Usage:
 *   import { describeMcpSuite } from "bun-mcp-assert"
 *   describeMcpSuite("my mcp server", "evals/")
 */
export function describeMcpSuite(
  name: string,
  suiteDir: string,
  options?: McpSuiteOptions
): void {
  const pattern = options?.pattern ?? "*.yaml";
  const suffix = pattern.replace("*", "");

  const files = readdirSync(suiteDir)
    .filter((f) => f.endsWith(suffix))
    .sort();

  if (files.length === 0) {
    throw new Error(`No YAML files matching ${pattern} found in ${suiteDir}`);
  }

  describe(name, () => {
    for (const file of files) {
      const testName = file.replace(/\.ya?ml$/, "");
      const yamlPath = join(suiteDir, file);

      test(testName, () => {
        const result = runMcpAssert(yamlPath, options);

        if (result.status === "SKIP") {
          console.log(`[SKIP] ${testName}: ${result.detail ?? "skipped"}`);
          return;
        }

        expect(result.status).toBe("PASS");
      });
    }
  });
}
