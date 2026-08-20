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
	dir, err := FindFilesInDirectory(badPath, ".http")

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
	// Create irrelevant files
	otherFile1 := filepath.Join(tempDir, "readme.txt")
	otherFile2 := filepath.Join(otherDir, "ignore.txt")
	os.WriteFile(targetFile1, []byte("content"), 0644)
	os.WriteFile(targetFile2, []byte("content"), 0644)
	os.WriteFile(otherFile1, []byte("content"), 0644)
	os.WriteFile(otherFile2, []byte("content"), 0644)

	// 2. Execution: Call the function under test
	actualDir, err := FindFilesInDirectory(tempDir, ".http")
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
	
	// Проверяем, что найденные файлы корректны (total 2)
	expectedFilesCount := 2
	if len(allFiles) != expectedFilesCount {
		t.Errorf("Expected %d files in total, got %d. Files found: %+v", expectedFilesCount, len(allFiles), allFiles)
	}
	
	// Проверяем, что оба файла были найдены и пути правильные
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

	// Убедимся, что 'readme.txt' и 'ignore.txt' проигнорированы.
}

func TestFindFilesInDirectory_NoMatchingFiles(t *testing.T) {
	// 1. Setup
	tempDir, err := os.MkdirTemp("", "test_finder_nofile_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create directories with non-matching files
	subDir := filepath.Join(tempDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "wrong.txt"), []byte("content"), 0644)

	// 2. Execution
	actualDir, err := FindFilesInDirectory(tempDir, ".http")
	if err != nil {
		t.Fatalf("FindFilesInDirectory failed unexpectedly: %v", err)
	}

	// 3. Assertions
	if len(actualDir.Files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(actualDir.Files))
	}
	// Проверяем, что все найденные поддиректории не содержат файлов
	for _, dir := range actualDir.Directories {
		if len(dir.Files) != 0 {
			t.Errorf("Expected sub-directory %s to have 0 files, but found %d", dir.Name, len(dir.Files))
		}
	}
}
