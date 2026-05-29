// lint_fix.go implements the "mcp-assert lint --fix" auto-repair feature.
//
// For each lint finding, generates a suggested fix: improved descriptions,
// added examples, format annotations, etc. Outputs as markdown report or JSON.
package runner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// SchemaFix represents a single auto-generated fix for a lint finding.
type SchemaFix struct {
	Tool    string         `json:"tool"`
	Code    report.ErrorCode `json:"code"`
	Field   string         `json:"field"`
	Action  string         `json:"action"`  // "add_description", "add_examples", "add_format", etc.
	Value   any            `json:"value"`   // the suggested value
	Message string         `json:"message"` // human-readable explanation
}

// FixReport holds all generated fixes for a server.
type FixReport struct {
	Server    string      `json:"server"`
	Tools     int         `json:"tools"`
	Findings  int         `json:"findings"`
	Fixable   int         `json:"fixable"`
	Fixes     []SchemaFix `json:"fixes"`
}

// generateFixes takes lint findings and generates auto-repair suggestions.
func generateFixes(tools []mcp.Tool, findings []LintFinding) []SchemaFix {
	var fixes []SchemaFix

	// Build tool lookup
	toolMap := make(map[string]mcp.Tool)
	for _, t := range tools {
		toolMap[t.Name] = t
	}

	for _, f := range findings {
		fix := generateFixForFinding(f, toolMap)
		if fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	return fixes
}

// generateFixForFinding produces a fix for a single lint finding.
func generateFixForFinding(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	switch f.Code {
	case report.E101:
		return fixMissingDescription(f, toolMap)
	case report.E103:
		return fixMissingParamDescription(f, toolMap)
	case report.W103:
		return fixMissingConstraints(f, toolMap)
	case report.W108:
		return fixHiddenSideEffects(f, toolMap)
	case report.W109:
		return fixMissingExamples(f, toolMap)
	case report.W111:
		return fixDescriptionLength(f, toolMap)
	case report.W116:
		return fixBroadOutput(f, toolMap)
	default:
		return nil
	}
}

// fixMissingDescription generates a description from the tool name.
func fixMissingDescription(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	tool := toolMap[f.Tool]
	desc := generateDescriptionFromName(tool.Name, tool.InputSchema)

	return &SchemaFix{
		Tool:    f.Tool,
		Code:    f.Code,
		Field:   "description",
		Action:  "set_description",
		Value:   desc,
		Message: fmt.Sprintf("Add description: %q", desc),
	}
}

// fixMissingParamDescription generates a param description from name + type.
func fixMissingParamDescription(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	// Extract param name from field (e.g., "args.entities.description" -> "entities")
	parts := strings.Split(f.Field, ".")
	if len(parts) < 2 {
		return nil
	}
	paramName := parts[1]

	tool := toolMap[f.Tool]
	propMap := getPropertyMap(tool, paramName)
	typ, _ := propMap["type"].(string)

	desc := generateParamDescription(paramName, typ, f.Tool)

	return &SchemaFix{
		Tool:    f.Tool,
		Code:    f.Code,
		Field:   f.Field,
		Action:  "set_description",
		Value:   desc,
		Message: fmt.Sprintf("Add description for %q: %q", paramName, desc),
	}
}

// fixMissingConstraints suggests enum/format/examples for unconstrained strings.
func fixMissingConstraints(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	parts := strings.Split(f.Field, ".")
	if len(parts) < 2 {
		return nil
	}
	paramName := parts[1]

	// Try to infer format from name
	format := inferFormat(paramName)
	if format != "" {
		return &SchemaFix{
			Tool:    f.Tool,
			Code:    f.Code,
			Field:   f.Field + ".format",
			Action:  "set_format",
			Value:   format,
			Message: fmt.Sprintf("Add format %q to %q", format, paramName),
		}
	}

	// Try to generate examples from name
	examples := inferExamples(paramName)
	if len(examples) > 0 {
		return &SchemaFix{
			Tool:    f.Tool,
			Code:    f.Code,
			Field:   f.Field + ".examples",
			Action:  "set_examples",
			Value:   examples,
			Message: fmt.Sprintf("Add examples to %q: %v", paramName, examples),
		}
	}

	return nil
}

// fixHiddenSideEffects prepends mutation acknowledgment to description.
func fixHiddenSideEffects(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	tool := toolMap[f.Tool]
	verb := extractMutationVerb(tool.Name)
	newDesc := fmt.Sprintf("%ss data. %s", capitalize(verb), tool.Description)

	return &SchemaFix{
		Tool:    f.Tool,
		Code:    f.Code,
		Field:   "description",
		Action:  "set_description",
		Value:   newDesc,
		Message: fmt.Sprintf("Acknowledge side effect: %q", newDesc),
	}
}

// fixMissingExamples generates examples from param name patterns.
func fixMissingExamples(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	parts := strings.Split(f.Field, ".")
	if len(parts) < 2 {
		return nil
	}
	paramName := parts[1]

	examples := inferExamples(paramName)
	if len(examples) == 0 {
		// Generic fallback
		examples = []string{"example_value"}
	}

	return &SchemaFix{
		Tool:    f.Tool,
		Code:    f.Code,
		Field:   strings.TrimSuffix(f.Field, ".examples"),
		Action:  "set_examples",
		Value:   examples,
		Message: fmt.Sprintf("Add examples to %q: %v", paramName, examples),
	}
}

// fixDescriptionLength suggests expansion for short descriptions.
func fixDescriptionLength(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	tool := toolMap[f.Tool]
	if len(tool.Description) >= 20 {
		return nil // too long case - can't auto-fix (would need to summarize)
	}

	expanded := generateDescriptionFromName(tool.Name, tool.InputSchema)
	if expanded == tool.Description {
		return nil
	}

	return &SchemaFix{
		Tool:    f.Tool,
		Code:    f.Code,
		Field:   "description",
		Action:  "set_description",
		Value:   expanded,
		Message: fmt.Sprintf("Expand description: %q", expanded),
	}
}

// fixBroadOutput appends a "Returns ..." clause to the description.
func fixBroadOutput(f LintFinding, toolMap map[string]mcp.Tool) *SchemaFix {
	tool := toolMap[f.Tool]
	returnClause := inferReturnClause(tool.Name)
	newDesc := tool.Description + " " + returnClause

	return &SchemaFix{
		Tool:    f.Tool,
		Code:    f.Code,
		Field:   "description",
		Action:  "append_description",
		Value:   returnClause,
		Message: fmt.Sprintf("Append return info: %q", newDesc),
	}
}

// capitalize uppercases the first letter of a string.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// --- Heuristic generators ---

// generateDescriptionFromName creates a description from tool name + params.
func generateDescriptionFromName(name string, schema mcp.ToolInputSchema) string {
	words := splitToolName(name)
	if len(words) == 0 {
		return "Perform an operation"
	}

	// First word is typically the verb
	verb := words[0]
	object := strings.Join(words[1:], " ")

	if object == "" {
		object = "data"
	}

	// Count params for context
	paramCount := len(schema.Properties)
	paramHint := ""
	if paramCount > 0 {
		paramNames := make([]string, 0, paramCount)
		for name := range schema.Properties {
			paramNames = append(paramNames, name)
		}
		if len(paramNames) <= 3 {
			paramHint = fmt.Sprintf(" Accepts: %s.", strings.Join(paramNames, ", "))
		}
	}

	desc := fmt.Sprintf("%s %s.%s", capitalize(verb), object, paramHint)
	return desc
}

// generateParamDescription creates a description from param name + type.
func generateParamDescription(name, typ, toolName string) string {
	readable := strings.ReplaceAll(name, "_", " ")

	// Common patterns
	patterns := map[string]string{
		"id":       "Unique identifier",
		"ids":      "List of identifiers",
		"name":     "Display name",
		"names":    "List of names",
		"path":     "File or directory path",
		"url":      "URL endpoint",
		"query":    "Search query string",
		"email":    "Email address",
		"limit":    "Maximum number of results to return",
		"offset":   "Number of results to skip (for pagination)",
		"page":     "Page number (for pagination)",
		"filter":   "Filter criteria",
		"sort":     "Sort order",
		"format":   "Output format",
		"timeout":  "Timeout duration",
		"content":  "Content body",
		"message":  "Message text",
		"title":    "Title or heading",
		"status":   "Current status",
		"type":     "Type classification",
		"key":      "Lookup key",
		"value":    "Value to set",
		"token":    "Authentication or pagination token",
		"password": "User password (sensitive)",
	}

	// Check for suffix matches (e.g., "user_id" matches "id")
	lower := strings.ToLower(name)
	for pattern, desc := range patterns {
		if lower == pattern || strings.HasSuffix(lower, "_"+pattern) {
			if typ != "" {
				return fmt.Sprintf("%s (%s)", desc, typ)
			}
			return desc
		}
	}

	// Fallback: humanize the name
	if typ != "" {
		return fmt.Sprintf("The %s value (%s)", readable, typ)
	}
	return fmt.Sprintf("The %s value", readable)
}

// inferFormat guesses a JSON Schema format from parameter name.
func inferFormat(name string) string {
	lower := strings.ToLower(name)
	formats := map[string]string{
		"email":      "email",
		"url":        "uri",
		"uri":        "uri",
		"href":       "uri",
		"link":       "uri",
		"date":       "date",
		"datetime":   "date-time",
		"created_at": "date-time",
		"updated_at": "date-time",
		"timestamp":  "date-time",
		"time":       "time",
		"uuid":       "uuid",
		"ip":         "ipv4",
		"ipv4":       "ipv4",
		"ipv6":       "ipv6",
		"hostname":   "hostname",
	}

	for pattern, format := range formats {
		if lower == pattern || strings.HasSuffix(lower, "_"+pattern) || strings.HasPrefix(lower, pattern+"_") {
			return format
		}
	}

	// Check for ID patterns
	if strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "id") && lower != "id" {
		return "uuid"
	}

	return ""
}

// inferExamples generates example values from parameter name patterns.
func inferExamples(name string) []string {
	lower := strings.ToLower(name)

	examples := map[string][]string{
		"email":      {"user@example.com"},
		"url":        {"https://api.example.com/resource"},
		"uri":        {"https://api.example.com/resource"},
		"query":      {"search term"},
		"name":       {"example-name"},
		"username":   {"johndoe"},
		"title":      {"Example Title"},
		"message":    {"Hello, world"},
		"path":       {"/path/to/file.txt"},
		"file":       {"document.pdf"},
		"directory":  {"/tmp/output"},
		"language":   {"en"},
		"locale":     {"en-US"},
		"timezone":   {"America/New_York"},
		"currency":   {"USD"},
		"country":    {"US"},
		"phone":      {"+1-555-0100"},
		"address":    {"123 Main St, City, ST 12345"},
		"color":      {"#ff5733"},
		"tag":        {"important"},
		"tags":       {"bug,enhancement"},
		"status":     {"active"},
		"priority":   {"high"},
		"format":     {"json"},
		"sort":       {"created_at:desc"},
		"filter":     {"status=active"},
		"limit":      {"10"},
		"offset":     {"0"},
		"page":       {"1"},
	}

	for pattern, exs := range examples {
		if lower == pattern || strings.HasSuffix(lower, "_"+pattern) || strings.Contains(lower, pattern) {
			return exs
		}
	}

	return nil
}

// inferReturnClause generates a "Returns ..." clause from the tool name.
func inferReturnClause(name string) string {
	words := splitToolName(name)
	if len(words) == 0 {
		return "Returns the operation result."
	}

	verb := strings.ToLower(words[0])
	object := strings.Join(words[1:], " ")

	switch verb {
	case "get", "fetch", "read", "load", "find":
		return fmt.Sprintf("Returns the %s data as JSON.", object)
	case "list", "search", "query":
		return fmt.Sprintf("Returns an array of matching %s.", object)
	case "create", "add", "insert":
		return fmt.Sprintf("Returns the created %s with its ID.", object)
	case "update", "modify", "patch":
		return fmt.Sprintf("Returns the updated %s.", object)
	case "delete", "remove", "destroy":
		return fmt.Sprintf("Returns confirmation of deletion.")
	case "count":
		return fmt.Sprintf("Returns the count as a number.")
	case "check", "validate", "verify":
		return fmt.Sprintf("Returns validation result (pass/fail with details).")
	default:
		return fmt.Sprintf("Returns the %s result.", object)
	}
}

// extractMutationVerb pulls the mutation verb from a tool name.
func extractMutationVerb(name string) string {
	lower := strings.ToLower(name)
	for _, verb := range mutationVerbs {
		if strings.Contains(lower, verb) {
			return verb
		}
	}
	return "modify"
}

// splitToolName splits a tool name into words (handles snake_case).
func splitToolName(name string) []string {
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	words := strings.Fields(name)
	return words
}

// getPropertyMap extracts the property map for a named parameter.
func getPropertyMap(tool mcp.Tool, paramName string) map[string]any {
	prop, ok := tool.InputSchema.Properties[paramName]
	if !ok {
		return nil
	}
	propMap, ok := prop.(map[string]any)
	if !ok {
		return nil
	}
	return propMap
}

// printFixReport outputs the fix report in human-readable format.
func printFixReport(r FixReport) {
	if len(r.Fixes) == 0 {
		fmt.Printf("✓ %s: %d tools, %d findings, none auto-fixable\n", r.Server, r.Tools, r.Findings)
		return
	}

	fmt.Printf("%s: %d tools, %d findings, %d auto-fixable\n\n", r.Server, r.Tools, r.Findings, r.Fixable)

	for _, fix := range r.Fixes {
		fmt.Printf("  %-5s  %-30s  %s\n", fix.Code, fix.Tool, fix.Message)
	}

	fmt.Printf("\n%d fixes generated. Apply to your tool schemas to resolve lint findings.\n", len(r.Fixes))
}

// printFixReportJSON outputs the fix report as JSON.
func printFixReportJSON(r FixReport) {
	data, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(data))
}
