// lint_quality.go implements Tier 1 quality lint rules:
// W111 (description length), W112 (tool count), W114 (schema depth),
// W115 (token cost), W116 (broad output), and overloaded tool detection.
package runner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// lintToolCount warns when a server exposes more than 20 tools.
// Research shows LLM tool selection accuracy degrades at scale.
func lintToolCount(tools []mcp.Tool) []LintFinding {
	if len(tools) > 20 {
		return []LintFinding{{
			Tool:     "(server)",
			Code:     report.W112,
			Severity: report.SeverityWarning,
			Message:  fmt.Sprintf("Server exposes %d tools. LLM tool selection accuracy degrades beyond 20 tools.", len(tools)),
			Field:    "tool_count",
		}}
	}
	return nil
}

// lintDescriptionLength flags descriptions that are too short (<20 chars)
// or too long (>500 chars) for effective tool selection.
func lintDescriptionLength(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding
	for _, tool := range tools {
		desc := strings.TrimSpace(tool.Description)
		if desc == "" {
			continue // handled by E101
		}
		if len(desc) < 20 {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W111,
				Severity: report.SeverityWarning,
				Message:  fmt.Sprintf("Description is only %d chars. Too short for LLMs to understand when to use this tool.", len(desc)),
				Field:    "description",
			})
		} else if len(desc) > 500 {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W111,
				Severity: report.SeverityWarning,
				Message:  fmt.Sprintf("Description is %d chars. Descriptions over 500 chars waste context budget.", len(desc)),
				Field:    "description",
			})
		}
	}
	return findings
}

// lintSchemaDepth flags input schemas nested deeper than 3 levels.
// LLMs struggle to construct deeply nested JSON correctly.
func lintSchemaDepth(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding
	for _, tool := range tools {
		depth := measureSchemaDepth(tool.InputSchema.Properties, 1)
		if depth > 3 {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W114,
				Severity: report.SeverityWarning,
				Message:  fmt.Sprintf("Input schema is %d levels deep. LLMs struggle with nesting beyond 3 levels.", depth),
				Field:    "schema",
			})
		}
	}
	return findings
}

// measureSchemaDepth recursively measures the deepest nesting level.
func measureSchemaDepth(props map[string]interface{}, current int) int {
	if len(props) == 0 {
		return current
	}
	maxDepth := current
	for _, prop := range props {
		propMap, ok := prop.(map[string]any)
		if !ok {
			continue
		}
		// Check nested object properties
		if nested, ok := propMap["properties"].(map[string]interface{}); ok {
			d := measureSchemaDepth(nested, current+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
		// Check array item properties
		if items, ok := propMap["items"].(map[string]any); ok {
			if nested, ok := items["properties"].(map[string]interface{}); ok {
				d := measureSchemaDepth(nested, current+1)
				if d > maxDepth {
					maxDepth = d
				}
			}
		}
	}
	return maxDepth
}

// lintTokenCost estimates per-tool token consumption and flags expensive tools.
// Tools consuming >1000 tokens eat significant context budget.
func lintTokenCost(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding
	for _, tool := range tools {
		data, err := json.Marshal(tool)
		if err != nil {
			continue
		}
		// ~4 bytes per token for JSON/code
		tokens := len(data) / 4
		if tokens > 1000 {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W115,
				Severity: report.SeverityWarning,
				Message:  fmt.Sprintf("Tool definition consumes ~%d tokens. Consider simplifying schema or description.", tokens),
				Field:    "token_cost",
			})
		}
	}
	return findings
}

// lintBroadOutput flags tools that have no meaningful output constraints.
// MCP tools don't typically have output schemas, but if the description
// doesn't mention what the tool returns, consumers can't reason about it.
func lintBroadOutput(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding
	for _, tool := range tools {
		desc := strings.ToLower(tool.Description)
		// Check if description mentions return value at all
		returnsKeywords := []string{"returns", "return", "output", "produces", "responds with", "result"}
		mentionsReturn := false
		for _, kw := range returnsKeywords {
			if strings.Contains(desc, kw) {
				mentionsReturn = true
				break
			}
		}
		// Only flag tools with descriptions (E101 handles missing desc)
		if tool.Description != "" && !mentionsReturn {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W116,
				Severity: report.SeverityWarning,
				Message:  "Description does not mention what the tool returns. Consumers cannot reason about output.",
				Field:    "description",
			})
		}
	}
	return findings
}

// mutationVerbs are words that indicate a tool performs state changes.
var mutationVerbs = []string{
	"create", "delete", "remove", "update", "modify", "write",
	"insert", "drop", "destroy", "set", "put", "post", "send",
	"publish", "push", "add", "patch", "replace",
}

// lintOverloadedTool flags tools whose descriptions contain more than 3
// distinct action verbs, suggesting the tool does too many things.
func lintOverloadedTool(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding

	actionVerbs := []string{
		"get", "fetch", "read", "list", "search", "find", "query",
		"create", "add", "insert", "write", "post", "send", "publish",
		"update", "modify", "patch", "set", "change", "replace",
		"delete", "remove", "destroy", "drop", "clear",
		"validate", "check", "verify", "test",
		"convert", "transform", "parse", "format",
	}

	for _, tool := range tools {
		desc := strings.ToLower(tool.Description)
		if desc == "" {
			continue
		}

		verbCount := 0
		for _, verb := range actionVerbs {
			// Match whole word
			if strings.Contains(" "+desc+" ", " "+verb+" ") ||
				strings.Contains(" "+desc+" ", " "+verb+"s ") ||
				strings.Contains(" "+desc+" ", " "+verb+"ing ") {
				verbCount++
			}
		}

		if verbCount > 3 {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W103,
				Severity: report.SeverityWarning,
				Message:  fmt.Sprintf("Description contains %d action verbs. Tool may have too many responsibilities; agents struggle with multi-purpose tools.", verbCount),
				Field:    "description",
			})
		}
	}

	return findings
}
