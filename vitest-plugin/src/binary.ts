import { execFileSync } from "node:child_process";
import { existsSync } from "node:fs";
import { join, dirname } from "node:path";
import { createRequire } from "node:module";

/**
 * Find the mcp-assert binary using a 3-tier strategy:
 * 1. Explicit path (user-provided)
 * 2. PATH lookup (which mcp-assert)
 * 3. npm package (@blackwell-systems/mcp-assert)
 */
export function findBinary(explicit?: string): string | null {
  // 1. Explicit path
  if (explicit) {
    return explicit;
  }

  // 2. PATH lookup
  try {
    const result = execFileSync("which", ["mcp-assert"], {
      encoding: "utf-8",
      stdio: ["pipe", "pipe", "pipe"],
    }).trim();
    if (result) return result;
  } catch {
    // Not found in PATH, try next strategy
  }

  // Windows fallback
  try {
    const result = execFileSync("where", ["mcp-assert"], {
      encoding: "utf-8",
      stdio: ["pipe", "pipe", "pipe"],
    }).trim();
    if (result) return result.split("\n")[0].trim();
  } catch {
    // Not found, try next strategy
  }

  // 3. npm package
  try {
    const require = createRequire(import.meta.url);
    const pkgPath = require.resolve("@blackwell-systems/mcp-assert/package.json");
    const binPath = join(dirname(pkgPath), "bin", "mcp-assert");
    if (existsSync(binPath)) return binPath;
  } catch {
    // Package not installed
  }

  return null;
}
