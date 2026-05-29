# Error Reference

Structured error codes used by mcp-assert for consistent issue classification.

---

## Success Codes

### E000: Success
**Category:** Execution  
**Severity:** Info

Tool executed successfully with no errors.

**Example:**
```
✓ read_query      1ms  [E000] responds, returns content
```

---

## Runtime Errors (E2xx)

### E201: Server Panic
**Category:** Runtime  
**Severity:** Error

Server crashed or returned MCP internal error (-32603). Indicates unhandled exception or panic in server code.

**Common Causes:**
- Null pointer dereference
- Unhandled exceptions
- Missing error handling
- Stack overflow

**Example:**
```
✗ create_table    0ms  [E201] internal error: panic: nil pointer dereference
```

**Remediation:**
1. Add error handling for edge cases
2. Validate all inputs before processing
3. Use proper null checks
4. Add recovery mechanisms for panics

---

### E202: Timeout
**Category:** Runtime  
**Severity:** Error

Tool did not respond within the configured timeout period.

**Common Causes:**
- Infinite loops
- Blocking I/O operations
- Network delays
- Resource exhaustion

**Example:**
```
✗ slow_query     30s  [E202] timed out after 30s
```

**Remediation:**
1. Optimize tool performance
2. Add timeout handling in tool code
3. Use async/non-blocking operations
4. Increase timeout threshold if operation is legitimately slow

---

### E203: Connection Failed
**Category:** Runtime  
**Severity:** Error

Failed to establish connection to MCP server.

**Common Causes:**
- Server not running
- Wrong transport configuration
- Network connectivity issues
- Port already in use

**Remediation:**
1. Verify server is running
2. Check transport type (stdio, HTTP, SSE)
3. Verify network connectivity
4. Check firewall rules

---

## Schema & Definition Errors (E1xx)

### E101: Missing Tool Description
**Category:** Schema  
**Severity:** Error

Tool has no description field or description is empty/whitespace.

**Why It Matters:**
LLM cannot understand what the tool does, leading to incorrect tool selection.

**Example:**
```
E101  Missing Tool Description    process_data
```

**Remediation:**
Add a clear description explaining:
- What the tool does
- What inputs it expects
- What it returns
- When to use it

**Good Example:**
```json
{
  "name": "fetch_user",
  "description": "Fetch user details by ID. Returns user object with name, email, and created_at timestamp. Use when you need current user information."
}
```

---

### E102: Missing Parameter Type
**Category:** Schema  
**Severity:** Error

Parameter has no `type` field defined in schema.

**Why It Matters:**
Schema cannot validate input. LLMs will send wrong value types.

**Example:**
```
E102  Missing Parameter Type    get_user.user_id
```

**Remediation:**
Add `type` field to parameter schema.

**Before:**
```json
{
  "name": "user_id",
  "description": "User ID"
}
```

**After:**
```json
{
  "name": "user_id",
  "type": "string",
  "format": "uuid",
  "description": "User ID"
}
```

---

### E103: Required Parameter No Description
**Category:** Schema  
**Severity:** Error

Required parameter has no description.

**Why It Matters:**
LLMs will guess what value to provide without guidance.

**Example:**
```
E103  Required Parameter No Description    get_user.user_id
```

**Remediation:**
Add description explaining what value the parameter expects.

**Before:**
```json
{
  "name": "user_id",
  "type": "string",
  "required": true
}
```

**After:**
```json
{
  "name": "user_id",
  "type": "string",
  "required": true,
  "description": "Unique identifier for the user (UUID format)"
}
```

---

### E112: Sensitive Parameter Exposed
**Category:** Schema  
**Severity:** Error

Parameter name suggests sensitive data (password, secret, token, api_key) without being marked `writeOnly`.

**Why It Matters:**
Agents may log, cache, or include secrets in conversation history.

**Detected Patterns:**
password, passwd, secret, token, api_key, apikey, auth, credential, private_key, access_key, client_secret, webhook_secret

**Example:**
```
E112  Sensitive Parameter Exposed    login.password
```

**Remediation:**
Mark sensitive parameters with `writeOnly: true` or use a dedicated secrets mechanism.

---

### E113: Duplicate Tool Names
**Category:** Schema  
**Severity:** Error

Multiple tools share the exact same name.

**Why It Matters:**
Tool routing becomes ambiguous. Agents cannot distinguish between tools.

**Example:**
```
E113  Duplicate Tool Names    get_user (appears 2 times)
```

**Remediation:**
Rename tools to be unique.

---

### E104-E110: Reserved
**Category:** Schema  
**Severity:** Error

Reserved for future static analysis features:
- E104: Parameter Not In Description
- E105: Free Text Propagation  
- E106: Output Not Guaranteed
- E107: Circular Dependency
- E110: Tool Ambiguity

---

## Output Issues (E3xx)

### E301: Output Explosion
**Category:** Output  
**Severity:** Error

Tool output exceeds recommended size limit, risking context window exhaustion.

**Why It Matters:**
Large outputs blow the context window, making LLM unable to process.

**Example:**
```
W301  Large Output Warning    list_all_users (estimated: 12MB)
```

**Remediation:**
1. Add pagination (limit/offset parameters)
2. Limit array sizes in schema (maxItems)
3. Add filtering parameters
4. Stream results instead of returning all at once

---

### E302: Reserved
**Category:** Output  
**Severity:** Error

Reserved for malformed JSON detection.

---

## Assertion Failures (E4xx)

### E401-E402: Reserved
**Category:** Test  
**Severity:** Error

Reserved for assertion and snapshot testing:
- E401: Assertion Failed
- E402: Snapshot Mismatch

---

## Behavioral Errors (E5xx)

### E500-E501: Reserved
**Category:** Behavior  
**Severity:** Error

Reserved for behavioral analysis:
- E500: Unexpected Side Effect
- E501: Non-Idempotent Operation

---

## Warnings (W1xx)

### W101: Generic/Vague Description
**Category:** Schema  
**Severity:** Warning

Tool description is too vague or generic for effective tool selection.

**Common Generic Phrases:**
- "Get data"
- "Process information"
- "Handle request"

**Remediation:**
Be specific about what the tool does, what inputs it needs, and what it returns.

---

### W102: Optional Parameter No Description
**Category:** Schema  
**Severity:** Warning

Optional parameter has no description.

**Remediation:**
Add description explaining what the parameter does and valid values.

---

### W103: String Without Constraints
**Category:** Schema  
**Severity:** Warning

String parameter has no enum, pattern, example, or default.

**Why It Matters:**
LLM may hallucinate values without guidance.

**Remediation:**
Add format, enum, pattern, or examples to constrain valid values.

---

### W104: Generic Parameter Names
**Category:** Schema  
**Severity:** Warning

Parameter has generic name (data, value, input, payload, etc.) with no description.

**Common Generic Names:**
- "data"
- "value"
- "input"
- "payload"
- "options"

**Example:**
```
W104  Generic Description    update_status: "Update status" — too vague
```

**Remediation:**
Be specific about:
- What data/resource is affected
- What inputs are required
- What the output represents
- When to use this vs similar tools

**Before:**
```
"description": "Get data"
```

**After:**
```
"description": "Fetch sales data for a specific date range. Returns array of sale records with amount, customer, and timestamp. Use when analyzing sales performance."
```

---

### W105: Tool Similarity
**Category:** Schema  
**Severity:** Warning

Two or more tools have very similar descriptions (>80% similarity).

**Why It Matters:**
LLMs may confuse similar tools and pick the wrong one.

**Example:**
```
W105  Tool Similarity    delete_entities ↔ delete_relations (82% similar)
```

**Remediation:**
1. Make tool names and descriptions more distinct
2. Clarify when to use each tool
3. Merge tools if they do the same thing

---

### W106: Schema Bloat
**Category:** Schema  
**Severity:** Warning

Total tools/list response exceeds 8K tokens (~32KB).

**Why It Matters:**
Consumes significant portion of LLM context window.

**Remediation:**
1. Reduce number of tools
2. Simplify schema definitions
3. Remove unused tools

---

### W107: Non-Deterministic Output
**Category:** Behavior  
**Severity:** Warning

Tool produces different outputs for identical inputs across multiple calls.

**Detected via:** `--detect-nondeterminism` flag (calls tool 3x, compares hashes)

**Remediation:**
Ensure tool is deterministic, or document non-deterministic behavior.

---

### W108: Hidden Side Effects
**Category:** Schema  
**Severity:** Warning

Tool name implies mutation (create, delete, update) but description doesn't acknowledge side effects.

**Why It Matters:**
Agents may retry mutation operations unsafely if they don't know the tool has side effects.

**Remediation:**
Add explicit mutation language to description (e.g., "Creates a new record", "Permanently deletes").

---

### W109: Missing Examples
**Category:** Schema  
**Severity:** Warning

User-facing parameter (query, email, url, name, etc.) has no examples.

**Why It Matters:**
LLMs perform significantly better when schemas include representative values.

**Remediation:**
Add `examples` array with 1-2 representative values.

---

### W110: Schema-Description Drift
**Category:** Schema  
**Severity:** Warning

More than 50% of tool parameters are not mentioned in the description.

**Why It Matters:**
Stale descriptions confuse agents about what inputs are needed.

**Remediation:**
Update description to reference all parameters, or remove unused parameters.

---

### W111: Description Quality
**Category:** Schema  
**Severity:** Warning

Description is too short (<20 chars) or too long (>500 chars).

**Remediation:**
Keep between 20-500 characters.

---

### W112: Too Many Tools
**Category:** Schema  
**Severity:** Warning

Server exposes more than 20 tools.

**Why It Matters:**
Research shows LLM tool selection accuracy degrades with scale.

**Remediation:**
Split into multiple focused servers or use tool grouping.

---

### W114: Schema Too Deep
**Category:** Schema  
**Severity:** Warning

Input schema nested deeper than 3 levels.

**Why It Matters:**
LLMs struggle to construct deeply nested JSON correctly.

**Remediation:**
Flatten schema or accept a JSON string parameter.

---

### W115: High Token Cost
**Category:** Schema  
**Severity:** Warning

Single tool definition consumes more than 1000 tokens of context.

**Remediation:**
Simplify schema, reduce parameters, or shorten description.

---

### W116: Broad Output
**Category:** Schema  
**Severity:** Warning

Tool description doesn't mention what it returns.

**Remediation:**
Add a "Returns ..." clause to the description.

---

## Using Error Codes

### In CI/CD

Filter or ignore specific error codes:

```bash
# Ignore warnings in CI
mcp-assert audit --server "..." --ignore W101,W104

# Only show specific categories
mcp-assert audit --server "..." --categories runtime,output

# Fail on any error
mcp-assert audit --server "..." --strict
```

### In JSON Output

Error codes appear in structured output:

```json
{
  "tool": "create_table",
  "status": "crash",
  "error_code": "E201",
  "detail": "panic: nil pointer dereference"
}
```

### Filtering by Category

```bash
# All schema-related issues
mcp-assert lint --server "..." --categories schema

# All runtime issues
mcp-assert audit --server "..." --categories runtime
```

---

## Error Code Ranges

| Range | Category | Codes |
|-------|----------|-------|
| E000 | Success | E000 |
| E1xx | Schema & Definition | E101, E102, E103, E105, E107, E112, E113 |
| E2xx | Runtime | E201, E202, E203 |
| E3xx | Output Issues | E301 |
| E4xx | Assertion Failures | E401, E402 (reserved) |
| E5xx | Behavioral | E500, E501 (reserved) |
| W1xx | Schema & Behavior Warnings | W101-W116 (14 rules) |
| W3xx | Output Warnings | W301 (reserved) |

---

## Auto-Fix (`--fix`)

The `--fix` flag auto-generates schema improvement suggestions:

```bash
mcp-assert lint --server "npx my-server" --fix
```

**Fixable codes:**

| Code | Fix Strategy |
|------|-------------|
| E101 | Generate description from tool name + parameters |
| E103 | Generate param description from name + type |
| W103 | Infer format or examples from param name (email→email, user_id→uuid) |
| W108 | Prepend mutation acknowledgment to description |
| W109 | Generate examples from param name patterns (query→"search term") |
| W111 | Expand short descriptions using tool name |
| W116 | Append "Returns ..." clause based on tool verb (get→"Returns X data") |

**Not auto-fixable (require human judgment):**
- E105, E107: Circular deps and free text propagation (need restructuring)
- E112: Sensitive params (need architectural decision)
- W105: Tool similarity (need renaming or merging)
- W112: Too many tools (need splitting)
- W114: Schema depth (need flattening)

**JSON output:**

```bash
mcp-assert lint --server "..." --fix --json
```

Returns structured fixes for programmatic consumption:

```json
{
  "server": "my-server",
  "tools": 9,
  "findings": 25,
  "fixable": 23,
  "fixes": [
    {
      "tool": "get_user",
      "code": "E103",
      "field": "args.user_id.description",
      "action": "set_description",
      "value": "Unique identifier (string)",
      "message": "Add description for \"user_id\": \"Unique identifier (string)\""
    }
  ]
}
```

---

## Adding New Error Codes

Error codes are defined in `internal/report/codes.go`. To add a new code:

1. Define the constant in the `const` block
2. Add to `ErrorRegistry` with full metadata
3. Implement detection logic in appropriate lint file
4. Add fix generation in `lint_fix.go` (if auto-fixable)
5. Update this documentation

---

**Last Updated:** 2026-05-29  
**Total Rules:** 24 (7 errors, 17 warnings)
**Version:** Phase 1 (Runtime errors only, lint errors coming in Phase 1.2)
