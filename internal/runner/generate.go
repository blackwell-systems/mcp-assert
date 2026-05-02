// generate.go handles two related commands: Generate (CLI entry point) and
// GenerateCore (reusable logic shared with Init).
//
// The flow:
//   1. Connect to the MCP server and run tools/list to discover all tools.
//   2. For each tool, build a stub YAML assertion from the tool's input schema.
//   3. Write one YAML file per tool into the output directory.
//
// Stub generation uses heuristics to pick sensible placeholder values for each
// argument type (see generateStringPlaceholder). Destructive tools are marked
// skip:true by default unless --include-writes is passed.
//
// This file also contains shared helpers used by other commands:
//   - generateArgsFromSchema: builds placeholder args from a JSON Schema (used by audit)
//   - isDestructiveTool: checks MCP tool annotations (used by audit, generate)
//   - isTransportError: detects server startup failures (used by generate, coverage)
//   - parseHeadersFlag: parses CLI header flags (used by audit, fuzz, generate)
//   - sanitizeFilename: converts tool names to safe filenames
package runner

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// GenerateResult holds the outcome of a generate operation.
type GenerateResult struct {
	ToolCount int
	Created   int
	Skipped   int
}

// GenerateOpts configures the generate operation. Used by both Generate (CLI)
// and Init (which calls GenerateCore internally).
type GenerateOpts struct {
	ServerSpec    string
	Output        string
	Fixture       string
	Timeout       time.Duration
	Overwrite     bool
	IncludeWrites bool
	Transport     string            // "stdio" (default), "http", "sse"
	Headers       map[string]string // custom headers for http/sse
}

// GenerateCore queries tools/list and creates stub assertion YAMLs. It returns
// a result summary without printing next-steps guidance (the caller decides).
// This is the shared implementation used by both the generate and init commands.
func GenerateCore(opts GenerateOpts) (*GenerateResult, error) {
	if opts.ServerSpec == "" {
		return nil, fmt.Errorf("--server cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// Build server config from options. This duplicates serverFlags.serverConfig()
	// because GenerateCore accepts GenerateOpts (not serverFlags) to support
	// programmatic callers like Init that don't go through CLI flag parsing.
	transport := strings.ToLower(opts.Transport)
	var serverCfg assertion.ServerConfig
	if transport == "http" || transport == "sse" {
		serverCfg = assertion.ServerConfig{
			Transport: transport,
			URL:       opts.ServerSpec,
			Headers:   opts.Headers,
		}
	} else {
		parts := strings.Fields(opts.ServerSpec)
		if len(parts) == 0 {
			return nil, fmt.Errorf("--server cannot be empty")
		}
		serverCfg = assertion.ServerConfig{
			Command: parts[0],
			Args:    parts[1:],
		}
	}

	mcpClient, _, err := connectAndInitialize(serverCfg)
	if err != nil {
		return nil, err
	}
	defer mcpClient.Close()

	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		if isTransportError(err) {
			return nil, fmt.Errorf("tools/list failed: %w\n\nhint: the server exited immediately. Check that any required environment variables (API keys, tokens) are set", err)
		}
		return nil, fmt.Errorf("tools/list failed: %w", err)
	}

	if err := os.MkdirAll(opts.Output, 0755); err != nil {
		return nil, fmt.Errorf("creating output dir: %w", err)
	}

	created, skipped := 0, 0
	for _, tool := range toolsResult.Tools {
		filename := sanitizeFilename(tool.Name) + ".yaml"
		path := filepath.Join(opts.Output, filename)

		// Skip existing files unless --overwrite is set.
		if !opts.Overwrite {
			if _, err := os.Stat(path); err == nil {
				skipped++
				continue
			}
		}

		// Mark destructive tools as skip:true so they don't run in CI by default.
		skip := !opts.IncludeWrites && isDestructiveTool(tool)
		stub := generateStub(tool, opts.ServerSpec, opts.Fixture, skip, opts)
		data, err := yaml.Marshal(stub)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", tool.Name, err)
			continue
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", tool.Name, err)
			continue
		}
		created++
		fmt.Printf("  created %s\n", filename)
	}

	result := &GenerateResult{
		ToolCount: len(toolsResult.Tools),
		Created:   created,
		Skipped:   skipped,
	}

	// Inform user about skipped destructive tools.
	if !opts.IncludeWrites {
		destructive := 0
		for _, tool := range toolsResult.Tools {
			if isDestructiveTool(tool) {
				destructive++
			}
		}
		if destructive > 0 {
			fmt.Printf("  %d tool(s) marked skip:true (destructive). Use --include-writes to include them.\n", destructive)
		}
	}

	return result, nil
}

// Generate is the CLI entry point for the generate command. Parses flags and
// delegates to GenerateCore, then prints next-steps guidance.
func Generate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	serverSpec := fs.String("server", "", "Server command (stdio) or URL (http/sse)")
	output := fs.String("output", "", "Output directory for generated YAML files")
	fixture := fs.String("fixture", "", "Fixture directory to use in generated assertions")
	timeout := fs.Duration("timeout", 15*time.Second, "Timeout for tools/list call")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing YAML files")
	includeWrites := fs.Bool("include-writes", false, "Include write/destructive tools (skipped by default)")
	transport := fs.String("transport", "stdio", "Transport type: stdio (default), http, sse")
	headersFlag := fs.String("headers", "", "Custom headers as key=value pairs, comma-separated (e.g. 'Authorization=Bearer tok')")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *serverSpec == "" || *output == "" {
		return fmt.Errorf("--server and --output are required")
	}

	headers := parseHeadersFlag(*headersFlag)

	result, err := GenerateCore(GenerateOpts{
		ServerSpec:    *serverSpec,
		Output:        *output,
		Fixture:       *fixture,
		Timeout:       *timeout,
		Overwrite:     *overwrite,
		IncludeWrites: *includeWrites,
		Transport:     *transport,
		Headers:       headers,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n%d tools discovered, %d assertions created, %d skipped (already exist)\n", result.ToolCount, result.Created, result.Skipped)
	if result.Created > 0 {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit the generated YAMLs to fill in realistic argument values\n")
		fmt.Printf("  2. Run: mcp-assert snapshot --suite %s --server %q --update\n", *output, *serverSpec)
		fmt.Printf("  3. Run: mcp-assert run --suite %s --server %q\n", *output, *serverSpec)
	}

	return nil
}

// --- Stub YAML types ---

// stubAssertion is the YAML structure written to each generated file.
type stubAssertion struct {
	Name    string         `yaml:"name"`
	Server  stubServer     `yaml:"server"`
	Setup   []stubToolCall `yaml:"setup,omitempty"`
	Assert  stubAssert     `yaml:"assert"`
	Timeout string         `yaml:"timeout"`
	Skip    bool           `yaml:"skip,omitempty"`
}

type stubServer struct {
	Command   string            `yaml:"command,omitempty"`
	Args      []string          `yaml:"args,omitempty,flow"`
	Transport string            `yaml:"transport,omitempty"`
	URL       string            `yaml:"url,omitempty"`
	Headers   map[string]string `yaml:"headers,omitempty"`
}

type stubToolCall struct {
	Tool string         `yaml:"tool"`
	Args map[string]any `yaml:"args"`
}

type stubAssert struct {
	Tool   string         `yaml:"tool"`
	Args   map[string]any `yaml:"args"`
	Expect stubExpect     `yaml:"expect"`
}

type stubExpect struct {
	NotError bool `yaml:"not_error"`
}

// generateStub builds a complete stub YAML assertion for a single tool.
// The stub uses generateArgsFromSchema to create placeholder values and
// sets expect.not_error as the only expectation.
func generateStub(tool mcp.Tool, serverSpec string, fixture string, skip bool, opts ...GenerateOpts) stubAssertion {
	toolArgs := generateArgsFromSchema(tool.InputSchema, fixture)

	var server stubServer
	var genOpts GenerateOpts
	if len(opts) > 0 {
		genOpts = opts[0]
	}

	// Build server block matching the transport type.
	transport := strings.ToLower(genOpts.Transport)
	if transport == "http" || transport == "sse" {
		server = stubServer{
			Transport: transport,
			URL:       serverSpec,
			Headers:   genOpts.Headers,
		}
	} else {
		parts := strings.Fields(serverSpec)
		cmd := parts[0]
		var args []string
		if len(parts) > 1 {
			args = parts[1:]
		}
		server = stubServer{
			Command: cmd,
			Args:    args,
		}
	}

	stub := stubAssertion{
		Name:    fmt.Sprintf("%s returns expected result", tool.Name),
		Server:  server,
		Assert: stubAssert{
			Tool:   tool.Name,
			Args:   toolArgs,
			Expect: stubExpect{NotError: true},
		},
		Timeout: "30s",
	}

	stub.Skip = skip

	return stub
}

// --- Schema-based argument generation ---

// generateArgsFromSchema creates placeholder arguments from a tool's JSON Schema.
// Only required properties get values. Optional properties are omitted so the
// generated YAML stays minimal. Used by both generate (stub creation) and
// audit (schema-based input for health checks).
func generateArgsFromSchema(schema mcp.ToolInputSchema, fixture string) map[string]any {
	args := make(map[string]any)

	props := schema.Properties
	required := make(map[string]bool)
	for _, r := range schema.Required {
		required[r] = true
	}

	for name, prop := range props {
		if !required[name] {
			continue
		}

		propMap, ok := prop.(map[string]any)
		if !ok {
			args[name] = "TODO"
			continue
		}

		typ, _ := propMap["type"].(string)
		desc, _ := propMap["description"].(string)

		switch typ {
		case "string":
			args[name] = generateStringPlaceholder(name, desc, fixture)
		case "integer", "number":
			args[name] = 1
		case "boolean":
			args[name] = true
		case "array":
			args[name] = []any{}
		case "object":
			args[name] = map[string]any{}
		default:
			args[name] = "TODO"
		}
	}

	return args
}

// generateStringPlaceholder picks a sensible default value for a string parameter
// based on its name and description. Uses heuristics: path-like names get fixture
// paths, query-like names get "TODO_QUERY", etc.
func generateStringPlaceholder(name, desc, fixture string) string {
	nameLower := strings.ToLower(name)

	if strings.Contains(nameLower, "path") || strings.Contains(nameLower, "dir") ||
		strings.Contains(nameLower, "root") || strings.Contains(nameLower, "file") ||
		strings.Contains(nameLower, "uri") {
		if fixture != "" {
			return "{{fixture}}/TODO"
		}
		return "/path/to/TODO"
	}

	if strings.Contains(nameLower, "query") || strings.Contains(nameLower, "search") {
		return "TODO_QUERY"
	}

	if strings.Contains(nameLower, "name") || strings.Contains(nameLower, "symbol") {
		return "TODO_NAME"
	}

	if strings.Contains(nameLower, "language") || nameLower == "language_id" {
		return "go"
	}

	return "TODO"
}

// --- Shared helpers ---

// isDestructiveTool returns true if the tool's MCP annotations indicate it
// performs write or destructive operations. Used by audit and generate to
// skip dangerous tools by default.
func isDestructiveTool(tool mcp.Tool) bool {
	a := tool.Annotations
	if a.DestructiveHint != nil && *a.DestructiveHint {
		return true
	}
	// readOnlyHint: false means "may modify state." Over-conservative but safe.
	if a.ReadOnlyHint != nil && !*a.ReadOnlyHint {
		return true
	}
	return false
}

// isTransportError checks if an error indicates the server exited or failed
// to start, which often means missing auth credentials or a bad command.
func isTransportError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "transport closed") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "connection refused")
}

// parseHeadersFlag parses "Key=Value,Key2=Value2" into a map. Returns nil
// for empty input. Used by audit, fuzz, and generate for the --headers flag.
func parseHeadersFlag(raw string) map[string]string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	headers := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

// sanitizeFilename converts a tool name into a safe filename by replacing
// dots, slashes, backslashes, and spaces with underscores.
func sanitizeFilename(name string) string {
	r := strings.NewReplacer(".", "_", "/", "_", "\\", "_", " ", "_")
	return r.Replace(name)
}
