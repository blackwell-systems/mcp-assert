package assertion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSuite_BasicYAML(t *testing.T) {
	dir := t.TempDir()
	yaml := `name: test hover
server:
  command: agent-lsp
  args: ["go:gopls"]
assert:
  tool: get_info_on_location
  args:
    file_path: main.go
    line: 1
    column: 1
  expect:
    not_error: true
`
	os.WriteFile(filepath.Join(dir, "hover.yaml"), []byte(yaml), 0644)

	suite, err := LoadSuite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suite.Assertions) != 1 {
		t.Fatalf("expected 1 assertion, got %d", len(suite.Assertions))
	}
	if suite.Assertions[0].Name != "test hover" {
		t.Errorf("expected name 'test hover', got %q", suite.Assertions[0].Name)
	}
	if suite.Assertions[0].Server.Command != "agent-lsp" {
		t.Errorf("expected command 'agent-lsp', got %q", suite.Assertions[0].Server.Command)
	}
	if suite.Assertions[0].Assert.Tool != "get_info_on_location" {
		t.Errorf("expected tool 'get_info_on_location', got %q", suite.Assertions[0].Assert.Tool)
	}
}

func TestLoadSuite_Subdirectory(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "go")
	os.MkdirAll(subDir, 0755)

	yaml := `name: sub test
server:
  command: test-server
assert:
  tool: ping
  args: {}
  expect:
    not_empty: true
`
	os.WriteFile(filepath.Join(subDir, "ping.yaml"), []byte(yaml), 0644)

	suite, err := LoadSuite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suite.Assertions) != 1 {
		t.Fatalf("expected 1 assertion from subdirectory, got %d", len(suite.Assertions))
	}
	if suite.Assertions[0].Name != "sub test" {
		t.Errorf("expected name 'sub test', got %q", suite.Assertions[0].Name)
	}
}

func TestLoadSuite_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadSuite(dir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
}

func TestLoadSuite_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	// Tabs in YAML values are invalid; indentation mix triggers parse errors.
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("name: test\nserver:\n\t- broken:\n  mixed: indent"), 0644)

	_, err := LoadSuite(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadSuite_NonexistentDir(t *testing.T) {
	_, err := LoadSuite("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestLoadSuite_DefaultName(t *testing.T) {
	dir := t.TempDir()
	// YAML without a name field — should default to filename.
	yaml := `server:
  command: test
assert:
  tool: ping
  args: {}
  expect:
    not_empty: true
`
	os.WriteFile(filepath.Join(dir, "unnamed.yaml"), []byte(yaml), 0644)

	suite, err := LoadSuite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if suite.Assertions[0].Name != "unnamed.yaml" {
		t.Errorf("expected default name 'unnamed.yaml', got %q", suite.Assertions[0].Name)
	}
}

func TestLoadSuite_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.yaml", "b.yml", "c.yaml"} {
		yaml := `name: ` + name + `
server:
  command: test
assert:
  tool: ping
  args: {}
  expect:
    not_empty: true
`
		os.WriteFile(filepath.Join(dir, name), []byte(yaml), 0644)
	}
	// Non-YAML file should be ignored.
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# ignore"), 0644)

	suite, err := LoadSuite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suite.Assertions) != 3 {
		t.Fatalf("expected 3 assertions, got %d", len(suite.Assertions))
	}
}

func TestLoadSuite_SingleFile(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `name: single file test
server:
  command: test-server
assert:
  tool: ping
  args: {}
  expect:
    not_empty: true
`
	filePath := filepath.Join(dir, "single.yaml")
	os.WriteFile(filePath, []byte(yamlContent), 0644)

	suite, err := LoadSuite(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suite.Assertions) != 1 {
		t.Fatalf("expected 1 assertion, got %d", len(suite.Assertions))
	}
	if suite.Assertions[0].Name != "single file test" {
		t.Errorf("expected name 'single file test', got %q", suite.Assertions[0].Name)
	}
}

func TestLoadSuite_SingleFile_NonYAML(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notes.txt")
	os.WriteFile(filePath, []byte("not yaml"), 0644)

	_, err := LoadSuite(filePath)
	if err == nil {
		t.Fatal("expected error for non-YAML file")
	}
	if !strings.Contains(err.Error(), "not a YAML file") {
		t.Errorf("expected error containing 'not a YAML file', got %q", err.Error())
	}
}

func TestLoadSuite_SingleFile_SetsDir(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `name: dir check
server:
  command: test-server
assert:
  tool: ping
  args: {}
  expect:
    not_empty: true
`
	filePath := filepath.Join(dir, "check.yaml")
	os.WriteFile(filePath, []byte(yamlContent), 0644)

	suite, err := LoadSuite(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if suite.Dir != dir {
		t.Errorf("expected Dir to be %q (parent directory), got %q", dir, suite.Dir)
	}
}

func TestLoadSuite_SingleFile_YmlExtension(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `name: yml extension test
server:
  command: test-server
assert:
  tool: ping
  args: {}
  expect:
    not_empty: true
`
	filePath := filepath.Join(dir, "test.yml")
	os.WriteFile(filePath, []byte(yamlContent), 0644)

	suite, err := LoadSuite(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suite.Assertions) != 1 {
		t.Fatalf("expected 1 assertion, got %d", len(suite.Assertions))
	}
	if suite.Assertions[0].Name != "yml extension test" {
		t.Errorf("expected name 'yml extension test', got %q", suite.Assertions[0].Name)
	}
}

func TestLoadSuite_SetupSteps(t *testing.T) {
	dir := t.TempDir()
	yaml := `name: with setup
server:
  command: agent-lsp
  args: ["go:gopls"]
  env:
    FOO: bar
setup:
  - tool: start_lsp
    args:
      root_dir: /tmp
  - tool: open_document
    args:
      file_path: /tmp/main.go
assert:
  tool: get_info_on_location
  args:
    file_path: /tmp/main.go
    line: 1
    column: 1
  expect:
    not_error: true
    contains: ["func"]
`
	os.WriteFile(filepath.Join(dir, "setup.yaml"), []byte(yaml), 0644)

	suite, err := LoadSuite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := suite.Assertions[0]
	if len(a.Setup) != 2 {
		t.Fatalf("expected 2 setup steps, got %d", len(a.Setup))
	}
	if a.Server.Env["FOO"] != "bar" {
		t.Errorf("expected env FOO=bar, got %q", a.Server.Env["FOO"])
	}
}
