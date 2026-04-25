# Spring AI MCP Server (Kotlin) Examples

These assertions test [jamesward/hello-spring-mcp-server](https://github.com/jamesward/hello-spring-mcp-server), a Spring Boot MCP server written in Kotlin. First JVM language (Kotlin/Java) in the mcp-assert suite collection.

## Setup

```bash
git clone --depth 1 https://github.com/jamesward/hello-spring-mcp-server.git /tmp/hello-spring-mcp
cd /tmp/hello-spring-mcp && ./gradlew bootRun
```

The server starts on port 8080. Then run assertions via HTTP transport:

```bash
mcp-assert run --suite examples/spring-mcp
```

## Coverage

3 assertions covering 2/2 tools (100%):

| Tool | Assertions | Notes |
|------|------------|-------|
| `getSkills` | 1 | Returns all employee skills |
| `getEmployeesWithSkill` | 2 | Matching skill ("java") and unknown skill ("cobol") |

## Notes

Clean scan: no bugs found. Uses HTTP transport (streamable HTTP). First JVM server tested by mcp-assert, bringing language coverage to 5 (Go, TypeScript, Python, Rust, Kotlin/Java).
