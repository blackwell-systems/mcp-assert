<?php

declare(strict_types=1);

namespace BlackwellSystems\McpAssert;

/**
 * Locates the mcp-assert binary.
 *
 * Resolution order:
 * 1. Explicit path (user-provided)
 * 2. PATH lookup (which mcp-assert)
 * 3. Common install locations
 */
final class BinaryFinder
{
    public static function find(?string $explicit = null): ?string
    {
        // 1. Explicit path
        if ($explicit !== null && file_exists($explicit)) {
            return $explicit;
        }

        // 2. PATH lookup
        $which = trim((string) shell_exec('which mcp-assert 2>/dev/null'));
        if ($which !== '' && file_exists($which)) {
            return $which;
        }

        // Windows fallback
        $where = trim((string) shell_exec('where mcp-assert 2>nul'));
        if ($where !== '') {
            $first = explode("\n", $where)[0];
            if (file_exists(trim($first))) {
                return trim($first);
            }
        }

        // 3. Common locations
        $home = getenv('HOME') ?: getenv('USERPROFILE') ?: '';
        $candidates = [
            $home . '/go/bin/mcp-assert',
            '/usr/local/bin/mcp-assert',
            '/opt/homebrew/bin/mcp-assert',
        ];
        foreach ($candidates as $path) {
            if (file_exists($path)) {
                return $path;
            }
        }

        return null;
    }
}
