package runner

import (
	"testing"

	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestInferDependencies_MatchingFields(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "get_user",
			Description: "Fetch user by ID",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"user_id": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "update_user",
			Description: "Update user profile",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"user_id": map[string]any{"type": "string"},
					"name":    map[string]any{"type": "string"},
				},
			},
		},
	}

	deps := inferDependencies(tools)
	if len(deps) == 0 {
		t.Fatal("expected dependencies for matching user_id fields")
	}

	// Should find user_id -> user_id edge
	found := false
	for _, d := range deps {
		if d.FromField == "user_id" && d.ToField == "user_id" {
			found = true
			if d.Confidence < 0.6 {
				t.Errorf("confidence too low: %f", d.Confidence)
			}
		}
	}
	if !found {
		t.Error("expected user_id -> user_id dependency")
	}
}

func TestInferDependencies_NoMatch(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "read_file",
			Description: "Read a file",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"path": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "send_email",
			Description: "Send an email message",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"recipient": map[string]any{"type": "string"},
					"body":      map[string]any{"type": "string"},
				},
			},
		},
	}

	deps := inferDependencies(tools)
	// Fields are completely different, should have low/no confidence
	for _, d := range deps {
		if d.Confidence >= 0.6 {
			t.Errorf("unexpected high-confidence dep: %s.%s -> %s.%s (%.2f)",
				d.FromTool, d.FromField, d.ToTool, d.ToField, d.Confidence)
		}
	}
}

func TestLintCircularDependency_Cycle(t *testing.T) {
	// Create tools where A and B share the same field names with high similarity
	tools := []mcp.Tool{
		{
			Name:        "create_order",
			Description: "Create a new order for the customer",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"order_id":    map[string]any{"type": "string"},
					"customer_id": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "get_customer",
			Description: "Get customer details for the order",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"customer_id": map[string]any{"type": "string"},
					"order_id":    map[string]any{"type": "string"},
				},
			},
		},
	}

	findings := lintCircularDependency(tools)
	// Both tools have identical fields, so a cycle is inferred
	for _, f := range findings {
		if f.Code != report.E107 {
			t.Errorf("expected E107, got %s", f.Code)
		}
	}
}

func TestLintCircularDependency_NoCycle(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "get_user",
			Description: "Fetch user data",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"user_id": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "send_notification",
			Description: "Send a push notification",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"message": map[string]any{"type": "string"},
					"target":  map[string]any{"type": "string"},
				},
			},
		},
	}

	findings := lintCircularDependency(tools)
	if len(findings) != 0 {
		t.Errorf("expected no cycles, got %d findings", len(findings))
	}
}

func TestLintFreeTextPropagation(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "search",
			Description: "Search for documents",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"query": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "summarize",
			Description: "Summarize search results",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"query": map[string]any{"type": "string"},
				},
			},
		},
	}

	findings := lintFreeTextPropagation(tools)
	// Both tools have unconstrained "query" string - free text propagation
	for _, f := range findings {
		if f.Code != report.E105 {
			t.Errorf("expected E105, got %s", f.Code)
		}
	}
}

func TestLintFreeTextPropagation_Constrained(t *testing.T) {
	tools := []mcp.Tool{
		{
			Name:        "get_status",
			Description: "Get order status",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"status": map[string]any{
						"type": "string",
						"enum": []string{"pending", "shipped", "delivered"},
					},
				},
			},
		},
		{
			Name:        "filter_orders",
			Description: "Filter orders by status",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"status": map[string]any{
						"type": "string",
						"enum": []string{"pending", "shipped", "delivered"},
					},
				},
			},
		},
	}

	findings := lintFreeTextPropagation(tools)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for constrained fields, got %d", len(findings))
	}
}

func TestParamNameSimilarity(t *testing.T) {
	tests := []struct {
		a, b     string
		minScore float64
	}{
		{"user_id", "user_id", 1.0},
		{"userId", "user_id", 0.9},
		{"user_id", "id", 0.0}, // "id" is too short (2 chars) to match
		{"path", "file_path", 0.7},
		{"query", "search_query", 0.7},
		{"foo", "bar", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			score := paramNameSimilarity(tt.a, tt.b)
			if score < tt.minScore {
				t.Errorf("paramNameSimilarity(%q, %q) = %f, want >= %f", tt.a, tt.b, score, tt.minScore)
			}
		})
	}
}
