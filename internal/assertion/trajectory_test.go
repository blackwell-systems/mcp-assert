package assertion

import (
	"os"
	"path/filepath"
	"testing"
)

func trace(tools ...string) []TraceEntry {
	entries := make([]TraceEntry, len(tools))
	for i, t := range tools {
		entries[i] = TraceEntry{Tool: t}
	}
	return entries
}

func traceWithArgs(tool string, args map[string]any) TraceEntry {
	return TraceEntry{Tool: tool, Args: args}
}

// --- order tests ---

func TestCheckTrajectory_Order_Pass(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "order", Tools: []string{"prepare_rename", "rename_symbol", "get_diagnostics"}}},
		trace("start_lsp", "open_document", "prepare_rename", "rename_symbol", "get_diagnostics"),
	)
	if err != nil {
		t.Fatalf("expected pass: %v", err)
	}
}

func TestCheckTrajectory_Order_NonAdjacent(t *testing.T) {
	// Tools don't need to be adjacent.
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "order", Tools: []string{"prepare_rename", "rename_symbol"}}},
		trace("prepare_rename", "get_diagnostics", "rename_symbol"),
	)
	if err != nil {
		t.Fatalf("non-adjacent order should pass: %v", err)
	}
}

func TestCheckTrajectory_Order_WrongOrder_Fail(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "order", Tools: []string{"prepare_rename", "rename_symbol"}}},
		trace("rename_symbol", "prepare_rename"), // reversed
	)
	if err == nil {
		t.Fatal("expected failure for wrong order")
	}
}

func TestCheckTrajectory_Order_Missing_Fail(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "order", Tools: []string{"prepare_rename", "rename_symbol"}}},
		trace("start_lsp", "get_diagnostics"),
	)
	if err == nil {
		t.Fatal("expected failure for missing tool")
	}
}

func TestCheckTrajectory_Order_Empty_Pass(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "order", Tools: []string{}}},
		trace("any_tool"),
	)
	if err != nil {
		t.Fatalf("empty tools list should pass: %v", err)
	}
}

// --- presence tests ---

func TestCheckTrajectory_Presence_Pass(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "presence", Tools: []string{"get_references", "rename_symbol"}}},
		trace("get_references", "rename_symbol", "get_diagnostics"),
	)
	if err != nil {
		t.Fatalf("expected pass: %v", err)
	}
}

func TestCheckTrajectory_Presence_Missing_Fail(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "presence", Tools: []string{"get_references", "missing_tool"}}},
		trace("get_references", "get_diagnostics"),
	)
	if err == nil {
		t.Fatal("expected failure for missing tool")
	}
}

// --- absence tests ---

func TestCheckTrajectory_Absence_Pass(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "absence", Tools: []string{"apply_edit"}}},
		trace("get_references", "rename_symbol", "get_diagnostics"),
	)
	if err != nil {
		t.Fatalf("expected pass: %v", err)
	}
}

func TestCheckTrajectory_Absence_Present_Fail(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "absence", Tools: []string{"apply_edit"}}},
		trace("get_references", "apply_edit", "get_diagnostics"),
	)
	if err == nil {
		t.Fatal("expected failure: apply_edit should not be in trace")
	}
}

// --- args_contain tests ---

func TestCheckTrajectory_ArgsContain_Pass(t *testing.T) {
	tr := []TraceEntry{
		traceWithArgs("rename_symbol", map[string]any{"new_name": "Entity", "line": float64(6)}),
	}
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "args_contain", Tool: "rename_symbol", Args: map[string]any{"new_name": "Entity"}}},
		tr,
	)
	if err != nil {
		t.Fatalf("expected pass: %v", err)
	}
}

func TestCheckTrajectory_ArgsContain_WrongValue_Fail(t *testing.T) {
	tr := []TraceEntry{
		traceWithArgs("rename_symbol", map[string]any{"new_name": "WrongName"}),
	}
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "args_contain", Tool: "rename_symbol", Args: map[string]any{"new_name": "Entity"}}},
		tr,
	)
	if err == nil {
		t.Fatal("expected failure for wrong value")
	}
}

func TestCheckTrajectory_ArgsContain_MissingKey_Fail(t *testing.T) {
	tr := []TraceEntry{
		traceWithArgs("rename_symbol", map[string]any{"line": float64(6)}),
	}
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "args_contain", Tool: "rename_symbol", Args: map[string]any{"new_name": "Entity"}}},
		tr,
	)
	if err == nil {
		t.Fatal("expected failure for missing key")
	}
}

func TestCheckTrajectory_ArgsContain_ToolNotFound_Fail(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "args_contain", Tool: "rename_symbol", Args: map[string]any{"new_name": "Entity"}}},
		trace("get_references", "get_diagnostics"),
	)
	if err == nil {
		t.Fatal("expected failure: rename_symbol not in trace")
	}
}

// --- unknown type ---

func TestCheckTrajectory_UnknownType_Fail(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "invalid_type"}},
		trace("any_tool"),
	)
	if err == nil {
		t.Fatal("expected failure for unknown assertion type")
	}
}

// --- empty trace ---

func TestCheckTrajectory_EmptyTrace(t *testing.T) {
	err := CheckTrajectory(
		[]TrajectoryAssertion{{Type: "presence", Tools: []string{"prepare_rename"}}},
		[]TraceEntry{},
	)
	if err == nil {
		t.Fatal("expected failure: tool not in empty trace")
	}
}

func TestCheckTrajectory_EmptyTrajectory(t *testing.T) {
	// No trajectory checks = always pass.
	err := CheckTrajectory(nil, trace("any_tool"))
	if err != nil {
		t.Fatalf("empty trajectory should pass: %v", err)
	}
}

// --- multiple trajectory assertions ---

func TestCheckTrajectory_Multiple_AllPass(t *testing.T) {
	tr := []TraceEntry{
		traceWithArgs("prepare_rename", map[string]any{"line": float64(6)}),
		{Tool: "rename_symbol"},
		{Tool: "get_diagnostics"},
	}
	err := CheckTrajectory(
		[]TrajectoryAssertion{
			{Type: "order", Tools: []string{"prepare_rename", "rename_symbol", "get_diagnostics"}},
			{Type: "presence", Tools: []string{"prepare_rename", "rename_symbol"}},
			{Type: "absence", Tools: []string{"apply_edit"}},
			{Type: "args_contain", Tool: "prepare_rename", Args: map[string]any{"line": float64(6)}},
		},
		tr,
	)
	if err != nil {
		t.Fatalf("all checks should pass: %v", err)
	}
}

// --- audit log loading ---

func TestLoadAuditLog_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	content := `{"tool":"prepare_rename","args":{"line":6,"column":6}}
{"tool":"rename_symbol","args":{"new_name":"Entity"}}
{"tool":"get_diagnostics","args":{"file_path":"main.go"}}
`
	os.WriteFile(path, []byte(content), 0644)

	entries, err := LoadAuditLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Tool != "prepare_rename" {
		t.Errorf("expected prepare_rename, got %q", entries[0].Tool)
	}
	if entries[1].Tool != "rename_symbol" {
		t.Errorf("expected rename_symbol, got %q", entries[1].Tool)
	}
}

func TestLoadAuditLog_AgentLspFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	// agent-lsp format includes action field.
	content := `{"timestamp":"2026-04-23T00:00:00Z","action":"tool_call","tool":"get_references","args":{"file_path":"main.go","line":6}}
{"timestamp":"2026-04-23T00:00:01Z","action":"tool_call","tool":"rename_symbol","args":{"new_name":"Entity"}}
`
	os.WriteFile(path, []byte(content), 0644)

	entries, err := LoadAuditLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Tool != "get_references" {
		t.Errorf("expected get_references, got %q", entries[0].Tool)
	}
}

func TestLoadAuditLog_SkipsNonToolEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	content := `{"action":"session_start","session_id":"abc"}
{"tool":"prepare_rename","args":{}}
{"action":"session_end"}
`
	os.WriteFile(path, []byte(content), 0644)

	entries, err := LoadAuditLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (non-tool entries skipped), got %d", len(entries))
	}
}

func TestLoadAuditLog_Missing(t *testing.T) {
	_, err := LoadAuditLog("/nonexistent/path.jsonl")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
