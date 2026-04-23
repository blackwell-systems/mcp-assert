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
