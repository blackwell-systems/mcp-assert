package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir(t *testing.T) {
	// Create a source directory with files and subdirectories.
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "root.txt"), []byte("root content"), 0644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(src, "subdir")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "nested.txt"), []byte("nested content"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "copy")
	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	// Verify root file.
	data, err := os.ReadFile(filepath.Join(dst, "root.txt"))
	if err != nil {
		t.Fatalf("read root.txt: %v", err)
	}
	if string(data) != "root content" {
		t.Errorf("root.txt = %q, want %q", data, "root content")
	}

	// Verify nested file.
	data, err = os.ReadFile(filepath.Join(dst, "subdir", "nested.txt"))
	if err != nil {
		t.Fatalf("read nested.txt: %v", err)
	}
	if string(data) != "nested content" {
		t.Errorf("nested.txt = %q, want %q", data, "nested content")
	}
}

func TestCopyDir_Empty(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "empty-copy")

	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir empty: %v", err)
	}

	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dir, got %d entries", len(entries))
	}
}

func TestCopyDir_Permissions(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "exec.sh"), []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "readonly.txt"), []byte("read only"), 0444); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "perm-copy")
	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	info, err := os.Stat(filepath.Join(dst, "exec.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("exec.sh perm = %o, want 0755", info.Mode().Perm())
	}

	info, err = os.Stat(filepath.Join(dst, "readonly.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0444 {
		t.Errorf("readonly.txt perm = %o, want 0444", info.Mode().Perm())
	}
}

func TestFixtureIsolation(t *testing.T) {
	// Create a fixture directory with a file.
	fixture := t.TempDir()
	original := filepath.Join(fixture, "data.txt")
	if err := os.WriteFile(original, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	isoFixture, cleanup := isolateFixture(fixture, "")
	defer cleanup()

	if isoFixture == fixture {
		t.Fatal("isolateFixture should return a different path")
	}

	// Modify the isolated copy.
	isoFile := filepath.Join(isoFixture, "data.txt")
	if err := os.WriteFile(isoFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("write to copy: %v", err)
	}

	// Verify original is unchanged.
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(data) != "original" {
		t.Errorf("original was modified: got %q, want %q", data, "original")
	}
}

func TestIsolateFixture_EmptyFixture(t *testing.T) {
	path, cleanup := isolateFixture("", "")
	defer cleanup()
	if path != "" {
		t.Errorf("expected empty path, got %q", path)
	}
}

func TestIsolateFixture_DockerSkips(t *testing.T) {
	fixture := t.TempDir()
	path, cleanup := isolateFixture(fixture, "some-image:latest")
	defer cleanup()
	if path != fixture {
		t.Errorf("expected original fixture %q with docker, got %q", fixture, path)
	}
}
