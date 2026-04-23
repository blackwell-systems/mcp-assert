package runner

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// Generate queries tools/list and creates stub assertion YAMLs for each tool.
func Generate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	serverSpec := fs.String("server", "", "Server command (e.g. 'agent-lsp go:gopls')")
	output := fs.String("output", "", "Output directory for generated YAML files")
	fixture := fs.String("fixture", "", "Fixture directory to use in generated assertions")
	timeout := fs.Duration("timeout", 15*time.Second, "Timeout for tools/list call")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing YAML files")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *serverSpec == "" || *output == "" {
		return fmt.Errorf("--server and --output are required")
	}

	// Start the server and query tools/list.
	parts := strings.Fields(*serverSpec)
	if len(parts) == 0 {
		return fmt.Errorf("--server cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	mcpClient, err := client.NewStdioMCPClient(parts[0], nil, parts[1:]...)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer mcpClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		return fmt.Errorf("MCP initialize failed: %w", err)
	}

	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("tools/list failed: %w", err)
	}

	// Create output directory.
	if err := os.MkdirAll(*output, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	created, skipped := 0, 0
	for _, tool := range toolsResult.Tools {
		filename := sanitizeFilename(tool.Name) + ".yaml"
		path := filepath.Join(*output, filename)

		if !*overwrite {
			if _, err := os.Stat(path); err == nil {
				skipped++
				continue
			}
		}

		stub := generateStub(tool, *serverSpec, *fixture)
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

	fmt.Printf("\n%d tools discovered, %d assertions created, %d skipped (already exist)\n", len(toolsResult.Tools), created, skipped)
	if created > 0 {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit the generated YAMLs to fill in realistic argument values\n")
		fmt.Printf("  2. Run: mcp-assert snapshot --suite %s --server %q --update\n", *output, *serverSpec)
		fmt.Printf("  3. Run: mcp-assert run --suite %s --server %q\n", *output, *serverSpec)
	}

	return nil
}

// stubAssertion is the YAML structure for a generated assertion file.
type stubAssertion struct {
	Name    string         `yaml:"name"`
	Server  stubServer     `yaml:"server"`
	Setup   []stubToolCall `yaml:"setup,omitempty"`
	Assert  stubAssert     `yaml:"assert"`
	Timeout string         `yaml:"timeout"`
}

type stubServer struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args,flow"`
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

func generateStub(tool mcp.Tool, serverSpec string, fixture string) stubAssertion {
	parts := strings.Fields(serverSpec)
	cmd := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	// Generate placeholder args from the input schema.
	toolArgs := generateArgsFromSchema(tool.InputSchema, fixture)

	stub := stubAssertion{
		Name: fmt.Sprintf("%s returns expected result", tool.Name),
		Server: stubServer{
			Command: cmd,
			Args:    args,
		},
		Assert: stubAssert{
			Tool:   tool.Name,
			Args:   toolArgs,
			Expect: stubExpect{NotError: true},
		},
		Timeout: "30s",
	}

	return stub
}

// generateArgsFromSchema creates placeholder args from a JSON Schema.
func generateArgsFromSchema(schema mcp.ToolInputSchema, fixture string) map[string]any {
	args := make(map[string]any)

	props := schema.Properties
	required := make(map[string]bool)
	for _, r := range schema.Required {
		required[r] = true
	}

	for name, prop := range props {
		// Only generate args for required properties.
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

// generateStringPlaceholder creates a sensible default for string params.
func generateStringPlaceholder(name, desc, fixture string) string {
	nameLower := strings.ToLower(name)

	// Path-like params get fixture prefix.
	if strings.Contains(nameLower, "path") || strings.Contains(nameLower, "dir") ||
		strings.Contains(nameLower, "root") || strings.Contains(nameLower, "file") ||
		strings.Contains(nameLower, "uri") {
		if fixture != "" {
			return "{{fixture}}/TODO"
		}
		return "/path/to/TODO"
	}

	// Query-like params.
	if strings.Contains(nameLower, "query") || strings.Contains(nameLower, "search") {
		return "TODO_QUERY"
	}

	// Name-like params.
	if strings.Contains(nameLower, "name") || strings.Contains(nameLower, "symbol") {
		return "TODO_NAME"
	}

	// Language-like params.
	if strings.Contains(nameLower, "language") || nameLower == "language_id" {
		return "go"
	}

	return "TODO"
}

func sanitizeFilename(name string) string {
	// Replace dots and special chars with underscores.
	r := strings.NewReplacer(".", "_", "/", "_", "\\", "_", " ", "_")
	return r.Replace(name)
}
