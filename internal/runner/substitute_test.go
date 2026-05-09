package runner

import (
	"testing"
)

func TestSubstituteMapKeys(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]string
		fixture string
		wantKey string
	}{
		{
			name:    "substitutes fixture in key",
			m:       map[string]string{"{{fixture}}/output.txt": "expected"},
			fixture: "/tmp/test",
			wantKey: "/tmp/test/output.txt",
		},
		{
			name:    "empty fixture returns original",
			m:       map[string]string{"{{fixture}}/output.txt": "expected"},
			fixture: "",
			wantKey: "{{fixture}}/output.txt",
		},
		{
			name:    "nil map returns nil",
			m:       nil,
			fixture: "/tmp",
			wantKey: "",
		},
		{
			name:    "no placeholder unchanged",
			m:       map[string]string{"/absolute/path": "content"},
			fixture: "/tmp",
			wantKey: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteMapKeys(tt.m, tt.fixture)
			if tt.m == nil {
				if result != nil && len(result) != 0 {
					t.Errorf("expected nil/empty map for nil input, got %v", result)
				}
				return
			}
			if tt.wantKey == "" {
				return
			}
			if _, ok := result[tt.wantKey]; !ok {
				t.Errorf("expected key %q in result, got keys: %v", tt.wantKey, result)
			}
		})
	}
}

func TestSubstituteSlice(t *testing.T) {
	tests := []struct {
		name    string
		s       []string
		fixture string
		want    []string
	}{
		{
			name:    "substitutes fixture",
			s:       []string{"{{fixture}}/a.txt", "{{fixture}}/b.txt"},
			fixture: "/data",
			want:    []string{"/data/a.txt", "/data/b.txt"},
		},
		{
			name:    "empty fixture returns original",
			s:       []string{"{{fixture}}/a.txt"},
			fixture: "",
			want:    []string{"{{fixture}}/a.txt"},
		},
		{
			name:    "nil slice returns nil",
			s:       nil,
			fixture: "/tmp",
			want:    nil,
		},
		{
			name:    "no placeholder unchanged",
			s:       []string{"/absolute/path"},
			fixture: "/tmp",
			want:    []string{"/absolute/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteSlice(tt.s, tt.fixture)
			if tt.want == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(result), len(tt.want))
			}
			for i, v := range result {
				if v != tt.want[i] {
					t.Errorf("[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestExtractJSONPath_RootDocument(t *testing.T) {
	input := `{"key": "value"}`
	val, err := extractJSONPath(input, "$")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != input {
		t.Errorf("expected full document, got %q", val)
	}
}

func TestExtractJSONPath_EmptyPath(t *testing.T) {
	input := `{"key": "value"}`
	val, err := extractJSONPath(input, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != input {
		t.Errorf("expected full document for empty path, got %q", val)
	}
}

func TestExtractJSONPath_FloatField(t *testing.T) {
	input := `{"pi": 3.14}`
	val, err := extractJSONPath(input, "$.pi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "3.14" {
		t.Errorf("expected 3.14, got %q", val)
	}
}

func TestExtractJSONPath_ObjectField(t *testing.T) {
	input := `{"nested": {"a": 1, "b": 2}}`
	val, err := extractJSONPath(input, "$.nested")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be valid JSON
	if val == "" {
		t.Error("expected non-empty JSON for object field")
	}
}

func TestExtractJSONPath_ArrayOutOfBounds(t *testing.T) {
	input := `{"items": [1, 2, 3]}`
	_, err := extractJSONPath(input, "$.items[5]")
	if err == nil {
		t.Error("expected error for out-of-bounds index")
	}
}
