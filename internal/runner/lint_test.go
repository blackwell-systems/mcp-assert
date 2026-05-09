package runner

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestIsVagueDescription(t *testing.T) {
	tests := []struct {
		desc string
		want bool
	}{
		{"", false},                        // empty handled by E101
		{"get data", true},                 // exact match
		{"Get Data", true},                 // case-insensitive
		{"execute", true},                  // exact match
		{"run", true},                      // exact match
		{"helper", true},                   // exact match
		{"xyz", true},                      // short, single word (<10 chars)
		{"Read a file from the filesystem", false}, // descriptive
		{"process data", true},             // exact match
		{"Process large data sets efficiently", false}, // long enough
		{"do something", true},             // exact match
		{"utility", true},                  // exact match
		{"perform action", true},           // exact match
		{"handle", true},                   // exact match
		{"two words", false},               // contains space, 9 chars
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := isVagueDescription(tt.desc)
			if got != tt.want {
				t.Errorf("isVagueDescription(%q) = %v, want %v", tt.desc, got, tt.want)
			}
		})
	}
}

func TestIsGenericParamName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"data", true},
		{"value", true},
		{"input", true},
		{"payload", true},
		{"config", true},
		{"Data", true},   // case-insensitive
		{"VALUE", true},  // uppercase
		{"file_path", false},
		{"query", false},
		{"session_id", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGenericParamName(tt.name)
			if got != tt.want {
				t.Errorf("isGenericParamName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestStringSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		min  float64
		max  float64
	}{
		{"hello", "hello", 1.0, 1.0},                     // identical
		{"abc", "xyz", 0.0, 0.1},                          // completely different
		{"read a file", "read a file from disk", 0.5, 1.0}, // partial overlap
		{"", "hello", 0.0, 0.0},                           // empty string
		{"a", "b", 0.0, 0.0},                              // single chars (no bigrams)
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := stringSimilarity(tt.a, tt.b)
			if got < tt.min || got > tt.max {
				t.Errorf("stringSimilarity(%q, %q) = %f, want [%f, %f]", tt.a, tt.b, got, tt.min, tt.max)
			}
		})
	}
}

func TestMakeBigrams(t *testing.T) {
	tests := []struct {
		input string
		want  int // expected number of unique bigrams
	}{
		{"", 0},
		{"a", 0},
		{"ab", 1},
		{"abc", 2},
		{"aaa", 1}, // repeated bigram
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := makeBigrams(tt.input)
			if len(got) != tt.want {
				t.Errorf("makeBigrams(%q) has %d bigrams, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{100, "s"},
	}

	for _, tt := range tests {
		got := plural(tt.n)
		if got != tt.want {
			t.Errorf("plural(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestLintTool_MissingDescription(t *testing.T) {
	tool := mcp.Tool{Name: "my_tool"}
	findings := lintTool(tool)

	found := false
	for _, f := range findings {
		if f.Code == "E101" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected E101 finding for missing description")
	}
}

func TestLintTool_MissingParamType(t *testing.T) {
	tool := mcp.Tool{
		Name:        "my_tool",
		Description: "A useful tool for testing",
		InputSchema: mcp.ToolInputSchema{
			Properties: map[string]interface{}{
				"query": map[string]any{
					"description": "search query",
					// no "type" field
				},
			},
			Required: []string{"query"},
		},
	}
	findings := lintTool(tool)

	found := false
	for _, f := range findings {
		if f.Code == "E102" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected E102 finding for missing parameter type")
	}
}

func TestLintTool_RequiredParamNoDescription(t *testing.T) {
	tool := mcp.Tool{
		Name:        "my_tool",
		Description: "A useful tool for testing",
		InputSchema: mcp.ToolInputSchema{
			Properties: map[string]interface{}{
				"path": map[string]any{
					"type": "string",
					// no description
				},
			},
			Required: []string{"path"},
		},
	}
	findings := lintTool(tool)

	found := false
	for _, f := range findings {
		if f.Code == "E103" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected E103 finding for required param without description")
	}
}

func TestLintTool_NoParamsNoFindings(t *testing.T) {
	tool := mcp.Tool{
		Name:        "list_all",
		Description: "List all items in the database",
		InputSchema: mcp.ToolInputSchema{},
	}
	findings := lintTool(tool)

	// Should only have no property-level findings (may have E101/W101 for description).
	for _, f := range findings {
		if f.Code == "E102" || f.Code == "E103" || f.Code == "W102" || f.Code == "W103" || f.Code == "W104" {
			t.Errorf("unexpected property-level finding %s for tool with no params", f.Code)
		}
	}
}

func TestLintTool_CleanTool(t *testing.T) {
	tool := mcp.Tool{
		Name:        "read_file",
		Description: "Read the contents of a file from the filesystem",
		InputSchema: mcp.ToolInputSchema{
			Properties: map[string]interface{}{
				"path": map[string]any{
					"type":        "string",
					"description": "The absolute path to the file to read",
					"examples":    []string{"/home/user/file.txt"},
				},
			},
			Required: []string{"path"},
		},
	}
	findings := lintTool(tool)

	if len(findings) != 0 {
		data, _ := json.MarshalIndent(findings, "", "  ")
		t.Errorf("expected 0 findings for clean tool, got %d: %s", len(findings), data)
	}
}

func TestLintToolSimilarity_HighSimilarity(t *testing.T) {
	tools := []mcp.Tool{
		{Name: "read_file", Description: "Read a file from disk and return its contents"},
		{Name: "get_file", Description: "Read a file from disk and return its contents"},
	}
	findings := lintToolSimilarity(tools)

	if len(findings) == 0 {
		t.Error("expected W105 finding for identical descriptions")
	}
	for _, f := range findings {
		if f.Code != "W105" {
			t.Errorf("expected code W105, got %s", f.Code)
		}
	}
}

func TestLintToolSimilarity_DifferentDescriptions(t *testing.T) {
	tools := []mcp.Tool{
		{Name: "read_file", Description: "Read a file from the filesystem"},
		{Name: "write_file", Description: "Write content to a database table"},
	}
	findings := lintToolSimilarity(tools)

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for different descriptions, got %d", len(findings))
	}
}

func TestLintSchemaBloat_SmallSchema(t *testing.T) {
	tools := []mcp.Tool{
		{Name: "small_tool", Description: "A small tool"},
	}
	findings := lintSchemaBloat(tools)

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for small schema, got %d", len(findings))
	}
}
