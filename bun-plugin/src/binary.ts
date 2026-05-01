import { existsSync } from "node:fs";
import { join, dirname } from "node:path";

/**
 * Find the mcp-assert binary using a 3-tier strategy:
 * 1. Explicit path (user-provided)
 * 2. PATH lookup (Bun.which)
 * 3. npm package (@blackwell-systems/mcp-assert)
 */
export function findBinary(explicit?: string): string | null {
  // 1. Explicit path
  if (explicit) {
    return explicit;
  }

  // 2. PATH lookup via Bun.which (native, no shell)
  const found = Bun.which("mcp-assert");
  if (found) return found;

  // 3. npm package
  try {
    const pkgPath = require.resolve("@blackwell-systems/mcp-assert/package.json");
    const binPath = join(dirname(pkgPath), "bin", "mcp-assert");
    if (existsSync(binPath)) return binPath;
  } catch {
    // Package not installed
  }

  return null;
}
