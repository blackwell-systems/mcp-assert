package assertion

import (
	"os"
	"path/filepath"
	"testing"
)

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
func strPtr(s string) *string { return &s }

func TestCheck_NotError(t *testing.T) {
	// Should pass when no error.
	err := Check(Expect{NotError: boolPtr(true)}, "ok", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fail when error.
	err = Check(Expect{NotError: boolPtr(true)}, "fail", true)
	if err == nil {
		t.Fatal("expected error for isError=true")
	}
}

func TestCheck_IsError(t *testing.T) {
	err := Check(Expect{IsError: boolPtr(true)}, "err", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{IsError: boolPtr(true)}, "ok", false)
	if err == nil {
		t.Fatal("expected error for isError=false")
	}
}

func TestCheck_NotEmpty(t *testing.T) {
	for _, empty := range []string{"", " ", "null", "[]", "{}"} {
		err := Check(Expect{NotEmpty: boolPtr(true)}, empty, false)
		if err == nil {
			t.Errorf("expected error for %q", empty)
		}
	}
	err := Check(Expect{NotEmpty: boolPtr(true)}, `{"key": "value"}`, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheck_Equals(t *testing.T) {
	err := Check(Expect{Equals: strPtr("hello")}, "hello", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{Equals: strPtr("hello")}, "world", false)
	if err == nil {
		t.Fatal("expected error for mismatch")
	}
	// Trims whitespace.
	err = Check(Expect{Equals: strPtr("hello")}, " hello ", false)
	if err != nil {
		t.Fatalf("unexpected error with whitespace: %v", err)
	}
}

func TestCheck_Contains(t *testing.T) {
	err := Check(Expect{Contains: []string{"foo", "bar"}}, "foo and bar here", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{Contains: []string{"missing"}}, "foo and bar", false)
	if err == nil {
		t.Fatal("expected error for missing substring")
	}
}

func TestCheck_NotContains(t *testing.T) {
	err := Check(Expect{NotContains: []string{"secret"}}, "public data", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{NotContains: []string{"secret"}}, "has secret inside", false)
	if err == nil {
		t.Fatal("expected error for present substring")
	}
}

func TestCheck_ContainsAny(t *testing.T) {
	// Should pass when response contains one of the values.
	err := Check(Expect{ContainsAny: []string{"foo", "bar"}}, "only bar here", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fail when response contains none of the values.
	err = Check(Expect{ContainsAny: []string{"foo", "bar"}}, "nothing relevant", false)
	if err == nil {
		t.Fatal("expected error when none of the values are present")
	}
	// Single value (equivalent to contains).
	err = Check(Expect{ContainsAny: []string{"hello"}}, "hello world", false)
	if err != nil {
		t.Fatalf("unexpected error for single value: %v", err)
	}
}

func TestCheck_MatchesRegex(t *testing.T) {
	err := Check(Expect{MatchesRegex: []string{`\d+ items`}}, "found 42 items", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{MatchesRegex: []string{`^\d+$`}}, "not a number", false)
	if err == nil {
		t.Fatal("expected error for non-matching regex")
	}
	// Invalid regex should error.
	err = Check(Expect{MatchesRegex: []string{`[invalid`}}, "text", false)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestCheck_JSONPath(t *testing.T) {
	json := `{"name": "Alice", "address": {"city": "NYC"}, "items": [{"id": 1}, {"id": 2}]}`

	// Simple field.
	err := Check(Expect{JSONPath: map[string]any{"$.name": "Alice"}}, json, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nested field.
	err = Check(Expect{JSONPath: map[string]any{"$.address.city": "NYC"}}, json, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Array index.
	err = Check(Expect{JSONPath: map[string]any{"$.items[0].id": float64(1)}}, json, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wrong value.
	err = Check(Expect{JSONPath: map[string]any{"$.name": "Bob"}}, json, false)
	if err == nil {
		t.Fatal("expected error for wrong value")
	}

	// Missing field.
	err = Check(Expect{JSONPath: map[string]any{"$.missing": "x"}}, json, false)
	if err == nil {
		t.Fatal("expected error for missing field")
	}

	// Not JSON.
	err = Check(Expect{JSONPath: map[string]any{"$.x": "y"}}, "not json", false)
	if err == nil {
		t.Fatal("expected error for non-JSON input")
	}
}

func TestCheck_MinMaxResults(t *testing.T) {
	arr := `[1, 2, 3]`

	err := Check(Expect{MinResults: intPtr(2)}, arr, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{MinResults: intPtr(5)}, arr, false)
	if err == nil {
		t.Fatal("expected error for too few results")
	}
	err = Check(Expect{MaxResults: intPtr(5)}, arr, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{MaxResults: intPtr(2)}, arr, false)
	if err == nil {
		t.Fatal("expected error for too many results")
	}

	// Object with array field.
	obj := `{"locations": [{"file": "a.go"}, {"file": "b.go"}]}`
	err = Check(Expect{MinResults: intPtr(2)}, obj, false)
	if err != nil {
		t.Fatalf("unexpected error for object array: %v", err)
	}
}

func TestCheck_NetDelta(t *testing.T) {
	json := `{"net_delta": 0, "safe": true}`
	err := Check(Expect{NetDelta: intPtr(0)}, json, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{NetDelta: intPtr(0)}, `{"net_delta": 3}`, false)
	if err == nil {
		t.Fatal("expected error for net_delta mismatch")
	}
	err = Check(Expect{NetDelta: intPtr(0)}, `not json`, false)
	if err == nil {
		t.Fatal("expected error for non-JSON")
	}
}

func TestCheck_FileContains(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0644)

	err := Check(Expect{FileContains: map[string]string{path: "world"}}, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Check(Expect{FileContains: map[string]string{path: "missing"}}, "", false)
	if err == nil {
		t.Fatal("expected error for missing content")
	}
	err = Check(Expect{FileContains: map[string]string{"/nonexistent": "x"}}, "", false)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCheck_FileNotContains(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0644)

	// Should pass when file does not contain the string.
	err := Check(Expect{FileNotContains: map[string]string{path: "missing"}}, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fail when file does contain the string.
	err = Check(Expect{FileNotContains: map[string]string{path: "world"}}, "", false)
	if err == nil {
		t.Fatal("expected error when file contains the unwanted string")
	}
	// Should fail for nonexistent file.
	err = Check(Expect{FileNotContains: map[string]string{"/nonexistent": "x"}}, "", false)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCheck_FileNotExists_Pass(t *testing.T) {
	// Path that does not exist: should pass.
	err := Check(Expect{FileNotExists: []string{"/nonexistent/path/should/not/exist.txt"}}, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheck_FileNotExists_Fail(t *testing.T) {
	// Create a temp file, then assert it should not exist: should fail.
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")
	os.WriteFile(path, []byte("data"), 0644)

	err := Check(Expect{FileNotExists: []string{path}}, "", false)
	if err == nil {
		t.Fatal("expected error when file exists")
	}
}

func TestCheck_InOrder(t *testing.T) {
	text := "first then second then third"

	err := Check(Expect{InOrder: []string{"first", "second", "third"}}, text, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Wrong order.
	err = Check(Expect{InOrder: []string{"third", "first"}}, text, false)
	if err == nil {
		t.Fatal("expected error for wrong order")
	}
	// Missing item.
	err = Check(Expect{InOrder: []string{"first", "missing"}}, text, false)
	if err == nil {
		t.Fatal("expected error for missing item")
	}
}

func TestCheckWithSnapshots_FileUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	os.WriteFile(path, []byte("original"), 0644)

	snapshots := map[string]string{path: "original"}

	// File unchanged — should pass.
	err := CheckWithSnapshots(Expect{FileUnchanged: []string{path}}, "ok", false, snapshots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify file — should fail.
	os.WriteFile(path, []byte("modified"), 0644)
	err = CheckWithSnapshots(Expect{FileUnchanged: []string{path}}, "ok", false, snapshots)
	if err == nil {
		t.Fatal("expected error for modified file")
	}

	// Missing snapshot — should fail.
	err = CheckWithSnapshots(Expect{FileUnchanged: []string{"/no/snapshot"}}, "ok", false, snapshots)
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

func TestCheck_Combined(t *testing.T) {
	// Multiple assertions at once.
	json := `{"name": "Alice", "count": 3}`
	err := Check(Expect{
		NotError:     boolPtr(true),
		NotEmpty:     boolPtr(true),
		Contains:     []string{"Alice"},
		NotContains:  []string{"Bob"},
		MatchesRegex: []string{`"count":\s*\d+`},
		JSONPath:     map[string]any{"$.name": "Alice"},
	}, json, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheck_CompletionJSON(t *testing.T) {
	// Simulate a marshaled CompleteResult with completion values.
	completionJSON := `{"completion":{"values":["formal","friendly","fun"],"total":3}}`

	// contains assertion.
	err := Check(Expect{Contains: []string{"formal"}}, completionJSON, false)
	if err != nil {
		t.Fatalf("expected contains 'formal' to pass, got: %v", err)
	}

	// json_path assertion.
	err = Check(Expect{JSONPath: map[string]any{"$.completion.total": float64(3)}}, completionJSON, false)
	if err != nil {
		t.Fatalf("expected json_path total=3 to pass, got: %v", err)
	}

	// json_path on values array.
	err = Check(Expect{JSONPath: map[string]any{"$.completion.values[0]": "formal"}}, completionJSON, false)
	if err != nil {
		t.Fatalf("expected json_path values[0]=formal to pass, got: %v", err)
	}

	// not_empty assertion.
	err = Check(Expect{NotEmpty: boolPtr(true)}, completionJSON, false)
	if err != nil {
		t.Fatalf("expected not_empty to pass, got: %v", err)
	}
}

func TestCheck_EmptyCompletion(t *testing.T) {
	// Empty completion result (no values).
	emptyJSON := `{"completion":{"values":[],"total":0}}`

	// not_error should pass (it's a valid result, just empty).
	err := Check(Expect{NotError: boolPtr(true)}, emptyJSON, false)
	if err != nil {
		t.Fatalf("expected not_error to pass, got: %v", err)
	}

	// contains should fail for a value that isn't present.
	err = Check(Expect{Contains: []string{"formal"}}, emptyJSON, false)
	if err == nil {
		t.Fatal("expected contains 'formal' to fail on empty completion")
	}

	// json_path total should be 0.
	err = Check(Expect{JSONPath: map[string]any{"$.completion.total": float64(0)}}, emptyJSON, false)
	if err != nil {
		t.Fatalf("expected json_path total=0 to pass, got: %v", err)
	}
}

func TestCheckProgress_PassesWhenCountMeetsMinimum(t *testing.T) {
	err := CheckProgress(Expect{MinProgress: intPtr(3)}, 3)
	if err != nil {
		t.Fatalf("expected pass when count == min_progress, got: %v", err)
	}
}

func TestCheckProgress_PassesWhenCountExceedsMinimum(t *testing.T) {
	err := CheckProgress(Expect{MinProgress: intPtr(2)}, 5)
	if err != nil {
		t.Fatalf("expected pass when count > min_progress, got: %v", err)
	}
}

func TestCheckProgress_FailsWhenCountBelowMinimum(t *testing.T) {
	err := CheckProgress(Expect{MinProgress: intPtr(3)}, 1)
	if err == nil {
		t.Fatal("expected error when count < min_progress")
	}
}

func TestCheckProgress_NoMinProgress_AlwaysPasses(t *testing.T) {
	// min_progress not set: no check, always passes regardless of count.
	err := CheckProgress(Expect{}, 0)
	if err != nil {
		t.Fatalf("expected pass when min_progress not set, got: %v", err)
	}
}
