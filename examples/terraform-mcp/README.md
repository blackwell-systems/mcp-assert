# Terraform MCP Server Examples

These assertions test [hashicorp/terraform-mcp-server](https://github.com/hashicorp/terraform-mcp-server) (1.3K stars), HashiCorp's official MCP server for the Terraform ecosystem. Queries the Terraform Registry for providers, modules, and policies.

## Setup

```bash
go install github.com/hashicorp/terraform-mcp-server/cmd/terraform-mcp-server@latest
# or: clone and build
```

Then run:

```bash
mcp-assert run --suite examples/terraform-mcp --server "terraform-mcp-server stdio"
```

## Coverage

5 assertions covering 5/9 tools:

| Tool | Notes |
|------|-------|
| `get_latest_provider_version` | AWS provider version lookup |
| `get_provider_capabilities` | AWS provider capabilities |
| `search_modules` | VPC module search |
| `search_policies` | Security policy search |
| `search_providers` | Error handling (nonexistent provider) |

Not covered: `get_module_details`, `get_policy_details`, `get_provider_details` (require valid IDs from prior searches), `search_providers` positive case (requires undocumented `service_slug` values).

## Notes

Clean scan: no bugs found. The server returns clear, actionable error messages (e.g., "use search_providers first to find valid IDs", "verify the namespace and provider name are correct"). No stack traces.
