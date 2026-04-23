package report

import (
	"path/filepath"
	"testing"
)

func TestChecksum(t *testing.T) {
	c1 := Checksum("hello world")
	c2 := Checksum("hello world")
	c3 := Checksum("different")

	if c1 != c2 {
		t.Error("same input should produce same checksum")
	}
	if c1 == c3 {
		t.Error("different input should produce different checksum")
	}
}

func TestSaveAndLoadSnapshots(t *testing.T) {
	dir := t.TempDir()

	sf := &SnapshotFile{
		Snapshots: []Snapshot{
			{Name: "hover", Tool: "get_info_on_location", Text: `{"type":"Person"}`, IsError: false, Checksum: Checksum(`{"type":"Person"}`)},
			{Name: "broken", Tool: "bad_tool", Text: "error", IsError: true, Checksum: Checksum("error")},
		},
	}

	if err := SaveSnapshots(dir, sf); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify file exists.
	path := SnapshotPath(dir)
	if filepath.Base(path) != ".snapshots.json" {
		t.Errorf("unexpected snapshot path: %s", path)
	}

	loaded, err := LoadSnapshots(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(loaded.Snapshots))
	}
	if loaded.Snapshots[0].Name != "hover" {
		t.Errorf("expected name 'hover', got %q", loaded.Snapshots[0].Name)
	}
	if loaded.Snapshots[1].IsError != true {
		t.Error("expected isError true for broken")
	}
}

func TestLoadSnapshots_Empty(t *testing.T) {
	dir := t.TempDir()
	sf, err := LoadSnapshots(dir)
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if len(sf.Snapshots) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(sf.Snapshots))
	}
}

func TestCompareSnapshot_Match(t *testing.T) {
	text := `{"result": "ok"}`
	saved := Snapshot{
		Name:     "test",
		Text:     text,
		IsError:  false,
		Checksum: Checksum(text),
	}

	err := CompareSnapshot(saved, text, false)
	if err != nil {
		t.Fatalf("expected match, got: %v", err)
	}
}

func TestCompareSnapshot_Changed(t *testing.T) {
	saved := Snapshot{
		Name:     "test",
		Text:     `{"result": "old"}`,
		IsError:  false,
		Checksum: Checksum(`{"result": "old"}`),
	}

	err := CompareSnapshot(saved, `{"result": "new"}`, false)
	if err == nil {
		t.Fatal("expected diff error")
	}
}

func TestCompareSnapshot_IsErrorChanged(t *testing.T) {
	saved := Snapshot{
		Name:     "test",
		Text:     "ok",
		IsError:  false,
		Checksum: Checksum("ok"),
	}

	err := CompareSnapshot(saved, "ok", true)
	if err == nil {
		t.Fatal("expected isError change error")
	}
}
