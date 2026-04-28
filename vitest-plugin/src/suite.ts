import { readFileSync, readdirSync, statSync } from "node:fs";
import { join, resolve } from "node:path";
import { describe, test } from "vitest";
import { runMcpAssert } from "./runner.js";
import type { RunOptions } from "./types.js";

/**
 * Discover all YAML assertion files in a directory and register them as
 * Vitest tests inside a describe() block.
 *
 * Usage:
 *   import { describeMcpSuite } from 'vitest-mcp-assert'
 *   describeMcpSuite('mcp assertions', 'evals/')
 *   describeMcpSuite('mcp assertions', 'evals/', { timeout: '60s' })
 */
export function describeMcpSuite(
  suiteName: string,
  suiteDir: string,
  options?: RunOptions
): void {
  const dir = resolve(suiteDir);
  const files = collectYamlFiles(dir);

  describe(suiteName, () => {
    for (const file of files) {
      const name = extractName(file);
      test(name, () => {
        runMcpAssert(file, options);
      });
    }
  });
}

/** Recursively collect .yaml files from a directory. */
function collectYamlFiles(dir: string): string[] {
  const results: string[] = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const stat = statSync(full);
    if (stat.isDirectory()) {
      results.push(...collectYamlFiles(full));
    } else if (entry.endsWith(".yaml") || entry.endsWith(".yml")) {
      results.push(full);
    }
  }
  return results.sort();
}

/** Extract the assertion name from a YAML file's name: field, falling back to the filename. */
function extractName(filePath: string): string {
  try {
    const content = readFileSync(filePath, "utf-8");
    for (const line of content.split("\n")) {
      const trimmed = line.trim();
      if (trimmed.startsWith("name:")) {
        return trimmed.slice(5).trim().replace(/^["']|["']$/g, "");
      }
    }
  } catch {
    // Fall through to filename
  }
  const parts = filePath.split("/");
  return parts[parts.length - 1].replace(/\.ya?ml$/, "");
}
