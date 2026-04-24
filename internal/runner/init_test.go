package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitTemplate(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "evals")

	err := Init([]string{outDir})
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Check that the template was created.
	assertPath := filepath.Join(outDir, "read_file.yaml")
	if _, err := os.Stat(assertPath); os.IsNotExist(err) {
		t.Error("expected read_file.yaml to be created")
	}

	// Check that the fixture was created.
	helloPath := filepath.Join(outDir, "fixtures", "hello.txt")
	if _, err := os.Stat(helloPath); os.IsNotExist(err) {
		t.Error("expected fixtures/hello.txt to be created")
	}

	data, err := os.ReadFile(helloPath)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	if string(data) != fixtureContent {
		t.Errorf("fixture content = %q, want %q", string(data), fixtureContent)
	}
}

func TestInitTemplate_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "evals")

	// First init succeeds.
	if err := Init([]string{outDir}); err != nil {
		t.Fatalf("first Init() returned error: %v", err)
	}

	// Second init should fail (not overwrite).
	err := Init([]string{outDir})
	if err == nil {
		t.Fatal("expected error on second Init(), got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestInitTemplate_DefaultDir(t *testing.T) {
	// When no args, uses "evals" directory. We test that the function
	// accepts empty args without panicking. We run it in a temp dir
	// to avoid polluting the working directory.
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	err := Init([]string{})
	if err != nil {
		t.Fatalf("Init() with empty args returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "evals", "read_file.yaml")); os.IsNotExist(err) {
		t.Error("expected evals/read_file.yaml to be created in default dir")
	}
}

func TestInitWithServer_InvalidServer(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "evals")

	err := Init([]string{"--server", "nonexistent-binary-that-does-not-exist-xyz", outDir})
	if err == nil {
		t.Fatal("expected error for invalid server, got nil")
	}
	if !strings.Contains(err.Error(), "generate:") {
		t.Errorf("expected error to mention generate, got: %v", err)
	}
}

func TestInitWithServer_EmptyServer(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "evals")

	err := Init([]string{"--server", "", outDir})
	if err != nil {
		// Empty server string should fall through to template mode
		t.Fatalf("Init() with empty --server should fall back to template: %v", err)
	}
}

func TestInitWithServer_FlagParsing(t *testing.T) {
	// Verify that --server and --fixture flags are parsed correctly
	// without actually connecting to a server. We use a nonexistent
	// binary to confirm flags are reaching the right code path.
	dir := t.TempDir()
	outDir := filepath.Join(dir, "evals")

	err := Init([]string{
		"--server", "fake-server --arg1",
		"--fixture", "/tmp/fixtures",
		outDir,
	})
	if err == nil {
		t.Fatal("expected error for fake server, got nil")
	}
	// The error should come from the generate step (trying to start the server),
	// not from flag parsing.
	if !strings.Contains(err.Error(), "generate:") {
		t.Errorf("expected generate error, got: %v", err)
	}
}
