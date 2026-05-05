// lint.go implements the lint command: static analysis of MCP tool schemas.
//
// The lint command connects to an MCP server, runs tools/list, and checks each
// tool's schema for common issues that cause agent misuse:
//   - Missing tool description (agent can't decide when to use the tool)
//   - Missing parameter descriptions (agent guesses values)
//   - Missing parameter types (schema can't validate input)
//   - Parameters without examples or enums (agent hallucinates values)
//   - Oversized default responses (checked during optional --call-tools mode)
//
// Usage:
//
//	mcp-assert lint --server <command-or-url> [--json] [--threshold <max-issues>]
package runner

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// LintSeverity indicates how critical a lint finding is.
type LintSeverity string

const (
	LintError   LintSeverity = "error"   // will cause agent misuse
	LintWarning LintSeverity = "warning" // may cause agent confusion
)

// LintFinding is a single issue found during schema analysis.
type LintFinding struct {
	Tool     string       `json:"tool"`
	Code     string       `json:"code"`
	Severity LintSeverity `json:"severity"`
	Message  string       `json:"message"`
	Field    string       `json:"field,omitempty"` // e.g. "args.query", "description"
}

// LintReport is the full output of a lint run.
type LintReport struct {
	Server   string        `json:"server"`
	Tools    int           `json:"tools"`
	Findings []LintFinding `json:"findings"`
	Errors   int           `json:"errors"`
	Warnings int           `json:"warnings"`
}

// Lint connects to an MCP server, discovers tools, and checks each tool's
// schema for common issues that cause agents to misuse tools.
func Lint(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	var sf serverFlags
	sf.register(fs)
	threshold := fs.Int("threshold", 0, "Fail if total findings exceed this count (0 = no limit)")
	callTools := fs.Bool("call-tools", false, "Also call each tool with empty args to check response size")
	maxResponseKB := fs.Int("max-response-kb", 100, "Maximum acceptable response size in KB (with --call-tools)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if sf.server == "" {
		return fmt.Errorf("--server is required\n\nUsage: mcp-assert lint --server <command-or-url> [--json] [--threshold N]")
	}

	serverCfg, err := sf.serverConfig()
	if err != nil {
		return err
	}

	if !sf.jsonOut {
		fmt.Fprintf(os.Stderr, "Connecting to server...\n")
	}
	mcpClient, initResult, err := connectAndInitialize(serverCfg, connectOpts{})
	if err != nil {
		return err
	}
	defer mcpClient.Close()

	// Discover tools.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("tools/list failed: %w", err)
	}

	serverName := initResult.ServerInfo.Name
	if serverName == "" {
		serverName = sf.server
	}

	if !sf.jsonOut {
		fmt.Fprintf(os.Stderr, "Server: %s\n", serverName)
		fmt.Fprintf(os.Stderr, "Tools:  %d discovered\n\n", len(toolsResult.Tools))
	}

	// Run lint checks on each tool.
	var findings []LintFinding
	for _, tool := range toolsResult.Tools {
		findings = append(findings, lintTool(tool)...)
	}

	// Cross-tool checks (require the full tool list).
	findings = append(findings, lintToolSimilarity(toolsResult.Tools)...)
	findings = append(findings, lintSchemaBloat(toolsResult.Tools)...)

	// Optional: call each tool with empty args to check response size.
	if *callTools {
		for _, tool := range toolsResult.Tools {
			if isDestructiveTool(tool) {
				continue
			}
			finding := lintResponseSize(ctx, mcpClient, tool, *maxResponseKB)
			if finding != nil {
				findings = append(findings, *finding)
			}
		}
	}

	// Build report.
	errors, warnings := 0, 0
	for _, f := range findings {
		if f.Severity == LintError {
			errors++
		} else {
			warnings++
		}
	}
	report := LintReport{
		Server:   serverName,
		Tools:    len(toolsResult.Tools),
		Findings: findings,
		Errors:   errors,
		Warnings: warnings,
	}

	// Output.
	if sf.jsonOut {
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
	} else {
		printLintReport(report)
	}

	// Exit code.
	if *threshold > 0 && len(findings) > *threshold {
		return fmt.Errorf("lint found %d issues (threshold: %d)", len(findings), *threshold)
	}
	if errors > 0 {
		return fmt.Errorf("lint found %d error(s)", errors)
	}

	return nil
}

// --- Lint checks ---

// lintTool runs all schema checks on a single tool.
func lintTool(tool mcp.Tool) []LintFinding {
	var findings []LintFinding

	// E101: Missing tool description.
	if strings.TrimSpace(tool.Description) == "" {
		findings = append(findings, LintFinding{
			Tool:     tool.Name,
			Code:     "E101",
			Severity: LintError,
			Message:  "Tool has no description. Agents cannot determine when to use this tool.",
			Field:    "description",
		})
	}

	// W101: Generic/vague description.
	if isVagueDescription(tool.Description) {
		findings = append(findings, LintFinding{
			Tool:     tool.Name,
			Code:     "W101",
			Severity: LintWarning,
			Message:  fmt.Sprintf("Tool description is too generic: %q. Be specific about what the tool does and when to use it.", tool.Description),
			Field:    "description",
		})
	}

	// Check input schema properties.
	props := tool.InputSchema.Properties
	required := make(map[string]bool)
	for _, r := range tool.InputSchema.Required {
		required[r] = true
	}

	if len(props) == 0 && len(tool.InputSchema.Required) == 0 {
		// No parameters is fine (e.g. list_all_users). Skip property checks.
		return findings
	}

	for name, prop := range props {
		propMap, ok := prop.(map[string]any)
		if !ok {
			continue
		}

		// E102: Missing parameter type.
		if _, hasType := propMap["type"]; !hasType {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     "E102",
				Severity: LintError,
				Message:  fmt.Sprintf("Parameter %q has no type defined. Agents will send wrong value types.", name),
				Field:    "args." + name + ".type",
			})
		}

		// E103: Required parameter with no description.
		desc, _ := propMap["description"].(string)
		if required[name] && strings.TrimSpace(desc) == "" {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     "E103",
				Severity: LintError,
				Message:  fmt.Sprintf("Required parameter %q has no description. Agents will guess what value to provide.", name),
				Field:    "args." + name + ".description",
			})
		}

		// W102: Parameter has no description (non-required).
		if !required[name] && strings.TrimSpace(desc) == "" {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     "W102",
				Severity: LintWarning,
				Message:  fmt.Sprintf("Optional parameter %q has no description.", name),
				Field:    "args." + name + ".description",
			})
		}

		// W104: Generic parameter name.
		if isGenericParamName(name) && strings.TrimSpace(desc) == "" {
			findings = append(findings, LintFinding{
				Tool:     tool.Name,
				Code:     "W104",
				Severity: LintWarning,
				Message:  fmt.Sprintf("Parameter %q is a generic name with no description. Agents cannot infer what to pass.", name),
				Field:    "args." + name,
			})
		}

		// W103: String parameter with no enum, pattern, or example.
		typ, _ := propMap["type"].(string)
		if typ == "string" && required[name] {
			_, hasEnum := propMap["enum"]
			_, hasPattern := propMap["pattern"]
			_, hasExample := propMap["examples"]
			_, hasDefault := propMap["default"]
			if !hasEnum && !hasPattern && !hasExample && !hasDefault {
				findings = append(findings, LintFinding{
					Tool:     tool.Name,
					Code:     "W103",
					Severity: LintWarning,
					Message:  fmt.Sprintf("Required string parameter %q has no enum, pattern, example, or default. Agents may hallucinate values.", name),
					Field:    "args." + name,
				})
			}
		}
	}

	return findings
}

// lintResponseSize calls a tool with empty/minimal args and checks response size.
func lintResponseSize(ctx context.Context, mcpClient interface {
	CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
}, tool mcp.Tool, maxKB int) *LintFinding {
	req := mcp.CallToolRequest{}
	req.Params.Name = tool.Name
	req.Params.Arguments = generateArgsFromSchema(tool.InputSchema, "")

	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := mcpClient.CallTool(callCtx, req)
	if err != nil {
		return nil // can't check, skip
	}

	// Measure response size.
	text := extractText(result)
	sizeKB := len(text) / 1024
	if sizeKB > maxKB {
		return &LintFinding{
			Tool:     tool.Name,
			Code:     "E301",
			Severity: LintError,
			Message:  fmt.Sprintf("Response is %dKB (limit: %dKB). Large responses blow the agent's context window.", sizeKB, maxKB),
			Field:    "response",
		}
	}

	return nil
}

// isVagueDescription returns true if the description is too generic to be useful.
func isVagueDescription(desc string) bool {
	if desc == "" {
		return false // handled by E101
	}
	lower := strings.ToLower(strings.TrimSpace(desc))
	vague := []string{
		"get data",
		"process data",
		"do something",
		"execute",
		"run",
		"handle",
		"perform action",
		"utility",
		"helper",
	}
	for _, v := range vague {
		if lower == v {
			return true
		}
	}
	// Very short descriptions (under 10 chars) are likely too vague.
	if len(lower) < 10 && !strings.Contains(lower, " ") {
		return true
	}
	return false
}

// genericParamNames is the blocklist of parameter names too vague for agents.
var genericParamNames = map[string]bool{
	"data": true, "value": true, "input": true, "output": true,
	"payload": true, "info": true, "content": true, "body": true,
	"params": true, "args": true, "options": true, "config": true,
	"settings": true, "result": true, "object": true, "item": true,
	"obj": true, "val": true, "str": true, "num": true,
}

// isGenericParamName returns true if the parameter name is too vague for agents.
func isGenericParamName(name string) bool {
	return genericParamNames[strings.ToLower(name)]
}

// lintToolSimilarity checks for pairs of tools with very similar descriptions.
// When two tools have >80% similar descriptions, agents pick between them randomly.
func lintToolSimilarity(tools []mcp.Tool) []LintFinding {
	var findings []LintFinding

	for i := 0; i < len(tools); i++ {
		for j := i + 1; j < len(tools); j++ {
			a := strings.TrimSpace(tools[i].Description)
			b := strings.TrimSpace(tools[j].Description)
			if a == "" || b == "" {
				continue // already caught by E101
			}
			sim := stringSimilarity(a, b)
			if sim >= 0.80 {
				findings = append(findings, LintFinding{
					Tool:     tools[i].Name,
					Code:     "W105",
					Severity: LintWarning,
					Message:  fmt.Sprintf("Tool %q and %q have %.0f%% similar descriptions. Agents may confuse them.", tools[i].Name, tools[j].Name, sim*100),
					Field:    "description",
				})
			}
		}
	}

	return findings
}

// stringSimilarity returns a 0-1 similarity score between two strings using
// bigram overlap (Dice coefficient). Fast, no external dependencies.
func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	bigramsA := makeBigrams(a)
	bigramsB := makeBigrams(b)

	if len(bigramsA) == 0 || len(bigramsB) == 0 {
		return 0.0
	}

	intersection := 0
	for bg := range bigramsA {
		if bigramsB[bg] {
			intersection++
		}
	}

	return 2.0 * float64(intersection) / float64(len(bigramsA)+len(bigramsB))
}

func makeBigrams(s string) map[string]bool {
	bigrams := make(map[string]bool)
	for i := 0; i < len(s)-1; i++ {
		bigrams[s[i:i+2]] = true
	}
	return bigrams
}

// lintSchemaBloat measures the total token cost of the tools/list response
// and warns if it exceeds a reasonable budget for an agent's context window.
func lintSchemaBloat(tools []mcp.Tool) []LintFinding {
	totalBytes := 0
	for _, tool := range tools {
		// Approximate the JSON size of each tool definition.
		data, err := json.Marshal(tool)
		if err != nil {
			continue
		}
		totalBytes += len(data)
	}

	// 1 token per 4 bytes (standard approximation for code/JSON).
	tokens := totalBytes / 4
	// Warn if tools/list consumes more than 8K tokens (a significant chunk
	// of a typical agent's context budget).
	if tokens > 8000 {
		return []LintFinding{{
			Tool:     "(all tools)",
			Code:     "W106",
			Severity: LintWarning,
			Message:  fmt.Sprintf("tools/list response is ~%d tokens (%dKB). This consumes a significant portion of the agent's context window.", tokens, totalBytes/1024),
			Field:    "schema_size",
		}}
	}

	return nil
}

// printLintReport outputs a human-readable lint report to stderr/stdout.
func printLintReport(report LintReport) {
	if len(report.Findings) == 0 {
		fmt.Printf("✓ %s: %d tools, no issues found\n", report.Server, report.Tools)
		return
	}

	fmt.Printf("%s: %d tools, %d issues (%d errors, %d warnings)\n\n",
		report.Server, report.Tools, len(report.Findings), report.Errors, report.Warnings)

	for _, f := range report.Findings {
		prefix := "W"
		if f.Severity == LintError {
			prefix = "E"
		}
		fmt.Printf("  %s  %-5s  %-30s  %s\n", prefix, f.Code, f.Tool, f.Message)
	}

	fmt.Printf("\n%d error(s), %d warning(s)\n", report.Errors, report.Warnings)
}
