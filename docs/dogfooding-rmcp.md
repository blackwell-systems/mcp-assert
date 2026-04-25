# Dogfooding: rmcp (Rust MCP SDK)

Testing the most-starred Rust MCP SDK's example server. First Rust server in the mcp-assert suite collection.

## Target

- **Server:** Counter example from [4t145/rmcp](https://github.com/4t145/rmcp) (71 stars)
- **Language:** Rust
- **Transport:** stdio
- **Tools:** 6 (increment, decrement, get_value, say_hello, echo, sum)
- **Resources:** 2 (cwd path, memo)
- **Prompts:** 1 (example_prompt with argument)
- **Status:** Archived (March 2025). Superseded by [rust-mcp-stack/rust-mcp-sdk](https://github.com/rust-mcp-stack/rust-mcp-sdk) (171 stars).

## Process

### Step 1: Build (22 seconds)

```bash
git clone --depth 1 https://github.com/4t145/rmcp.git /tmp/rmcp
cd /tmp/rmcp && cargo build --example std_io
```

Compiled with 2 warnings (dead code in Calculator struct). Build time: 22s. No errors.

### Step 2: Auto-generate (instant)

```bash
mcp-assert generate --server "/tmp/rmcp/target/debug/examples/std_io" --output examples/rmcp-counter/
```

Output:
```
6 tools discovered, 6 assertions created, 0 skipped (already exist)
```

The generator correctly:
- Discovered all 6 tools from `tools/list`
- Set `a: 1, b: 1` for `sum` (integer params get default 1)
- Set `saying: TODO` for `echo` (string params get TODO placeholder)
- Left `increment`, `decrement`, `get_value`, `say_hello` with empty args (no params)

### Step 3: Run stubs (6/6 pass)

```bash
mcp-assert run --suite examples/rmcp-counter/
```

All 6 pass with `not_error: true`. This is the baseline: every tool responds without error.

### Step 4: Enhance assertions

Added specific expectations:
- `say_hello` contains "hello"
- `echo` contains the input string
- `sum(3, 7)` contains "10"
- `sum(-5, 3)` contains "-2" (negative numbers)
- `sum(0, 42)` contains "42" (zero identity)
- `increment` contains "1" (counter starts at 0)
- `decrement` contains "-1"
- `get_value` with setup: increment twice, expect "2"
- `get_value` idempotent: call twice, expect same value

Added protocol coverage:
- `resources/list` returns resources (contains "memo://insights")
- `resources/read` memo returns "Business Intelligence"
- `prompts/list` contains "example_prompt"
- `prompts/get` returns prompt template

### Step 5: Run enhanced suite (12/14 pass)

```
PASS  decrement returns decremented counter                          27ms
PASS  echo repeats input                                             24ms
PASS  echo handles empty string                                      23ms
FAIL  get_value returns current counter without modifying it         22ms
      expected result to contain "2", got: 1
FAIL  BUG: get_value mutates state (should be read-only)             22ms
      expected result to contain "2", got: 1
PASS  increment returns incremented counter                          21ms
PASS  prompts/get returns example prompt template                    20ms
PASS  prompts/list returns available prompts                         18ms
PASS  resources/list returns available resources                     18ms
PASS  resources/read returns memo content                            17ms
PASS  say_hello returns hello                                        18ms
PASS  sum adds two numbers                                           18ms
PASS  sum handles negative numbers                                   18ms
PASS  sum with zero returns other operand                            18ms

14 assertions, 12 passed, 2 failed
```

### Step 6: Coverage

```
Server exposes 6 tools, 6 have assertions (100% coverage)
```

## Bug found

### `get_value` decrements instead of reading

**Location:** `examples/servers/src/common/counter.rs`, line 49

```rust
#[tool(description = "Get the current counter value")]
async fn get_value(&self) -> Result<CallToolResult, McpError> {
    let mut counter = self.counter.lock().await;
    *counter -= 1;  // BUG: should not modify the counter
    Ok(CallToolResult::success(vec![Content::text(
        counter.to_string(),
    )]))
}
```

**Severity:** Logic bug in example code. Every developer learning from this example would copy a getter that mutates state.

**Impact:** The `get_value` tool is not idempotent. Calling it changes the counter value. An agent using this tool to "check" the counter would unknowingly decrement it each time.

**How mcp-assert caught it:** The assertion calls `increment` twice (counter = 2), then calls `get_value` and expects "2". It returns "1" because `get_value` decremented.

**Status:** Cannot file upstream issue (repo archived March 2025). Bug documented in the assertion suite.

## DX observations

1. **`generate` worked perfectly.** Zero friction from "here's a binary" to "here are 6 assertion stubs." The integer/string default heuristics were correct.
2. **Time to first bug: ~5 minutes.** Clone (22s) + generate (instant) + enhance assertions (3 min) + run (instant) = bug found.
3. **Rust builds are slower than Go.** 22s for a small example server. Go examples build in <2s. Not a mcp-assert issue, but worth noting for the Rust community.
4. **Archived repos are a dead end for issues.** We found a bug but can't file it. The suite still has value as documentation and as a Rust reference for mcp-assert users.
5. **Resources and prompts worked first try.** `assert_resources` and `assert_prompts` connected to the rmcp server without protocol issues. The Rust SDK implements these correctly.

## Summary

| Metric | Value |
|--------|-------|
| Time to first assertion | ~1 minute |
| Time to first bug | ~5 minutes |
| Total assertions | 14 |
| Tool coverage | 100% (6/6) |
| Protocol coverage | Tools, resources, prompts |
| Bugs found | 1 (get_value mutates state) |
| Upstream issues filed | 0 (repo archived) |
| Languages tested | 4 (Go, TypeScript, Python, Rust) |
