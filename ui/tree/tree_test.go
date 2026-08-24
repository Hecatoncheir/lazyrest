package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Mock implementation for parameters
func newMockParams(root string) Parameters {
	return Parameters{
		RootDirectoryPath: root,
		FilesExtension:    []string{".http"},
		Theme: theme.Theme{
			Tree: theme.TreeTheme{
				Border:          tcell.ColorBlue,
				Background:      tcell.ColorBlack,
				BorderFocus:     tcell.ColorBlue,
				BackgroundFocus: tcell.ColorDarkGray,
				Title:           tcell.ColorWhite,
				TitleFocus:      tcell.ColorLightGray,
			},
		},
		OnSelectFileCallback: func(file finder.File) {},
	}
}

func TestTree_Build_Success(t *testing.T) {
	// 1. Setup: Create a temporary directory and files to be found by Build.
	tempDir, err := os.MkdirTemp("", "test_tree_build_")
	if err != nil {
		t.Fatalf("Failed to create temp root: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sub-directory and a file within it.
	subDir := filepath.Join(tempDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create sub-directory: %v", err)
	}
	targetFile := filepath.Join(subDir, "test.http")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// 2. Execution: Call Build with the temp directory as root.
	params := newMockParams(tempDir)
	widget := New()
	element := widget.Build(params)

	// 3. Assertions
	if element == nil {
		t.Fatal("Build returned nil element")
	}
}

func TestTree_Build_NoFilesFound(t *testing.T) {
	// 1. Setup: Create an empty directory.
	tempDir, err := os.MkdirTemp("", "test_tree_empty_")
	if err != nil {
		t.Fatalf("Failed to create temp root: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 2. Execution
	params := newMockParams(tempDir)
	widget := New()
	element := widget.Build(params)

	// 3. Assertions
	if element == nil {
		t.Fatal("Build returned nil element when no files found")
	}
}

func TestCollectMatchingNodes(t *testing.T) {
	root := tview.NewTreeNode("root").SetReference(finder.Directory{Name: "root"})
	directory := tview.NewTreeNode("api").SetReference(finder.Directory{Name: "api"})
	match := tview.NewTreeNode("users.http").SetReference(finder.File{Name: "users.http"})
	other := tview.NewTreeNode("projects.http").SetReference(finder.File{Name: "projects.http"})
	directory.AddChild(match).AddChild(other)
	root.AddChild(directory)

	var matches []*tview.TreeNode
	collectMatchingNodes(root, "users", &matches)
	if len(matches) != 1 || matches[0] != match {
		t.Fatalf("unexpected matches: %+v", matches)
	}
	if !directory.IsExpanded() || !root.IsExpanded() {
		t.Fatal("parents of the match were not expanded")
	}
}
