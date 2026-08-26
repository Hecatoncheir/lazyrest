package finder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFindFilesInDirectoryContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := FindFilesInDirectoryContext(ctx, t.TempDir(), []string{".http"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

// TestFailingOnReadDir checks that the function correctly returns an error if os.ReadDir fails.
func TestFailingOnReadDir(t *testing.T) {
	// This test requires simulating a failure condition for os.ReadDir,
	// which is hard without mocking the OS, but we can proceed by
	// testing the error propagation path structure.
	// NOTE: In a real-world scenario, we would mock os.ReadDir.
	// For now, we test the structure of the error handling.

	// Create a path unlikely to exist or readable to force an error on os.ReadDir
	badPath := filepath.Join("nonexistent", "path", "for", "test")
	dir, err := FindFilesInDirectory(badPath, []string{".http"})

	if err == nil {
		t.Errorf("Expected an error when reading a non-existent directory, but got nil")
	}
	if dir.Path != badPath {
		t.Errorf("Expected directory path to be correctly set to %s, got %s", badPath, dir.Path)
	}
}

func TestFindFilesInDirectory_Basic(t *testing.T) {
	// 1. Setup: Create a temporary structure
	tempDir, err := os.MkdirTemp("", "test_finder_")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }() // Clean up afterwards

	// Create deep structure
	subDir := filepath.Join(tempDir, "sub")
	otherDir := filepath.Join(tempDir, "other")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(otherDir, 0755); err != nil {
		t.Fatal(err)
	}
	targetFile1 := filepath.Join(tempDir, "file1.http")
	targetFile2 := filepath.Join(subDir, "file2.http")
	targetFileUpper := filepath.Join(tempDir, "file3.HTTP")

	// Create irrelevant files
	otherFile1 := filepath.Join(tempDir, "readme.txt")
	otherFile2 := filepath.Join(otherDir, "ignore.txt")

	if err := os.WriteFile(targetFile1, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetFile2, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetFileUpper, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(otherFile1, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(otherFile2, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	actualDir, err := FindFilesInDirectory(tempDir, []string{".http"})
	if err != nil {
		t.Fatalf("FindFilesInDirectory failed unexpectedly: %v", err)
	}

	// 3. Assertions

	// Helper to collect all files from the tree
	var allFiles []File
	var collectFiles func(d Directory)
	collectFiles = func(d Directory) {
		allFiles = append(allFiles, d.Files...)
		for _, subDir := range d.Directories {
			collectFiles(subDir)
		}
	}
	collectFiles(actualDir)

	// Проверяем, что нашли корректное количество директорий (sub, other)
	// Note: 'other' was skipped because it's empty. This is expected behavior based on current implementation.
	if len(actualDir.Directories) != 1 {
		t.Errorf("Expected 1 non-empty directory (sub), got %d. Directories found: %+v", len(actualDir.Directories), actualDir.Directories)
	}

	// Проверяем, что найденные файлы корректны (total 3)
	expectedFilesCount := 3
	if len(allFiles) != expectedFilesCount {
		t.Errorf("Expected %d files in total, got %d. Files found: %+v", expectedFilesCount, len(allFiles), allFiles)
	}

	// Проверяем, что все три файла были найдены и пути правильные
	foundPaths := make(map[string]bool)
	for _, file := range allFiles {
		foundPaths[file.Path] = true
	}
	if !foundPaths[targetFile1] {
		t.Errorf("Did not find the expected file: %s", targetFile1)
	}
	if !foundPaths[targetFile2] {
		t.Errorf("Did not find the expected file: %s", targetFile2)
	}
	if !foundPaths[filepath.Join(tempDir, "file3.HTTP")] {
		t.Errorf("Did not find the expected uppercase file: %s", filepath.Join(tempDir, "file3.HTTP"))
	}

	// Убедимся, что 'readme.txt' и 'ignore.txt' проигнорированы.
}

func TestFindFilesInDirectory_DotsInNames(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_dots_")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	targetFile := filepath.Join(tempDir, "my.test.file.http")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	actualDir, err := FindFilesInDirectory(tempDir, []string{".http"})
	if err != nil {
		t.Fatalf("FindFilesInDirectory failed: %v", err)
	}

	if len(actualDir.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(actualDir.Files))
	} else if actualDir.Files[0].Name != "my.test.file.http" {
		t.Errorf("Expected filename my.test.file.http, got %s", actualDir.Files[0].Name)
	}
}

func TestFindFilesInDirectory_ExtensionWithoutDot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_no_dot_")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	targetFile := filepath.Join(tempDir, "file.http")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	actualDir, err := FindFilesInDirectory(tempDir, []string{"http"})
	if err != nil {
		t.Fatalf("FindFilesInDirectory failed: %v", err)
	}

	if len(actualDir.Files) != 0 {
		t.Errorf("Expected 0 files when searching with extension without dot, got %d", len(actualDir.Files))
	}
}

func TestFindFilesInDirectory_IgnoresHeavyDirectories(t *testing.T) {
	tempDir := t.TempDir()
	ignoredDir := filepath.Join(tempDir, "node_modules", "package")
	if err := os.MkdirAll(ignoredDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ignoredDir, "hidden.http"), []byte("GET http://example.com"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "visible.http"), []byte("GET http://example.com"), 0644); err != nil {
		t.Fatal(err)
	}

	directory, err := FindFilesInDirectory(tempDir, []string{".http"})
	if err != nil {
		t.Fatal(err)
	}
	if len(directory.Files) != 1 || directory.Files[0].Name != "visible.http" {
		t.Fatalf("unexpected files: %+v", directory.Files)
	}
	if len(directory.Directories) != 0 {
		t.Fatalf("ignored directory was included: %+v", directory.Directories)
	}
}

func TestFindFollowsASymlinkedDirectoryOnce(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "shared")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "linked.http"), []byte("GET https://example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(project, "requests")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	directory, err := Find(context.Background(), project, Options{Extensions: []string{".http"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(directory.Directories) != 1 || len(directory.Directories[0].Files) != 1 {
		t.Fatalf("the symlinked directory was not read: %+v", directory)
	}
	if directory.Directories[0].Files[0].Name != "linked.http" {
		t.Fatalf("unexpected file: %+v", directory.Directories[0].Files)
	}
}

func TestFindTerminatesOnASymlinkLoop(t *testing.T) {
	root := t.TempDir()
	inner := filepath.Join(root, "inner")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "one.http"), []byte("GET https://example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A link pointing back at the root would recurse without end.
	if err := os.Symlink(root, filepath.Join(inner, "loop")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := Find(context.Background(), root, Options{Extensions: []string{".http"}}); err != nil {
			t.Error(err)
		}
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the scan did not terminate on a symlink loop")
	}
}

func TestFindHonoursExtraIgnoredDirectories(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"keep", "skip"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name, "request.http"), []byte("GET https://example.test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	directory, err := Find(context.Background(), root, Options{Extensions: []string{".http"}, Ignore: []string{"skip"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(directory.Directories) != 1 || directory.Directories[0].Name != "keep" {
		t.Fatalf("the ignored directory was read: %+v", directory.Directories)
	}
}

func TestFindStopsAtTheDepthLimitAndSaysSo(t *testing.T) {
	root := t.TempDir()
	deep := root
	for range 6 {
		deep = filepath.Join(deep, "level")
	}
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deep, "deep.http"), []byte("GET https://example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	directory, err := Find(context.Background(), root, Options{Extensions: []string{".http"}, MaxDepth: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(directory.Warnings) == 0 {
		t.Fatal("the depth limit was reached without a warning")
	}
	if !strings.Contains(directory.Warnings[0], "depth of 3") {
		t.Fatalf("unexpected warning: %q", directory.Warnings[0])
	}
}
