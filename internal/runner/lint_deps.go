// lint_deps.go implements dependency inference and dependency-based lint rules.
//
// MCP tools don't declare explicit output schemas, so we infer dependencies
// heuristically: if tool A has an input parameter whose name/type closely
// matches tool B's input, and B's description mentions producing that data,
// we infer a data-flow edge A -> B.
//
// This powers:
//   - E105: Free text propagation (unconstrained string flows between tools)
//   - E107: Circular dependency (tool dependency graph has a cycle)
package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// ToolDependency represents an inferred data-flow edge between two tools.
type ToolDependency struct {
	FromTool   string  // tool that produces the value
	FromField  string  // field name in the producer
	ToTool     string  // tool that consumes the value
	ToField    string  // field name in the consumer
	Confidence float64 // 0.0 - 1.0
}

// depFieldInfo holds extracted schema info for a single parameter.
type depFieldInfo struct {
	name       string
	typ        string
	hasEnum    bool
	hasPattern bool
}

// depToolFields holds extracted info for a single tool.
type depToolFields struct {
	name   string
	desc   string
	fields []depFieldInfo
}

// inferDependencies builds a dependency graph between tools by matching
// input parameter names across tools. When tool A has a parameter "user_id"
// and tool B also has "user_id", we infer a potential data flow.
//
// Since MCP tools don't have explicit output schemas, we treat each tool's
// inputs as potential outputs of other tools (the LLM chains them together).
func inferDependencies(tools []mcp.Tool) []ToolDependency {
	var deps []ToolDependency

	toolData := make([]depToolFields, 0, len(tools))
	for _, tool := range tools {
		tf := depToolFields{
			name: tool.Name,
			desc: tool.Description,
		}

		props := tool.InputSchema.Properties
		for name, prop := range props {
			propMap, ok := prop.(map[string]any)
			if !ok {
				continue
			}
			typ, _ := propMap["type"].(string)
			_, hasEnum := propMap["enum"]
			_, hasPattern := propMap["pattern"]
			tf.fields = append(tf.fields, depFieldInfo{
				name:       name,
				typ:        typ,
				hasEnum:    hasEnum,
				hasPattern: hasPattern,
			})
		}

		toolData = append(toolData, tf)
	}

	// For each pair of tools, check if fields match
	for i := range toolData {
		for j := range toolData {
			if i == j {
				continue
			}

			from := toolData[i]
			to := toolData[j]

			for _, fromField := range from.fields {
				for _, toField := range to.fields {
					confidence := fieldMatchConfidence(fromField, toField, from.desc, to.desc)
					if confidence >= 0.6 {
						deps = append(deps, ToolDependency{
							FromTool:   from.name,
							FromField:  fromField.name,
							ToTool:     to.name,
							ToField:    toField.name,
							Confidence: confidence,
						})
					}
				}
			}
		}
	}

	return deps
}

// commonParamNames are exact parameter names too generic to indicate real data flow.
var commonParamNames = map[string]bool{
	"path": true, "file": true, "name": true, "type": true,
	"id": true, "url": true, "dir": true, "directory": true,
	"format": true, "output": true, "input": true, "mode": true,
	"source": true, "destination": true, "target": true,
	"content": true, "pattern": true, "query": true, "text": true,
	"encoding": true, "recursive": true, "limit": true, "offset": true,
	"exclude": true, "include": true, "filter": true, "options": true,
	"command": true, "language": true, "uri": true, "root": true,
}

// commonParamSuffixes are suffix patterns that indicate unconstrained-by-nature params.
var commonParamSuffixes = []string{
	"_path", "_file", "_dir", "_root", "_directory",
	"_id", "_ids", "_uri", "_url",
	"_name", "_command", "_language",
}

// isCommonParam returns true if the parameter name is too generic to indicate real data flow.
func isCommonParam(name string) bool {
	lower := strings.ToLower(name)
	if commonParamNames[lower] {
		return true
	}
	for _, suffix := range commonParamSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

// fieldMatchConfidence calculates how likely two fields represent the same data.
// Requires strong field name match (>= 0.8) as a prerequisite. Description
// overlap alone is insufficient because tools often cross-reference each other
// in documentation ("see also blast_radius") without implying data flow.
func fieldMatchConfidence(from, to depFieldInfo, fromDesc, toDesc string) float64 {
	// Skip common/generic params - they don't indicate real dependencies
	if isCommonParam(from.name) || isCommonParam(to.name) {
		return 0.0
	}

	// Name similarity is the primary signal. Without a strong name match,
	// description overlap alone causes false positives from "see also" references.
	nameSim := paramNameSimilarity(from.name, to.name)
	if nameSim < 0.8 {
		return 0.0 // require strong field name match as prerequisite
	}

	var confidence float64

	// Name similarity (weight: 0.5)
	confidence += nameSim * 0.5

	// Type compatibility (weight: 0.3)
	if from.typ == to.typ && from.typ != "" {
		confidence += 0.3
	} else if typeCompatible(from.typ, to.typ) {
		confidence += 0.15
	}

	// Description token overlap (weight: 0.2, reduced from 0.3)
	if fromDesc != "" && toDesc != "" {
		descSim := tokenJaccard(fromDesc, toDesc)
		confidence += descSim * 0.2
	}

	// Bonus for exact name match + same type
	if nameSim == 1.0 && from.typ == to.typ {
		confidence += 0.1
	}

	// Clamp
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// paramNameSimilarity compares two parameter names.
func paramNameSimilarity(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	if a == b {
		return 1.0
	}

	// Normalize separators
	aNorm := strings.ReplaceAll(strings.ReplaceAll(a, "_", ""), "-", "")
	bNorm := strings.ReplaceAll(strings.ReplaceAll(b, "_", ""), "-", "")
	if aNorm == bNorm {
		return 0.95
	}

	// One contains the other (e.g., "user_id" and "id")
	if len(a) >= 3 && strings.Contains(b, a) {
		return 0.8
	}
	if len(b) >= 3 && strings.Contains(a, b) {
		return 0.8
	}

	// Word overlap
	wordsA := splitWords(a)
	wordsB := splitWords(b)
	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}

	intersection := 0
	for _, w := range wordsA {
		for _, w2 := range wordsB {
			if w == w2 {
				intersection++
				break
			}
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// splitWords splits a parameter name into words (handles snake_case, camelCase).
func splitWords(s string) []string {
	// Split on underscores and hyphens
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Split camelCase
	var result []string
	current := ""
	for i, r := range s {
		if r == ' ' {
			if current != "" {
				result = append(result, strings.ToLower(current))
				current = ""
			}
			continue
		}
		if i > 0 && r >= 'A' && r <= 'Z' && s[i-1] >= 'a' && s[i-1] <= 'z' {
			if current != "" {
				result = append(result, strings.ToLower(current))
			}
			current = string(r)
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, strings.ToLower(current))
	}

	// Filter short words
	var filtered []string
	for _, w := range result {
		if len(w) > 2 {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

// typeCompatible returns true if two types can be coerced.
func typeCompatible(from, to string) bool {
	if from == "" || to == "" {
		return false
	}
	compatible := map[string][]string{
		"string":  {"string"},
		"number":  {"string", "integer"},
		"integer": {"string", "number"},
		"boolean": {"string"},
	}
	for _, t := range compatible[from] {
		if t == to {
			return true
		}
	}
	return false
}

// tokenJaccard computes Jaccard similarity on word tokens.
func tokenJaccard(a, b string) float64 {
	tokensA := make(map[string]bool)
	tokensB := make(map[string]bool)

	for _, w := range strings.Fields(strings.ToLower(a)) {
		if len(w) > 2 {
			tokensA[w] = true
		}
	}
	for _, w := range strings.Fields(strings.ToLower(b)) {
		if len(w) > 2 {
			tokensB[w] = true
		}
	}

	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}

	intersection := 0
	for t := range tokensA {
		if tokensB[t] {
			intersection++
		}
	}

	union := len(tokensA) + len(tokensB) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// --- Lint rules that use the dependency graph ---

// lintCircularDependency detects cycles in the tool dependency graph.
func lintCircularDependency(tools []mcp.Tool) []LintFinding {
	deps := inferDependencies(tools)

	// Filter to high-confidence only (0.75+ to avoid false positives from shared generic fields)
	var highConf []ToolDependency
	for _, d := range deps {
		if d.Confidence >= 0.75 {
			highConf = append(highConf, d)
		}
	}

	if len(highConf) == 0 {
		return nil
	}

	// Build adjacency list
	graph := make(map[string][]string)
	for _, d := range highConf {
		graph[d.FromTool] = append(graph[d.FromTool], d.ToTool)
	}

	// DFS cycle detection
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycles [][]string

	var dfs func(node string, path []string)
	dfs = func(node string, path []string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				dfs(neighbor, path)
			} else if recStack[neighbor] {
				// Found cycle
				start := -1
				for i, n := range path {
					if n == neighbor {
						start = i
						break
					}
				}
				if start >= 0 {
					cycle := append(path[start:], neighbor)
					cycles = append(cycles, cycle)
				}
			}
		}

		recStack[node] = false
	}

	for node := range graph {
		if !visited[node] {
			dfs(node, nil)
		}
	}

	// Deduplicate cycles: normalize by rotating to smallest element, then dedup on node set.
	var findings []LintFinding
	seen := make(map[string]bool)

	for _, cycle := range cycles {
		if len(cycle) < 2 {
			continue
		}
		// Remove the trailing repeated node (A→B→A becomes the set {A,B})
		nodes := cycle[:len(cycle)-1]

		// Canonical key: sorted set of participating nodes (not path order)
		sorted := make([]string, len(nodes))
		copy(sorted, nodes)
		sort.Strings(sorted)
		key := strings.Join(sorted, ",")
		if seen[key] {
			continue
		}
		seen[key] = true

		// Report shortest representation
		findings = append(findings, LintFinding{
			Tool:     sorted[0],
			Code:     report.E107,
			Severity: report.SeverityWarning,
			Message:  fmt.Sprintf("Circular dependency between %s. Agents may loop.", strings.Join(sorted, ", ")),
			Field:    "dependencies",
		})
	}

	return findings
}

// lintFreeTextPropagation detects unconstrained string data flowing between tools.
func lintFreeTextPropagation(tools []mcp.Tool) []LintFinding {
	deps := inferDependencies(tools)

	// Filter to high-confidence
	var highConf []ToolDependency
	for _, d := range deps {
		if d.Confidence >= 0.65 {
			highConf = append(highConf, d)
		}
	}

	if len(highConf) == 0 {
		return nil
	}

	// Build field info map
	type fieldConstraints struct {
		typ        string
		hasEnum    bool
		hasPattern bool
		hasFormat  bool
	}
	fieldMap := make(map[string]map[string]fieldConstraints) // tool -> field -> constraints

	for _, tool := range tools {
		fields := make(map[string]fieldConstraints)
		for name, prop := range tool.InputSchema.Properties {
			propMap, ok := prop.(map[string]any)
			if !ok {
				continue
			}
			typ, _ := propMap["type"].(string)
			_, hasEnum := propMap["enum"]
			_, hasPattern := propMap["pattern"]
			_, hasFormat := propMap["format"]
			fields[name] = fieldConstraints{
				typ:        typ,
				hasEnum:    hasEnum,
				hasPattern: hasPattern,
				hasFormat:  hasFormat,
			}
		}
		fieldMap[tool.Name] = fields
	}

	var findings []LintFinding
	seen := make(map[string]bool)

	for _, dep := range highConf {
		// Check if the FROM field is an unconstrained string
		fromFields, ok := fieldMap[dep.FromTool]
		if !ok {
			continue
		}
		fromField, ok := fromFields[dep.FromField]
		if !ok {
			continue
		}

		if fromField.typ == "string" && !fromField.hasEnum && !fromField.hasPattern && !fromField.hasFormat {
			// Check the TO field is also unconstrained
			toFields, ok := fieldMap[dep.ToTool]
			if !ok {
				continue
			}
			toField, ok := toFields[dep.ToField]
			if !ok {
				continue
			}

			if toField.typ == "string" && !toField.hasEnum && !toField.hasPattern && !toField.hasFormat {
				key := dep.FromTool + ":" + dep.FromField + "->" + dep.ToTool + ":" + dep.ToField
				if seen[key] {
					continue
				}
				seen[key] = true

				findings = append(findings, LintFinding{
					Tool:     dep.FromTool,
					Code:     report.E105,
					Severity: report.SeverityError,
					Message:  fmt.Sprintf("Unconstrained string %q flows from %q to %q. Agents may pass free text where structured data is expected.", dep.FromField, dep.FromTool, dep.ToTool),
					Field:    "args." + dep.FromField,
				})
			}
		}
	}

	return findings
}

// lintNonDeterminism calls a tool 3 times with identical args and compares
// output hashes. If outputs differ, the tool is non-deterministic.
func lintNonDeterminism(ctx context.Context, mcpClient interface {
	CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
}, tool mcp.Tool) *LintFinding {
	const trials = 3

	args := generateArgsFromSchema(tool.InputSchema, "")
	req := mcp.CallToolRequest{}
	req.Params.Name = tool.Name
	req.Params.Arguments = args

	var hashes []string

	for i := 0; i < trials; i++ {
		callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		result, err := mcpClient.CallTool(callCtx, req)
		cancel()

		if err != nil {
			return nil // can't test, skip
		}

		text := extractText(result)
		h := sha256.Sum256([]byte(text))
		hashes = append(hashes, hex.EncodeToString(h[:]))
	}

	// Compare: all hashes should be identical for deterministic tools
	unique := make(map[string]bool)
	for _, h := range hashes {
		unique[h] = true
	}

	if len(unique) > 1 {
		return &LintFinding{
			Tool:     tool.Name,
			Code:     report.W107,
			Severity: report.SeverityWarning,
			Message:  fmt.Sprintf("Tool produced %d different outputs across %d identical calls. Non-deterministic tools are unreliable for agents.", len(unique), trials),
			Field:    "behavior",
		}
	}

	return nil
}
