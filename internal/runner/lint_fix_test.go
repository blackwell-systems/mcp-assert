package runner

import (
	"testing"

	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestGenerateFixes_E103(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "get_user",
			Description: "Get a user",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"user_id": map[string]any{"type": "string"},
				},
				Required: []string{"user_id"},
			},
		},
	}

	findings := []LintFinding{
		{Tool: "get_user", Code: report.E103, Field: "args.user_id.description"},
	}

	fixes := generateFixes(tools, findings)
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	if fixes[0].Action != "set_description" {
		t.Errorf("action = %q, want set_description", fixes[0].Action)
	}
	if fixes[0].Value == nil || fixes[0].Value == "" {
		t.Error("expected non-empty description suggestion")
	}
}

func TestGenerateFixes_E101(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "create_user",
			Description: "",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"name":  map[string]any{"type": "string"},
					"email": map[string]any{"type": "string"},
				},
			},
		},
	}

	findings := []LintFinding{
		{Tool: "create_user", Code: report.E101, Field: "description"},
	}

	fixes := generateFixes(tools, findings)
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	desc, ok := fixes[0].Value.(string)
	if !ok || desc == "" {
		t.Fatal("expected non-empty string description")
	}
	if fixes[0].Action != "set_description" {
		t.Errorf("action = %q, want set_description", fixes[0].Action)
	}
}

func TestGenerateFixes_W109(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "search",
			Description: "Search documents",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"query": map[string]any{"type": "string"},
				},
			},
		},
	}

	findings := []LintFinding{
		{Tool: "search", Code: report.W109, Field: "args.query.examples"},
	}

	fixes := generateFixes(tools, findings)
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	if fixes[0].Action != "set_examples" {
		t.Errorf("action = %q, want set_examples", fixes[0].Action)
	}
	examples, ok := fixes[0].Value.([]string)
	if !ok || len(examples) == 0 {
		t.Error("expected non-empty examples")
	}
}

func TestGenerateFixes_W116(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "list_users",
			Description: "List all users in the system",
			InputSchema: mcp.ToolInputSchema{},
		},
	}

	findings := []LintFinding{
		{Tool: "list_users", Code: report.W116, Field: "description"},
	}

	fixes := generateFixes(tools, findings)
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	val, ok := fixes[0].Value.(string)
	if !ok || val == "" {
		t.Fatal("expected non-empty return clause")
	}
	if fixes[0].Action != "append_description" {
		t.Errorf("action = %q, want append_description", fixes[0].Action)
	}
}

func TestInferFormat(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"email", "email"},
		{"user_email", "email"},
		{"url", "uri"},
		{"callback_url", "uri"},
		{"created_at", "date-time"},
		{"timestamp", "date-time"},
		{"user_id", "uuid"},
		{"session_id", "uuid"},
		{"hostname", "hostname"},
		{"foo_bar", ""},
		{"count", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferFormat(tt.name)
			if got != tt.want {
				t.Errorf("inferFormat(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestInferExamples(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{"email", 1},
		{"query", 1},
		{"url", 1},
		{"timezone", 1},
		{"username", 1},
		{"random_thing", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferExamples(tt.name)
			if len(got) != tt.wantLen {
				t.Errorf("inferExamples(%q) len = %d, want %d", tt.name, len(got), tt.wantLen)
			}
		})
	}
}

func TestInferReturnClause(t *testing.T) {
	tests := []struct {
		name     string
		contains string
	}{
		{"get_user", "Returns"},
		{"list_items", "array"},
		{"create_record", "created"},
		{"delete_file", "deletion"},
		{"search_documents", "array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferReturnClause(tt.name)
			if got == "" {
				t.Error("expected non-empty return clause")
			}
			if !contains(got, tt.contains) {
				t.Errorf("inferReturnClause(%q) = %q, expected to contain %q", tt.name, got, tt.contains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestGenerateParamDescription(t *testing.T) {
	tests := []struct {
		name, typ, tool string
		wantNonEmpty    bool
	}{
		{"user_id", "string", "get_user", true},
		{"email", "string", "send_email", true},
		{"path", "string", "read_file", true},
		{"xyz_abc", "number", "do_thing", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateParamDescription(tt.name, tt.typ, tt.tool)
			if tt.wantNonEmpty && got == "" {
				t.Errorf("expected non-empty description for %q", tt.name)
			}
		})
	}
}
