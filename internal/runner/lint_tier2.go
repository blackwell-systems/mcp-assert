// lint_tier2.go implements Tier 2 lint rules:
// W108 (hidden side effects), W109 (missing examples on user-facing params),
// W110 (schema-description drift).
package runner

import (
	"fmt"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// lintHiddenSideEffects flags tools whose names suggest mutation but whose
// descriptions don't acknowledge side effects. Agents may retry these unsafely.
func lintHiddenSideEffects(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding

	for _, tool := range tools {
		name := strings.ToLower(tool.Name)
		desc := strings.ToLower(tool.Description)

		if desc == "" {
			continue // handled by E101
		}

		// Check if tool name implies mutation
		isMutation := false
		for _, verb := range mutationVerbs {
			if strings.Contains(name, verb) {
				isMutation = true
				break
			}
		}

		if !isMutation {
			continue
		}

		// Check if description acknowledges side effects
		sideEffectSignals := []string{
			"creates", "deletes", "removes", "updates", "modifies", "writes",
			"sends", "publishes", "mutates", "changes", "inserts", "drops",
			"side effect", "not idempotent", "destructive", "irreversible",
			"will create", "will delete", "will update", "will modify",
		}

		acknowledged := false
		for _, signal := range sideEffectSignals {
			if strings.Contains(desc, signal) {
				acknowledged = true
				break
			}
		}

		if !acknowledged {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W108,
				Severity: report.SeverityWarning,
				Message:  fmt.Sprintf("Tool name %q suggests mutation but description doesn't acknowledge side effects. Agents may retry unsafely.", tool.Name),
				Field:    "description",
			})
		}
	}

	return findings
}

// userFacingParamPatterns are parameter name patterns that benefit most from examples.
var userFacingParamPatterns = []string{
	"query", "search", "email", "url", "uri", "phone",
	"address", "username", "name", "title", "message",
	"prompt", "question", "filter", "sort", "locale",
	"timezone", "language", "currency", "country", "region",
}

// lintMissingExamples flags user-facing parameters that lack examples.
// LLMs perform significantly better when schemas include representative values.
func lintMissingExamples(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding

	for _, tool := range tools {
		props := tool.InputSchema.Properties
		if len(props) == 0 {
			continue
		}

		for name, prop := range props {
			propMap, ok := prop.(map[string]any)
			if !ok {
				continue
			}

			// Only check user-facing params
			lower := strings.ToLower(name)
			isUserFacing := false
			for _, pattern := range userFacingParamPatterns {
				if strings.Contains(lower, pattern) {
					isUserFacing = true
					break
				}
			}

			if !isUserFacing {
				continue
			}

			// Check if it already has examples, enum, or default
			_, hasExamples := propMap["examples"]
			_, hasEnum := propMap["enum"]
			_, hasDefault := propMap["default"]

			if !hasExamples && !hasEnum && !hasDefault {
				findings = append(findings, LintFinding{
					Tool:     tool.Name,
					Code:     report.W109,
					Severity: report.SeverityWarning,
					Message:  fmt.Sprintf("User-facing parameter %q has no examples. LLMs perform better with representative values.", name),
					Field:    "args." + name + ".examples",
				})
			}
		}
	}

	return findings
}

// lintSchemaDescriptionDrift flags tools where more than half of the parameters
// are not mentioned anywhere in the description. This indicates stale docs.
func lintSchemaDescriptionDrift(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding

	for _, tool := range tools {
		desc := strings.ToLower(tool.Description)
		if desc == "" {
			continue // handled by E101
		}

		props := tool.InputSchema.Properties
		if len(props) < 2 {
			continue // not meaningful for single-param tools
		}

		mentioned := 0
		total := 0

		for name := range props {
			total++
			// Check if param name (or its words) appears in description
			lower := strings.ToLower(name)
			// Check exact match
			if strings.Contains(desc, lower) {
				mentioned++
				continue
			}
			// Check individual words (for snake_case params like "user_id")
			words := strings.Split(lower, "_")
			wordMentioned := false
			for _, w := range words {
				if len(w) > 2 && strings.Contains(desc, w) {
					wordMentioned = true
					break
				}
			}
			if wordMentioned {
				mentioned++
			}
		}

		if total > 0 && float64(mentioned)/float64(total) < 0.5 {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     report.W110,
				Severity: report.SeverityWarning,
				Message:  fmt.Sprintf("Only %d/%d parameters mentioned in description. Schema and description may be out of sync.", mentioned, total),
				Field:    "description",
			})
		}
	}

	return findings
}
