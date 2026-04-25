# Excel MCP Server Examples

These assertions test [haris-musa/excel-mcp-server](https://github.com/haris-musa/excel-mcp-server) (3,750 stars), a Python MCP server for Excel file operations.

## Setup

```bash
uv tool install excel-mcp-server
# or: pip install excel-mcp-server
```

Then run (requires absolute fixture path):

```bash
mcp-assert run --suite examples/excel-mcp --fixture $(pwd)/examples/excel-mcp/fixtures
```

## Coverage

15 assertions covering 13/25 tools (52%):

| Category | Tools | Assertions |
|----------|-------|------------|
| Workbook | `create_workbook`, `get_workbook_metadata` | 2 |
| Sheet | `create_worksheet`, `rename_worksheet` | 2 |
| Data | `write_data_to_excel`, `read_data_from_excel` | 2 (including round-trip) |
| Formula | `apply_formula`, `validate_formula_syntax` | 3 (valid + invalid) |
| Formatting | `format_range`, `merge_cells` | 2 |
| Validation | `validate_excel_range` | 1 |
| Chart | `create_chart` | 1 |
| Pivot | `create_pivot_table` | 1 |
| Row ops | `insert_rows` | 1 |
| Error | `read_data_from_excel` (missing file) | 1 |

## Notes

Clean scan: no bugs found. All 15 assertions pass.

The server requires absolute file paths in stdio mode. The `EXCEL_FILES_PATH` env var is set to `{{fixture}}` in each assertion. Pass an absolute path to `--fixture` when running.

Fixture isolation keeps the fixture directory clean: each assertion creates its own workbooks via setup steps, and the temp copies are cleaned up after each assertion.
