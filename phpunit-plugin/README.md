# phpunit-mcp-assert

PHPUnit integration for [mcp-assert](https://github.com/blackwell-systems/mcp-assert). Run MCP server assertion YAML files as PHPUnit tests.

Same YAML files work across PHPUnit, Bun, Jest, Vitest, pytest, Go test, and the CLI.

## Install

```bash
composer require --dev blackwell-systems/phpunit-mcp-assert
```

Also install the mcp-assert binary:

```bash
brew install blackwell-systems/tap/mcp-assert
# or: pip install mcp-assert
# or: npm install @blackwell-systems/mcp-assert
```

## Usage

### Single assertion

```php
use BlackwellSystems\McpAssert\McpAssertRunner;
use PHPUnit\Framework\TestCase;

class McpServerTest extends TestCase
{
    public function testEchoTool(): void
    {
        McpAssertRunner::assertYaml('evals/echo.yaml');
    }

    public function testWithOptions(): void
    {
        McpAssertRunner::assertYaml('evals/search.yaml', [
            'timeout' => '60s',
            'fixture' => 'tests/fixtures',
            'server' => 'php artisan mcp:serve',
        ]);
    }
}
```

### Suite (all YAML files in a directory)

```php
public function testAllAssertions(): void
{
    $results = McpAssertRunner::suite('evals/');
    foreach ($results as $name => $result) {
        $this->assertTrue($result->isPassed(), "{$name}: {$result->detail}");
    }
}
```

### Data provider pattern

```php
public static function assertionProvider(): array
{
    $files = glob('evals/*.yaml') ?: [];
    $cases = [];
    foreach ($files as $file) {
        $name = pathinfo($file, PATHINFO_FILENAME);
        $cases[$name] = [$file];
    }
    return $cases;
}

#[DataProvider('assertionProvider')]
public function testAssertion(string $yamlPath): void
{
    McpAssertRunner::assertYaml($yamlPath);
}
```

## How it works

phpunit-mcp-assert shells out to the mcp-assert Go binary with `--json`, parses the result, and maps it to PHPUnit assertions. The Go binary handles all MCP protocol logic.

Binary resolution (in order):
1. Explicit `binary` option
2. `which mcp-assert` (PATH lookup)
3. Common install locations (`~/go/bin`, `/usr/local/bin`, Homebrew)

## License

MIT
