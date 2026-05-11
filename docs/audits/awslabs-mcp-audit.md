# AWS Labs MCP Servers: Schema Quality Audit

**Repository:** https://github.com/awslabs/mcp
**Date:** 2026-05-11
**Tool:** mcp-assert lint v0.11.0
**Servers scanned:** 43 of 57
**Transport:** stdio

## Summary

**4,484 findings (2,160 errors, 2,324 warnings) across 870 tools in 43 servers. Only 5 servers have zero errors.**

### Top 10 by errors

| Server | Tools | Errors | Warnings | Total |
|--------|-------|--------|----------|-------|
| aws-healthomics-mcp-server | 73 | 313 | 387 | 700 |
| valkey-mcp-server | 105 | 260 | 444 | 704 |
| aws-dataprocessing-mcp-server | 36 | 259 | 79 | 338 |
| amazon-bedrock-agentcore-mcp-server | 122 | 177 | 249 | 426 |
| elasticache-mcp-server | 39 | 147 | 171 | 318 |
| billing-cost-management-mcp | 33 | 138 | 54 | 192 |
| aws-serverless-mcp-server | 25 | 93 | 49 | 142 |
| cloudwatch-mcp-server | 26 | 90 | 29 | 119 |
| aws-iot-sitewise-mcp-server | 41 | 88 | 92 | 180 |
| aws-core-network-mcp-server | 27 | 69 | 112 | 181 |

### All servers

| Server | Tools | Errors | Warnings |
|--------|-------|--------|----------|
| aws-healthomics-mcp-server | 73 | 313 | 387 |
| valkey-mcp-server | 105 | 260 | 444 |
| aws-dataprocessing-mcp-server | 36 | 259 | 79 |
| amazon-bedrock-agentcore-mcp-server | 122 | 177 | 249 |
| elasticache-mcp-server | 39 | 147 | 171 |
| billing-cost-management-mcp | 33 | 138 | 54 |
| aws-serverless-mcp-server | 25 | 93 | 49 |
| cloudwatch-mcp-server | 26 | 90 | 29 |
| aws-iot-sitewise-mcp-server | 41 | 88 | 92 |
| aws-core-network-mcp-server | 27 | 69 | 112 |
| timestream-for-influxdb-mcp-server | 26 | 65 | 38 |
| aws-for-sap-management-mcp-server | 22 | 64 | 50 |
| aws-appsync-mcp-server | 10 | 59 | 22 |
| aws-transform-mcp-server | 19 | 55 | 32 |
| cloudwatch-applicationsignals-mcp-server | 22 | 50 | 48 |
| memcached-mcp-server | 22 | 37 | 55 |
| healthimaging-mcp-server | 39 | 31 | 86 |
| sagemaker-ai-mcp-server | 4 | 20 | 4 |
| iam-mcp-server | 29 | 19 | 83 |
| s3-tables-mcp-server | 16 | 16 | 26 |
| ccapi-mcp-server | 14 | 14 | 20 |
| aws-pricing-mcp-server | 9 | 14 | 10 |
| dynamodb-mcp-server | 8 | 12 | 8 |
| finch-mcp-server | 3 | 11 | 5 |
| aws-support-mcp-server | 6 | 11 | 8 |
| cloudtrail-mcp-server | 5 | 8 | 3 |
| well-architected-security-mcp-server | 6 | 7 | 2 |
| postgres-mcp-server | 7 | 6 | 16 |
| documentdb-mcp-server | 16 | 5 | 54 |
| aurora-dsql-mcp-server | 7 | 5 | 6 |
| amazon-kendra-index-mcp-server | 2 | 4 | 4 |
| neptune-mcp-server | 4 | 3 | 4 |
| aws-documentation-mcp-server | 4 | 2 | 4 |
| aws-bedrock-custom-model-import-mcp-server | 6 | 2 | 6 |
| aws-iac-mcp-server | 9 | 2 | 11 |
| aws-api-mcp-server | 2 | 2 | 1 |
| bedrock-kb-retrieval-mcp-server | 2 | 1 | 2 |
| aws-location-mcp-server | 8 | 1 | 7 |
| keyspaces-mcp-server | 6 | 0 | 9 |
| document-loader-mcp-server | 3 | 0 | 5 |
| stepfunctions-tool-mcp-server | 0 | 0 | 0 |
| redshift-mcp-server | 6 | 0 | 28 |
| amazon-qbusiness-anonymous-mcp-server | 1 | 0 | 1 |
| **Total** | **870** | **2,160** | **2,324** |

### Clean servers (0 errors)

- stepfunctions-tool-mcp-server (0 tools exposed via lint)
- redshift-mcp-server (0 errors, 28 warnings)
- keyspaces-mcp-server (0 errors, 9 warnings)
- document-loader-mcp-server (0 errors, 5 warnings)
- amazon-qbusiness-anonymous-mcp-server (0 errors, 1 warning)

## Systemic Issues

### 1. E102: Parameters with no type defined (116 occurrences)

The dominant error across all servers. Parameters appear in JSON Schema without a `type` field, meaning agents cannot determine whether to send a string, integer, boolean, or object.

**Affected parameters by server:**

| Server | Parameter | Count | Impact |
|--------|-----------|-------|--------|
| cloudwatch | `region`, `max_items`, `period`, `limit` | 88 | Agents send "300" (string) instead of 300 (int) for period |
| s3-tables | `region_name` | 16 | Every tool affected |
| iam | `ctx` | 16 | Internal context leaked to agents as a required parameter |
| dynamodb | `usage_data_path`, `region` | 12 | Path format ambiguity |
| aws-documentation | `product_types` | 2 | Array vs string ambiguity |

**Root cause:** Python union types (`int | None`, `str | None`) serialize to JSON Schema without explicit `type` when not annotated with Pydantic Field metadata. The `mcp` Python SDK's schema generation strips `None` from unions but doesn't always emit the remaining type.

**Fix:** Add explicit `Field(json_schema_extra={"type": "integer"})` or use `Annotated[int, Field(...)]` patterns.

### 2. W105: Near-duplicate tool descriptions (57 occurrences)

IAM server has 44 tool pairs with 80-96% similar descriptions. Examples:
- `list_users` and `list_groups`: 96% similar
- `get_user` and `get_group`: 94% similar
- `list_user_policies` and `list_group_policies`: 92% similar

**Impact:** Agents confuse tools with nearly identical descriptions, calling the wrong one. A description that says "List IAM users with optional filtering" vs "List IAM groups with optional filtering" gives the agent no semantic anchor to distinguish them.

**Fix:** Include the key differentiator in the first sentence. "List IAM **users** (people with console/API access)" vs "List IAM **groups** (collections of users sharing permissions)."

### 3. W103: Required parameters without examples (72 occurrences)

Parameters like `log_group_arn`, `namespace`, `metric_name`, `user_name` have no example values. Agents must guess the format.

**Impact:** Agents hallucinate ARN formats, namespace conventions, and naming patterns. An example like `"arn:aws:logs:us-east-1:123456789012:log-group:my-app"` eliminates format ambiguity.

### 4. W106: Oversized tools/list response

| Server | Token cost |
|--------|-----------|
| cloudwatch | 27,542 tokens (107KB) |
| iam | 15,230 tokens (59KB) |

CloudWatch's tool schema alone consumes 27K tokens of the agent's context window on every `tools/list` call. For Claude with 200K context, that's 14% consumed by schema alone. For Gemini with 32K, it's 86%.

### 5. E103: Required parameters without description + IAM `ctx` leak

IAM server exposes a `ctx` parameter on 16 tools with no description and no type. This appears to be an internal context object leaked into the public schema. Agents receive a required parameter called `ctx` with no guidance on what to provide.

## Per-Server Details

### cloudwatch-mcp-server (26 tools)

| Code | Count | Severity | Representative finding |
|------|-------|----------|----------------------|
| E102 | 88 | error | `region` has no type on every tool; `max_items`, `period`, `limit` also untyped |
| W103 | 23 | warning | `log_group_arn`, `end_time`, `start_time` lack examples |
| W105 | 4 | warning | `analyze_metric` and `get_recommended_metric_alarms` 83% similar |
| E103 | 2 | error | `namespace` required with no description |
| W102 | 1 | warning | `dimensions` optional with no description |
| W106 | 1 | warning | 27,542 tokens (107KB) tools/list |

### iam-mcp-server (29 tools)

| Code | Count | Severity | Representative finding |
|------|-------|----------|----------------------|
| W105 | 44 | warning | 44 tool pairs with 80-96% similar descriptions |
| W103 | 38 | warning | `user_name`, `group_name`, `policy_arn` lack examples |
| E102 | 16 | error | `ctx` parameter has no type on 16 tools |
| E103 | 3 | error | `ctx` required with no description |
| W106 | 1 | warning | 15,230 tokens (59KB) tools/list |

### s3-tables-mcp-server (16 tools)

| Code | Count | Severity | Representative finding |
|------|-------|----------|----------------------|
| W103 | 17 | warning | `metadata_location`, `table_bucket_arn` lack examples |
| E102 | 16 | error | `region_name` has no type on every tool |
| W105 | 9 | warning | `list_table_buckets` and `list_namespaces` 84% similar |

### dynamodb-mcp-server (8 tools)

| Code | Count | Severity | Representative finding |
|------|-------|----------|----------------------|
| E102 | 12 | error | `usage_data_path`, `region` have no type |
| W103 | 8 | warning | `schema_path`, `table_name` lack examples |

### aws-documentation-mcp-server (4 tools)

| Code | Count | Severity | Representative finding |
|------|-------|----------|----------------------|
| W103 | 4 | warning | `url`, `search_phrase` lack examples |
| E102 | 2 | error | `product_types` has no type |

## Reproduction

```bash
pip install mcp-assert

# Install a server
cd src/cloudwatch-mcp-server && uv venv && uv pip install -e .

# Lint
mcp-assert lint -server ".venv/bin/awslabs.cloudwatch-mcp-server"
```

## Recommendations

1. **Add a shared `region` field definition** with type and example across all servers. A single base class or shared Field would fix 88+ errors.
2. **Remove `ctx` from IAM tool schemas** or make it internal. It should not appear as a user-facing parameter.
3. **Differentiate tool descriptions** in IAM. The 96% similarity between `list_users` and `list_groups` is a tool-selection failure waiting to happen.
4. **Consider splitting large servers.** CloudWatch at 107KB schema is unusable on smaller-context models. Split into monitoring, logging, and alerting sub-servers.
5. **Run `mcp-assert lint` in CI** to catch these regressions automatically.

## Servers not scanned (14 of 57)

14 servers failed to install or start via stdio (missing dependencies, Rust compilation required, or non-standard entry points). These were skipped:
- mysql-mcp-server, amazon-mq-mcp-server, amazon-sns-sqs-mcp-server
- mcp-lambda-handler (library, not a server)
- Various servers with native dependencies or custom launch requirements
