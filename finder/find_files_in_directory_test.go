package finder

import (
	"os"
	"path/filepath"
	"testing"
)

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
	defer os.RemoveAll(tempDir) // Clean up afterwards

	// Create deep structure
	subDir := filepath.Join(tempDir, "sub")
	otherDir := filepath.Join(tempDir, "other")
	os.Mkdir(subDir, 0755)
	os.Mkdir(otherDir, 0755)

	// Create target files
	targetFile1 := filepath.Join(tempDir, "file1.http")
	targetFile2 := filepath.Join(subDir, "file2.http")
	targetFileUpper := filepath.Join(tempDir, "file3.HTTP")

	// Create irrelevant files
	otherFile1 := filepath.Join(tempDir, "readme.txt")
	otherFile2 := filepath.Join(otherDir, "ignore.txt")

	os.WriteFile(targetFile1, []byte("content"), 0644)
	os.WriteFile(targetFile2, []byte("content"), 0644)
	os.WriteFile(targetFileUpper, []byte("content"), 0644)
	os.WriteFile(otherFile1, []byte("content"), 0644)
	os.WriteFile(otherFile2, []byte("content"), 0644)

	// 2. Execution: Call the function under test
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
	defer os.RemoveAll(tempDir)

	targetFile := filepath.Join(tempDir, "my.test.file.http")
	os.WriteFile(targetFile, []byte("content"), 0644)

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
	defer os.RemoveAll(tempDir)

	targetFile := filepath.Join(tempDir, "file.http")
	os.WriteFile(targetFile, []byte("content"), 0644)

	// Searching with "http" instead of ".http"
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
