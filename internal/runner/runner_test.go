package runner

import (
	"strings"
	"testing"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// --- substituteFixture tests ---

func TestSubstituteFixture_Strings(t *testing.T) {
	args := map[string]any{
		"file_path": "{{fixture}}/main.go",
		"root_dir":  "{{fixture}}",
		"line":      42, // non-string left alone
	}
	out := substituteFixture(args, "/tmp/test")

	if out["file_path"] != "/tmp/test/main.go" {
		t.Errorf("expected /tmp/test/main.go, got %v", out["file_path"])
	}
	if out["root_dir"] != "/tmp/test" {
		t.Errorf("expected /tmp/test, got %v", out["root_dir"])
	}
	if out["line"] != 42 {
		t.Errorf("expected 42, got %v", out["line"])
	}
}

func TestSubstituteFixture_Arrays(t *testing.T) {
	args := map[string]any{
		"paths": []any{"{{fixture}}/a.txt", "{{fixture}}/b.txt"},
	}
	out := substituteFixture(args, "/data")

	paths, ok := out["paths"].([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", out["paths"])
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	if paths[0] != "/data/a.txt" {
		t.Errorf("expected /data/a.txt, got %v", paths[0])
	}
	if paths[1] != "/data/b.txt" {
		t.Errorf("expected /data/b.txt, got %v", paths[1])
	}
}

func TestSubstituteFixture_NestedMaps(t *testing.T) {
	args := map[string]any{
		"outer": map[string]any{
			"inner": "{{fixture}}/nested.go",
			"count": 3,
		},
	}
	out := substituteFixture(args, "/ws")

	outer, ok := out["outer"].(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", out["outer"])
	}
	if outer["inner"] != "/ws/nested.go" {
		t.Errorf("expected /ws/nested.go, got %v", outer["inner"])
	}
	if outer["count"] != 3 {
		t.Errorf("expected 3, got %v", outer["count"])
	}
}

func TestSubstituteFixture_MixedNesting(t *testing.T) {
	args := map[string]any{
		"edits": []any{
			map[string]any{
				"file_path": "{{fixture}}/main.go",
				"new_text":  "return 42",
			},
			map[string]any{
				"file_path": "{{fixture}}/lib.go",
				"new_text":  "// comment",
			},
		},
	}
	out := substituteFixture(args, "/project")

	edits := out["edits"].([]any)
	edit0 := edits[0].(map[string]any)
	edit1 := edits[1].(map[string]any)

	if edit0["file_path"] != "/project/main.go" {
		t.Errorf("expected /project/main.go, got %v", edit0["file_path"])
	}
	if edit0["new_text"] != "return 42" {
		t.Errorf("new_text should not be modified")
	}
	if edit1["file_path"] != "/project/lib.go" {
		t.Errorf("expected /project/lib.go, got %v", edit1["file_path"])
	}
}

func TestSubstituteFixture_EmptyFixture(t *testing.T) {
	args := map[string]any{"path": "{{fixture}}/test"}
	out := substituteFixture(args, "")

	// Empty fixture means no substitution.
	if out["path"] != "{{fixture}}/test" {
		t.Errorf("expected no substitution, got %v", out["path"])
	}
}

func TestSubstituteFixture_NoPlaceholders(t *testing.T) {
	args := map[string]any{"path": "/absolute/path", "num": 42}
	out := substituteFixture(args, "/fixture")

	if out["path"] != "/absolute/path" {
		t.Errorf("should not modify paths without placeholder")
	}
}

// --- substituteAll with captured variables ---

func TestSubstituteAll_CapturedVariables(t *testing.T) {
	args := map[string]any{
		"session_id": "{{session_id}}",
		"file_path":  "{{fixture}}/main.go",
	}
	captured := map[string]string{"session_id": "abc-123"}
	out := substituteAll(args, "/ws", captured)

	if out["session_id"] != "abc-123" {
		t.Errorf("expected abc-123, got %v", out["session_id"])
	}
	if out["file_path"] != "/ws/main.go" {
		t.Errorf("expected /ws/main.go, got %v", out["file_path"])
	}
}

func TestSubstituteAll_MultipleCaptured(t *testing.T) {
	args := map[string]any{
		"id":   "{{session_id}}",
		"name": "{{entity_name}}",
	}
	captured := map[string]string{
		"session_id":  "sess-1",
		"entity_name": "Alice",
	}
	out := substituteAll(args, "", captured)

	if out["id"] != "sess-1" {
		t.Errorf("expected sess-1, got %v", out["id"])
	}
	if out["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", out["name"])
	}
}

func TestSubstituteAll_CapturedInArrays(t *testing.T) {
	args := map[string]any{
		"ids": []any{"{{session_id}}", "other"},
	}
	captured := map[string]string{"session_id": "xyz"}
	out := substituteAll(args, "", captured)

	ids := out["ids"].([]any)
	if ids[0] != "xyz" {
		t.Errorf("expected xyz in array, got %v", ids[0])
	}
	if ids[1] != "other" {
		t.Errorf("expected other unchanged, got %v", ids[1])
	}
}

func TestSubstituteAll_NilCaptured(t *testing.T) {
	args := map[string]any{"path": "{{fixture}}/test"}
	out := substituteAll(args, "/ws", nil)

	if out["path"] != "/ws/test" {
		t.Errorf("nil captured should not break fixture substitution, got %v", out["path"])
	}
}

// --- extractJSONPath tests ---

func TestExtractJSONPath_SimpleField(t *testing.T) {
	json := `{"session_id": "abc-123", "status": "created"}`

	val, err := extractJSONPath(json, "$.session_id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "abc-123" {
		t.Errorf("expected abc-123, got %q", val)
	}
}

func TestExtractJSONPath_NestedField(t *testing.T) {
	json := `{"result": {"id": "xyz", "count": 42}}`

	val, err := extractJSONPath(json, "$.result.id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "xyz" {
		t.Errorf("expected xyz, got %q", val)
	}
}

func TestExtractJSONPath_NumericField(t *testing.T) {
	json := `{"net_delta": 3}`

	val, err := extractJSONPath(json, "$.net_delta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "3" {
		t.Errorf("expected '3', got %q", val)
	}
}

func TestExtractJSONPath_ArrayIndex(t *testing.T) {
	json := `{"items": [{"name": "first"}, {"name": "second"}]}`

	val, err := extractJSONPath(json, "$.items[1].name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "second" {
		t.Errorf("expected second, got %q", val)
	}
}

func TestExtractJSONPath_MissingField(t *testing.T) {
	json := `{"a": 1}`

	_, err := extractJSONPath(json, "$.missing")
	if err == nil {
		t.Error("expected error for missing field")
	}
}

func TestExtractJSONPath_InvalidJSON(t *testing.T) {
	_, err := extractJSONPath("not json", "$.field")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExtractJSONPath_BooleanField(t *testing.T) {
	json := `{"safe": true}`

	val, err := extractJSONPath(json, "$.safe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "true" {
		t.Errorf("expected 'true', got %q", val)
	}
}

// --- applyServerOverride tests ---

func TestApplyServerOverride_Simple(t *testing.T) {
	a := assertion.Assertion{
		Server: assertion.ServerConfig{
			Command: "original",
			Args:    []string{"--old"},
		},
	}
	applyServerOverride(&a, "new-server arg1 arg2")

	if a.Server.Command != "new-server" {
		t.Errorf("expected command 'new-server', got %q", a.Server.Command)
	}
	if len(a.Server.Args) != 2 || a.Server.Args[0] != "arg1" || a.Server.Args[1] != "arg2" {
		t.Errorf("expected args [arg1 arg2], got %v", a.Server.Args)
	}
}

func TestApplyServerOverride_SingleCommand(t *testing.T) {
	a := assertion.Assertion{}
	applyServerOverride(&a, "my-server")

	if a.Server.Command != "my-server" {
		t.Errorf("expected 'my-server', got %q", a.Server.Command)
	}
	if len(a.Server.Args) != 0 {
		t.Errorf("expected no args, got %v", a.Server.Args)
	}
}

func TestApplyServerOverride_Empty(t *testing.T) {
	a := assertion.Assertion{
		Server: assertion.ServerConfig{Command: "original"},
	}
	applyServerOverride(&a, "")

	// Empty string should not change anything.
	if a.Server.Command != "original" {
		t.Errorf("expected 'original', got %q", a.Server.Command)
	}
}

// --- countFails / countPasses tests ---

func TestCountFails(t *testing.T) {
	results := []assertion.Result{
		{Status: assertion.StatusPass},
		{Status: assertion.StatusFail},
		{Status: assertion.StatusFail},
		{Status: assertion.StatusSkip},
	}
	if n := countFails(results); n != 2 {
		t.Errorf("expected 2 fails, got %d", n)
	}
}

func TestCountPasses(t *testing.T) {
	results := []assertion.Result{
		{Status: assertion.StatusPass},
		{Status: assertion.StatusPass},
		{Status: assertion.StatusFail},
	}
	if n := countPasses(results); n != 2 {
		t.Errorf("expected 2 passes, got %d", n)
	}
}

func TestCountFails_Empty(t *testing.T) {
	if n := countFails(nil); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

// --- extractText tests ---

func TestExtractText_Nil(t *testing.T) {
	// extractText with nil should not panic.
	// It takes *mcp.CallToolResult which we can't easily construct without
	// the mcp package, so we test the edge case indirectly through runAssertion.
	// This test documents the gap.
}

// --- runAssertion error paths ---

func TestRunAssertion_BadServerBinary(t *testing.T) {
	a := assertion.Assertion{
		Name: "bad server",
		Server: assertion.ServerConfig{
			Command: "nonexistent-binary-that-does-not-exist",
			Args:    []string{},
		},
		Assert: assertion.AssertBlock{
			Tool: "test",
			Args: map[string]any{},
		},
	}

	r := runAssertion(a, "", 2*time.Second, "")
	if r.Status != assertion.StatusFail {
		t.Errorf("expected FAIL for bad binary, got %s", r.Status)
	}
	if r.Detail == "" {
		t.Error("expected error detail for bad binary")
	}
}

func TestRunAssertion_Timeout(t *testing.T) {
	// Use a command that hangs (cat with no input).
	a := assertion.Assertion{
		Name: "timeout test",
		Server: assertion.ServerConfig{
			Command: "cat",
			Args:    []string{},
		},
		Assert: assertion.AssertBlock{
			Tool: "test",
			Args: map[string]any{},
		},
	}

	start := time.Now()
	r := runAssertion(a, "", 1*time.Second, "")
	elapsed := time.Since(start)

	if r.Status != assertion.StatusFail {
		t.Errorf("expected FAIL for timeout, got %s", r.Status)
	}
	if elapsed > 5*time.Second {
		t.Errorf("should have timed out in ~1s, took %s", elapsed)
	}
}

// --- Docker flag construction ---

func TestRunAssertion_DockerWrapsCommand(t *testing.T) {
	// We can't actually run Docker in tests, but we can verify
	// the command construction doesn't panic with Docker set.
	a := assertion.Assertion{
		Name: "docker test",
		Server: assertion.ServerConfig{
			Command: "my-server",
			Args:    []string{"--flag"},
			Env:     map[string]string{"KEY": "val"},
		},
		Assert: assertion.AssertBlock{
			Tool: "test",
			Args: map[string]any{},
		},
	}

	// This will fail (docker not running / image doesn't exist)
	// but it should fail gracefully, not panic.
	r := runAssertion(a, "/tmp/fixture", 2*time.Second, "nonexistent-image:latest")
	if r.Status != assertion.StatusFail {
		t.Errorf("expected FAIL for docker, got %s", r.Status)
	}
}

// --- Run/CI command error paths ---

func TestRun_MissingSuite(t *testing.T) {
	err := Run([]string{})
	if err == nil {
		t.Error("expected error for missing --suite")
	}
}

func TestRun_NonexistentSuite(t *testing.T) {
	err := Run([]string{"--suite", "/nonexistent/path"})
	if err == nil {
		t.Error("expected error for nonexistent suite")
	}
}

func TestCI_MissingSuite(t *testing.T) {
	err := CI([]string{})
	if err == nil {
		t.Error("expected error for missing --suite")
	}
}

func TestCI_FailOnRegressionWithoutBaseline(t *testing.T) {
	err := CI([]string{"--suite", "/tmp", "--fail-on-regression"})
	if err == nil {
		t.Error("expected error for --fail-on-regression without --baseline")
	}
}

func TestMatrix_MissingFlags(t *testing.T) {
	err := Matrix([]string{"--suite", "/tmp"})
	if err == nil {
		t.Error("expected error for missing --languages")
	}
}

func TestCoverage_MissingFlags(t *testing.T) {
	err := Coverage([]string{})
	if err == nil {
		t.Error("expected error for missing flags")
	}
}

func TestSnapshot_MissingSuite(t *testing.T) {
	err := Snapshot([]string{})
	if err == nil {
		t.Error("expected error for missing --suite")
	}
}

// --- createMCPClient transport tests ---

func TestCreateMCPClient_DefaultsToStdio(t *testing.T) {
	// Empty transport should attempt stdio with the given command.
	// Using a nonexistent binary to verify it goes through the stdio path.
	server := assertion.ServerConfig{
		Command:   "nonexistent-binary-for-transport-test",
		Transport: "",
	}
	_, err := createMCPClient(server, "", "")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
	// The error should be about exec/command, not about unknown transport or missing URL.
	if strings.Contains(err.Error(), "unknown transport") || strings.Contains(err.Error(), "requires a url") {
		t.Errorf("empty transport should default to stdio, got: %v", err)
	}
}

func TestCreateMCPClient_StdioExplicit(t *testing.T) {
	server := assertion.ServerConfig{
		Command:   "nonexistent-binary-for-transport-test",
		Transport: "stdio",
	}
	_, err := createMCPClient(server, "", "")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
	// Should be a command exec error, not a transport config error.
	if strings.Contains(err.Error(), "unknown transport") || strings.Contains(err.Error(), "requires a url") {
		t.Errorf("stdio transport should produce exec error, got: %v", err)
	}
}

func TestCreateMCPClient_SSERequiresURL(t *testing.T) {
	server := assertion.ServerConfig{
		Transport: "sse",
		URL:       "",
	}
	_, err := createMCPClient(server, "", "")
	if err == nil {
		t.Fatal("expected error for SSE without URL")
	}
	if !strings.Contains(err.Error(), "requires a url") {
		t.Errorf("expected 'requires a url' error, got: %v", err)
	}
}

func TestCreateMCPClient_HTTPRequiresURL(t *testing.T) {
	server := assertion.ServerConfig{
		Transport: "http",
		URL:       "",
	}
	_, err := createMCPClient(server, "", "")
	if err == nil {
		t.Fatal("expected error for HTTP without URL")
	}
	if !strings.Contains(err.Error(), "requires a url") {
		t.Errorf("expected 'requires a url' error, got: %v", err)
	}
}

func TestCreateMCPClient_UnknownTransport(t *testing.T) {
	server := assertion.ServerConfig{
		Transport: "grpc",
	}
	_, err := createMCPClient(server, "", "")
	if err == nil {
		t.Fatal("expected error for unknown transport")
	}
	if !strings.Contains(err.Error(), "unknown transport") {
		t.Errorf("expected 'unknown transport' error, got: %v", err)
	}
}

func TestCreateMCPClient_SSEWithURL(t *testing.T) {
	// SSE with a valid URL should succeed in creating the client
	// (it won't connect until Initialize is called).
	server := assertion.ServerConfig{
		Transport: "sse",
		URL:       "http://localhost:99999/sse",
	}
	c, err := createMCPClient(server, "", "")
	if err != nil {
		t.Fatalf("SSE client creation should succeed with valid URL: %v", err)
	}
	defer c.Close()
}

func TestCreateMCPClient_HTTPWithURL(t *testing.T) {
	server := assertion.ServerConfig{
		Transport: "http",
		URL:       "http://localhost:99999/mcp",
	}
	c, err := createMCPClient(server, "", "")
	if err != nil {
		t.Fatalf("HTTP client creation should succeed with valid URL: %v", err)
	}
	defer c.Close()
}

func TestCreateMCPClient_TransportCaseInsensitive(t *testing.T) {
	server := assertion.ServerConfig{
		Transport: "SSE",
		URL:       "http://localhost:99999/sse",
	}
	c, err := createMCPClient(server, "", "")
	if err != nil {
		t.Fatalf("transport should be case-insensitive: %v", err)
	}
	defer c.Close()
}

func TestCreateMCPClient_DockerIgnoredForHTTP(t *testing.T) {
	// Docker flag is only used for stdio; HTTP/SSE should ignore it.
	server := assertion.ServerConfig{
		Transport: "sse",
		URL:       "http://localhost:99999/sse",
	}
	c, err := createMCPClient(server, "", "some-docker-image")
	if err != nil {
		t.Fatalf("HTTP transport should succeed even with docker image set: %v", err)
	}
	defer c.Close()
}

func TestRunAssertion_SSEWithoutURL(t *testing.T) {
	a := assertion.Assertion{
		Name: "sse no url",
		Server: assertion.ServerConfig{
			Transport: "sse",
		},
		Assert: assertion.AssertBlock{
			Tool: "test",
			Args: map[string]any{},
		},
	}
	r := runAssertion(a, "", 2*time.Second, "")
	if r.Status != assertion.StatusFail {
		t.Errorf("expected FAIL, got %s", r.Status)
	}
	if !strings.Contains(r.Detail, "requires a url") {
		t.Errorf("expected url error in detail, got: %s", r.Detail)
	}
}

func TestRunAssertion_UnknownTransport(t *testing.T) {
	a := assertion.Assertion{
		Name: "bad transport",
		Server: assertion.ServerConfig{
			Transport: "websocket",
		},
		Assert: assertion.AssertBlock{
			Tool: "test",
			Args: map[string]any{},
		},
	}
	r := runAssertion(a, "", 2*time.Second, "")
	if r.Status != assertion.StatusFail {
		t.Errorf("expected FAIL, got %s", r.Status)
	}
	if !strings.Contains(r.Detail, "unknown transport") {
		t.Errorf("expected unknown transport error, got: %s", r.Detail)
	}
}
