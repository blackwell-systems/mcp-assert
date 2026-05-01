<?php

declare(strict_types=1);

namespace BlackwellSystems\McpAssert;

use RuntimeException;

/**
 * Runs mcp-assert on YAML assertion files and returns results.
 *
 * Usage in PHPUnit:
 *
 *     use BlackwellSystems\McpAssert\McpAssertRunner;
 *
 *     public function testEchoTool(): void
 *     {
 *         McpAssertRunner::assertYaml('evals/echo.yaml');
 *     }
 */
final class McpAssertRunner
{
    /**
     * Run a single YAML assertion file and assert it passes.
     *
     * @param string $yamlPath Path to the YAML assertion file
     * @param array<string, string> $options Optional settings: binary, timeout, fixture, server
     * @throws RuntimeException If the binary is not found or output cannot be parsed
     * @throws \PHPUnit\Framework\AssertionFailedError If the assertion fails
     */
    public static function assertYaml(string $yamlPath, array $options = []): McpAssertResult
    {
        $result = self::run($yamlPath, $options);

        if ($result->isSkipped()) {
            \PHPUnit\Framework\Assert::markTestSkipped($result->detail ?? 'skipped');
        }

        if ($result->isFailed()) {
            \PHPUnit\Framework\Assert::fail($result->detail ?? 'assertion failed');
        }

        \PHPUnit\Framework\Assert::assertTrue($result->isPassed());

        return $result;
    }

    /**
     * Run a single YAML assertion file and return the result without asserting.
     *
     * @param string $yamlPath Path to the YAML assertion file
     * @param array<string, string> $options Optional settings: binary, timeout, fixture, server
     */
    public static function run(string $yamlPath, array $options = []): McpAssertResult
    {
        $binary = BinaryFinder::find($options['binary'] ?? null);
        if ($binary === null) {
            throw new RuntimeException(
                'mcp-assert binary not found. Install via: '
                . 'brew install blackwell-systems/tap/mcp-assert, '
                . 'pip install mcp-assert, or '
                . 'go install github.com/blackwell-systems/mcp-assert@latest'
            );
        }

        $cmd = [
            $binary,
            'run',
            '--suite', $yamlPath,
            '--json',
            '--timeout', $options['timeout'] ?? '30s',
        ];

        if (isset($options['fixture'])) {
            $cmd[] = '--fixture';
            $cmd[] = $options['fixture'];
        }
        if (isset($options['server'])) {
            $cmd[] = '--server';
            $cmd[] = $options['server'];
        }

        $process = proc_open(
            $cmd,
            [
                0 => ['pipe', 'r'],
                1 => ['pipe', 'w'],
                2 => ['pipe', 'w'],
            ],
            $pipes
        );

        if (!is_resource($process)) {
            throw new RuntimeException('Failed to start mcp-assert process');
        }

        fclose($pipes[0]);
        $stdout = stream_get_contents($pipes[1]);
        $stderr = stream_get_contents($pipes[2]);
        fclose($pipes[1]);
        fclose($pipes[2]);
        $exitCode = proc_close($process);

        if ($stdout === '' || $stdout === false) {
            if ($exitCode !== 0) {
                throw new RuntimeException(
                    "mcp-assert failed (exit {$exitCode}): " . trim((string) $stderr)
                );
            }
            throw new RuntimeException('mcp-assert returned no output');
        }

        $results = json_decode($stdout, true);
        if (!is_array($results) || empty($results)) {
            throw new RuntimeException(
                'Could not parse mcp-assert output: ' . substr((string) $stdout, 0, 500)
            );
        }

        return McpAssertResult::fromArray($results[0]);
    }

    /**
     * Discover all YAML files in a directory and run each as a test.
     * Returns an array of results keyed by filename.
     *
     * @param string $suiteDir Directory containing YAML assertion files
     * @param array<string, string> $options Optional settings
     * @return array<string, McpAssertResult>
     */
    public static function suite(string $suiteDir, array $options = []): array
    {
        $files = glob($suiteDir . '/*.yaml') ?: [];
        $ymlFiles = glob($suiteDir . '/*.yml') ?: [];
        $files = array_merge($files, $ymlFiles);
        sort($files);

        if (empty($files)) {
            throw new RuntimeException("No YAML files found in {$suiteDir}");
        }

        $results = [];
        foreach ($files as $file) {
            $name = pathinfo($file, PATHINFO_FILENAME);
            $results[$name] = self::run($file, $options);
        }

        return $results;
    }
}
