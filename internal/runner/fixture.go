package runner

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copyDir recursively copies the directory tree rooted at src into dst.
// dst must not already exist; it is created with the same permissions as src.
// File permissions are preserved. Symlinks are not followed.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}

// copyFile copies a single file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// isolateFixture copies the fixture directory to a temporary location and
// returns the temp path along with a cleanup function. If fixture is empty
// or docker isolation is in use (non-empty dockerImage), the original fixture
// is returned unchanged with a no-op cleanup.
func isolateFixture(fixture, dockerImage string) (string, func()) {
	if fixture == "" || dockerImage != "" {
		return fixture, func() {}
	}

	tmp, err := os.MkdirTemp("", "mcp-assert-fixture-*")
	if err != nil {
		// If we can't create a temp dir, fall back to the original.
		return fixture, func() {}
	}

	// Copy into a subdirectory so the structure mirrors the original.
	dst := filepath.Join(tmp, filepath.Base(fixture))
	if err := copyDir(fixture, dst); err != nil {
		os.RemoveAll(tmp)
		return fixture, func() {}
	}

	return dst, func() { os.RemoveAll(tmp) }
}
