package fuzz

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGenerateInputs_EmptySchema(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Type:       "object",
		Properties: map[string]any{},
	}
	cases := GenerateInputs(schema, 10, 42)
	if len(cases) != 10 {
		t.Errorf("expected 10 cases, got %d", len(cases))
	}
	// First two should always be structural cases.
	if cases[0].Label != "empty object" {
		t.Errorf("expected first case to be 'empty object', got %q", cases[0].Label)
	}
	if cases[1].Label != "null args" {
		t.Errorf("expected second case to be 'null args', got %q", cases[1].Label)
	}
}

func TestGenerateInputs_StringProperty(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query",
			},
		},
		Required: []string{"query"},
	}
	cases := GenerateInputs(schema, 50, 42)
	if len(cases) != 50 {
		t.Errorf("expected 50 cases, got %d", len(cases))
	}

	// Check that we have category-based cases.
	labels := make(map[string]bool)
	for _, c := range cases {
		labels[c.Label] = true
	}

	// Should include structural cases.
	if !labels["empty object"] {
		t.Error("missing 'empty object' case")
	}
	if !labels["missing required: query"] {
		t.Error("missing 'missing required: query' case")
	}

	// Should include string-specific cases.
	if !labels["query: empty string"] {
		t.Error("missing 'query: empty string' case")
	}
	if !labels["query: path traversal"] {
		t.Error("missing 'query: path traversal' case")
	}
}

func TestGenerateInputs_Reproducible(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]any{
			"name": map[string]any{"type": "string"},
		},
		Required: []string{"name"},
	}

	cases1 := GenerateInputs(schema, 30, 42)
	cases2 := GenerateInputs(schema, 30, 42)

	for i := range cases1 {
		if cases1[i].Label != cases2[i].Label {
			t.Errorf("case %d: labels differ: %q vs %q", i, cases1[i].Label, cases2[i].Label)
		}
	}
}

func TestGenerateInputs_DifferentSeeds(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]any{
			"name": map[string]any{"type": "string"},
		},
		Required: []string{"name"},
	}

	cases1 := GenerateInputs(schema, 30, 42)
	cases2 := GenerateInputs(schema, 30, 99)

	// The category-based cases at the start will be the same,
	// but the random mutations at the end should differ.
	foundDiff := false
	for i := range cases1 {
		if cases1[i].Label != cases2[i].Label {
			foundDiff = true
			break
		}
	}
	if !foundDiff {
		t.Error("expected different seeds to produce different cases")
	}
}

func TestGenerateInputs_MultipleProperties(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]any{
			"query": map[string]any{"type": "string"},
			"limit": map[string]any{"type": "integer"},
		},
		Required: []string{"query", "limit"},
	}
	cases := GenerateInputs(schema, 50, 42)

	// Should have missing-required cases for both properties.
	labels := make(map[string]bool)
	for _, c := range cases {
		labels[c.Label] = true
	}
	if !labels["missing required: query"] {
		t.Error("missing 'missing required: query' case")
	}
	if !labels["missing required: limit"] {
		t.Error("missing 'missing required: limit' case")
	}
	// Should have type-specific cases for both.
	if !labels["query: empty string"] {
		t.Error("missing string cases for query")
	}
	if !labels["limit: zero"] {
		t.Error("missing number cases for limit")
	}
}
