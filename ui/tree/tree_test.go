package tree

import (
	"context"
	"errors"
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

func TestReloadInputInvokesCallback(t *testing.T) {
	called := false
	widget := New()
	parameters := newMockParams(t.TempDir())
	parameters.OnReloadCallback = func() { called = true }
	widget.Build(parameters)

	event := tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone)
	if returned := onInputCallback(widget)(event); returned != nil {
		t.Fatal("reload event was not consumed")
	}
	if !called {
		t.Fatal("reload callback was not invoked")
	}
}

func TestApplyScanResultPreservesSelectedFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "request.http")
	if err := os.WriteFile(filePath, []byte("GET https://example.com"), 0644); err != nil {
		t.Fatal(err)
	}

	widget := New()
	widget.Build(newMockParams(tempDir))
	widget.ApplyScanResult(widget.Scan(context.Background()))
	element := widget.Element.(*tview.TreeView)
	selected := findFileNode(element.GetRoot(), filePath)
	if selected == nil {
		t.Fatal("fixture file is missing from the tree")
	}
	element.SetCurrentNode(selected)

	result := widget.Scan(context.Background())
	widget.ApplyScanResult(result)
	if got := currentFilePath(element.GetCurrentNode()); got != filePath {
		t.Fatalf("selected file was not preserved: got %q, want %q", got, filePath)
	}
}

func TestStartReloadCancelsPreviousReload(t *testing.T) {
	widget := New()
	firstContext, firstID := widget.StartReload()
	secondContext, secondID := widget.StartReload()

	if !errors.Is(firstContext.Err(), context.Canceled) {
		t.Fatalf("previous reload was not cancelled: %v", firstContext.Err())
	}
	if widget.IsCurrentReload(firstID) || !widget.IsCurrentReload(secondID) {
		t.Fatal("current reload identifier was not updated")
	}

	widget.CancelReload()
	if !errors.Is(secondContext.Err(), context.Canceled) {
		t.Fatalf("active reload was not cancelled: %v", secondContext.Err())
	}
}
