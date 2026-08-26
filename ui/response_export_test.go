package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteResponseFileCreatesPrivateFileAndRequiresExplicitOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "response.json")
	if err := writeResponseFile(path, []byte("first"), false); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "first" {
		t.Fatalf("unexpected contents: %q", contents)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected permissions: %o", info.Mode().Perm())
	}

	if err := writeResponseFile(path, []byte("second"), false); err == nil {
		t.Fatal("existing response file was overwritten without confirmation")
	}
	contents, _ = os.ReadFile(path)
	if string(contents) != "first" {
		t.Fatalf("failed write changed the file: %q", contents)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeResponseFile(path, []byte("second"), true); err != nil {
		t.Fatal(err)
	}
	contents, _ = os.ReadFile(path)
	if string(contents) != "second" {
		t.Fatalf("confirmed overwrite did not replace the file: %q", contents)
	}
	info, _ = os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("overwritten file permissions were not secured: %o", info.Mode().Perm())
	}
}
