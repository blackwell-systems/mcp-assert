package runner

import (
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGenerateArgsFromSchema_Required(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Properties: map[string]interface{}{
			"file_path": map[string]any{"type": "string", "description": "Path to file"},
			"line":      map[string]any{"type": "integer"},
			"column":    map[string]any{"type": "integer"},
			"optional":  map[string]any{"type": "string"},
		},
		Required: []string{"file_path", "line", "column"},
	}

	args := generateArgsFromSchema(schema, "/fixture")

	if args["file_path"] != "{{fixture}}/TODO" {
		t.Errorf("expected {{fixture}}/TODO for file_path, got %v", args["file_path"])
	}
	if args["line"] != 1 {
		t.Errorf("expected 1 for line, got %v", args["line"])
	}
	if args["column"] != 1 {
		t.Errorf("expected 1 for column, got %v", args["column"])
	}
	if _, exists := args["optional"]; exists {
		t.Error("optional param should not be in generated args")
	}
}

func TestGenerateArgsFromSchema_PathDetection(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"file_path", "{{fixture}}/TODO"},
		{"root_dir", "{{fixture}}/TODO"},
		{"workspace_dir", "{{fixture}}/TODO"},
		{"uri", "{{fixture}}/TODO"},
	}

	for _, tt := range tests {
		schema := mcp.ToolInputSchema{
			Properties: map[string]interface{}{
				tt.name: map[string]any{"type": "string"},
			},
			Required: []string{tt.name},
		}
		args := generateArgsFromSchema(schema, "/fixture")
		if args[tt.name] != tt.expected {
			t.Errorf("%s: expected %q, got %v", tt.name, tt.expected, args[tt.name])
		}
	}
}

func TestGenerateArgsFromSchema_NoFixture(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Properties: map[string]interface{}{
			"file_path": map[string]any{"type": "string"},
		},
		Required: []string{"file_path"},
	}

	args := generateArgsFromSchema(schema, "")
	if args["file_path"] != "/path/to/TODO" {
		t.Errorf("expected /path/to/TODO without fixture, got %v", args["file_path"])
	}
}

func TestGenerateArgsFromSchema_AllTypes(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Properties: map[string]interface{}{
			"str":  map[string]any{"type": "string"},
			"num":  map[string]any{"type": "number"},
			"int":  map[string]any{"type": "integer"},
			"bool": map[string]any{"type": "boolean"},
			"arr":  map[string]any{"type": "array"},
			"obj":  map[string]any{"type": "object"},
		},
		Required: []string{"str", "num", "int", "bool", "arr", "obj"},
	}

	args := generateArgsFromSchema(schema, "")

	if _, ok := args["str"].(string); !ok {
		t.Errorf("str should be string, got %T", args["str"])
	}
	if args["num"] != 1 {
		t.Errorf("num should be 1, got %v", args["num"])
	}
	if args["int"] != 1 {
		t.Errorf("int should be 1, got %v", args["int"])
	}
	if args["bool"] != true {
		t.Errorf("bool should be true, got %v", args["bool"])
	}
	if _, ok := args["arr"].([]any); !ok {
		t.Errorf("arr should be []any, got %T", args["arr"])
	}
	if _, ok := args["obj"].(map[string]any); !ok {
		t.Errorf("obj should be map[string]any, got %T", args["obj"])
	}
}

func TestGenerateArgsFromSchema_Empty(t *testing.T) {
	schema := mcp.ToolInputSchema{}
	args := generateArgsFromSchema(schema, "")
	if len(args) != 0 {
		t.Errorf("expected 0 args for empty schema, got %d", len(args))
	}
}

func TestGenerateStub(t *testing.T) {
	tool := mcp.Tool{
		Name:        "get_references",
		Description: "Find all references to a symbol",
		InputSchema: mcp.ToolInputSchema{
			Properties: map[string]interface{}{
				"file_path": map[string]any{"type": "string"},
				"line":      map[string]any{"type": "integer"},
				"column":    map[string]any{"type": "integer"},
			},
			Required: []string{"file_path", "line", "column"},
		},
	}

	stub := generateStub(tool, "agent-lsp go:gopls", "/fixture", false)

	if stub.Name != "get_references returns expected result" {
		t.Errorf("unexpected name: %q", stub.Name)
	}
	if stub.Server.Command != "agent-lsp" {
		t.Errorf("unexpected command: %q", stub.Server.Command)
	}
	if len(stub.Server.Args) != 1 || stub.Server.Args[0] != "go:gopls" {
		t.Errorf("unexpected args: %v", stub.Server.Args)
	}
	if stub.Assert.Tool != "get_references" {
		t.Errorf("unexpected tool: %q", stub.Assert.Tool)
	}
	if !stub.Assert.Expect.NotError {
		t.Error("expected not_error: true")
	}
	if stub.Timeout != "30s" {
		t.Errorf("unexpected timeout: %q", stub.Timeout)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"get_references", "get_references"},
		{"gopls.list_known_packages", "gopls_list_known_packages"},
		{"my/tool", "my_tool"},
		{"has spaces", "has_spaces"},
	}
	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.expected {
			t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGenerate_MissingFlags(t *testing.T) {
	err := Generate([]string{})
	if err == nil {
		t.Error("expected error for missing flags")
	}

	err = Generate([]string{"--server", "test"})
	if err == nil {
		t.Error("expected error for missing --output")
	}
}

func boolPtr(b bool) *bool { return &b }

func TestIsDestructiveTool_DestructiveTrue(t *testing.T) {
	tool := mcp.Tool{
		Annotations: mcp.ToolAnnotation{DestructiveHint: boolPtr(true)},
	}
	if !isDestructiveTool(tool) {
		t.Error("expected true for DestructiveHint=true")
	}
}

func TestIsDestructiveTool_ReadOnlyFalse(t *testing.T) {
	tool := mcp.Tool{
		Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(false)},
	}
	if !isDestructiveTool(tool) {
		t.Error("expected true for ReadOnlyHint=false")
	}
}

func TestIsDestructiveTool_ReadOnlyTrue(t *testing.T) {
	tool := mcp.Tool{
		Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)},
	}
	if isDestructiveTool(tool) {
		t.Error("expected false for ReadOnlyHint=true")
	}
}

func TestIsDestructiveTool_NoAnnotations(t *testing.T) {
	tool := mcp.Tool{}
	if isDestructiveTool(tool) {
		t.Error("expected false for zero-value annotations")
	}
}

func TestIsDestructiveTool_BothSet(t *testing.T) {
	tool := mcp.Tool{
		Annotations: mcp.ToolAnnotation{
			DestructiveHint: boolPtr(true),
			ReadOnlyHint:    boolPtr(true),
		},
	}
	if !isDestructiveTool(tool) {
		t.Error("expected true when DestructiveHint=true even with ReadOnlyHint=true")
	}
}

func TestGenerateStub_SkipTrue(t *testing.T) {
	tool := mcp.Tool{
		Name: "write_file",
		InputSchema: mcp.ToolInputSchema{},
	}
	stub := generateStub(tool, "server arg1", "", true)
	if !stub.Skip {
		t.Error("expected Skip=true")
	}
}

func TestGenerateStub_SkipFalse(t *testing.T) {
	tool := mcp.Tool{
		Name: "read_file",
		InputSchema: mcp.ToolInputSchema{},
	}
	stub := generateStub(tool, "server arg1", "", false)
	if stub.Skip {
		t.Error("expected Skip=false")
	}
}

func TestIsTransportError(t *testing.T) {
	tests := []struct {
		msg    string
		expect bool
	}{
		{"transport closed", true},
		{"EOF", true},
		{"connection refused", true},
		{"some other error", false},
		{"read: transport closed unexpectedly", true},
		{"dial tcp: connection refused", true},
	}
	for _, tt := range tests {
		err := fmt.Errorf("%s", tt.msg)
		got := isTransportError(err)
		if got != tt.expect {
			t.Errorf("isTransportError(%q) = %v, want %v", tt.msg, got, tt.expect)
		}
	}
}
