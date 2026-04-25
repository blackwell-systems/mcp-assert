# AntV Chart MCP Server Examples

These assertions test [antvis/mcp-server-chart](https://github.com/antvis/mcp-server-chart) (4K stars), a visualization MCP server from Ant Group's AntV team. Generates 25+ chart types as PNG images.

## Setup

```bash
npx -y @antv/mcp-server-chart
```

Then run:

```bash
mcp-assert run --suite examples/antvis-chart --server "npx -y @antv/mcp-server-chart"
```

## Coverage

16 assertions covering 16/27 tools (59%):

| Category | Tools | Assertions |
|----------|-------|------------|
| Standard charts | `line`, `bar`, `column`, `area`, `pie`, `scatter`, `histogram` | 7 |
| Statistical | `boxplot`, `violin`, `waterfall` | 3 |
| Specialized | `dual_axes`, `sankey`, `treemap`, `liquid`, `word_cloud` | 5 |
| Data | `spreadsheet` | 1 |

## Known bugs ([antvis/mcp-server-chart#291](https://github.com/antvis/mcp-server-chart/issues/291))

9 tools crash with unhandled JavaScript exceptions on default/minimal input instead of returning graceful `isError` responses. Assertions for these tools are excluded from the suite until the upstream fix lands:

`generate_fishbone_diagram`, `generate_mind_map`, `generate_organization_chart`, `generate_flow_diagram`, `generate_network_graph`, `generate_funnel_chart`, `generate_venn_chart`, `generate_district_map`, `generate_radar_chart`

2 additional tools (`generate_path_map`, `generate_pin_map`) fail due to external POI API dependency (not bugs).
