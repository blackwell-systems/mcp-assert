<?php

declare(strict_types=1);

namespace BlackwellSystems\McpAssert;

/**
 * Result from a single mcp-assert assertion run.
 */
final class McpAssertResult
{
    public function __construct(
        public readonly string $name,
        public readonly string $status,
        public readonly ?string $detail = null,
        public readonly ?int $durationMs = null,
        public readonly ?int $trial = null,
    ) {}

    public function isPassed(): bool
    {
        return $this->status === 'PASS';
    }

    public function isFailed(): bool
    {
        return $this->status === 'FAIL';
    }

    public function isSkipped(): bool
    {
        return $this->status === 'SKIP';
    }

    /**
     * @param array<string, mixed> $data
     */
    public static function fromArray(array $data): self
    {
        return new self(
            name: $data['name'] ?? '',
            status: $data['status'] ?? 'FAIL',
            detail: $data['detail'] ?? null,
            durationMs: $data['duration_ms'] ?? null,
            trial: $data['trial'] ?? null,
        );
    }
}
