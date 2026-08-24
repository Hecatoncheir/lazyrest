package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestTUIRemainsInteractiveDuringBackgroundStartup(t *testing.T) {
	root := t.TempDir()
	application := BuildApplication(root, Config{Environment: environment.Config{Name: "test"}})
	releaseScan := make(chan struct{})
	application.scanFiles = func(ctx context.Context) tree.ScanResult {
		select {
		case <-releaseScan:
			return tree.ScanResult{Directory: finder.Directory{Name: filepath.Base(root), Path: root}}
		case <-ctx.Done():
			return tree.ScanResult{Err: ctx.Err()}
		}
	}
	application.loadEnvironment = func(string, environment.Config) (environment.Environment, error) {
		return environment.Environment{Name: "test", Values: map[string]string{}}, nil
	}

	screen, _ := runTestApplication(t, application)
	application.Start()
	waitFor(t, "startup loading state", func() bool {
		return application.Model.Snapshot().Files.Phase == PhaseLoading
	})

	screen.InjectKey(tcell.KeyRune, '?', tcell.ModNone)
	waitFor(t, "help during startup", func() bool {
		return application.Model.Snapshot().Overlay == OverlayHelp &&
			strings.Contains(applicationText(application, screen), "Toggle Pretty / Raw body")
	})

	screen.InjectKey(tcell.KeyRune, '?', tcell.ModNone)
	close(releaseScan)
	waitFor(t, "completed startup", func() bool {
		state := application.Model.Snapshot()
		return state.Files.Phase == PhaseReady && state.Startup.Phase == PhaseReady
	})
}

func TestTUIDiagnosticsAndHelpWorkflow(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "requests.http")
	request := "# @name Missing variable\nGET https://{{missing}}/users\n"
	if err := os.WriteFile(filePath, []byte(request), 0o600); err != nil {
		t.Fatal(err)
	}

	application := BuildApplication(root, Config{})
	screen, _ := runTestApplication(t, application)
	application.Start()
	waitFor(t, "file discovery", func() bool {
		return application.Model.Snapshot().Files.Phase == PhaseReady
	})

	application.Element.QueueUpdateDraw(func() {
		treeView := application.HttpFilesTree.Element.(*tview.TreeView)
		node := findFileReference(treeView.GetRoot(), filePath)
		if node == nil {
			t.Errorf("request file %q is missing from the tree", filePath)
			return
		}
		treeView.SetCurrentNode(node)
		application.Element.SetFocus(treeView)
	})
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	waitFor(t, "parser diagnostics", func() bool {
		state := application.Model.Snapshot()
		return state.Parser.Phase == PhaseReady && len(state.Diagnostics) > 0
	})

	screen.InjectKey(tcell.KeyRune, 'd', tcell.ModNone)
	waitFor(t, "diagnostics window", func() bool {
		return application.Model.Snapshot().Overlay == OverlayDiagnostics &&
			strings.Contains(applicationText(application, screen), "undefined variable: missing")
	})

	screen.InjectKey(tcell.KeyRune, '?', tcell.ModNone)
	waitFor(t, "help window", func() bool {
		return application.Model.Snapshot().Overlay == OverlayHelp &&
			strings.Contains(applicationText(application, screen), "Ctrl+h/j/k/l")
	})

	screen.InjectKey(tcell.KeyEsc, 0, tcell.ModNone)
	waitFor(t, "closed overlay", func() bool {
		return application.Model.Snapshot().Overlay == OverlayNone
	})
}

func runTestApplication(t *testing.T, application *Application) (tcell.SimulationScreen, <-chan error) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	application.Element.SetScreen(screen)
	screen.SetSize(120, 36)
	done := make(chan error, 1)
	go func() {
		done <- application.Element.Run()
	}()
	waitFor(t, "initial draw", func() bool {
		return strings.Contains(applicationText(application, screen), "Files")
	})
	t.Cleanup(func() {
		application.Producer.CancelActive()
		application.Suites.CancelLoad()
		application.HttpFilesTree.CancelReload()
		application.Element.Stop()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("TUI stopped with an error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("TUI did not stop")
		}
	})
	return screen, done
}

func waitFor(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

func simulationText(screen tcell.SimulationScreen) string {
	cells, width, height := screen.GetContents()
	var output strings.Builder
	for row := 0; row < height; row++ {
		for column := 0; column < width; column++ {
			cell := cells[row*width+column]
			if len(cell.Runes) == 0 {
				output.WriteRune(' ')
				continue
			}
			output.WriteRune(cell.Runes[0])
		}
		output.WriteByte('\n')
	}
	return output.String()
}

func applicationText(application *Application, screen tcell.SimulationScreen) string {
	var content string
	application.Element.QueueUpdate(func() {
		content = simulationText(screen)
	})
	return content
}

func findFileReference(node *tview.TreeNode, path string) *tview.TreeNode {
	if node == nil {
		return nil
	}
	if file, ok := node.GetReference().(finder.File); ok && file.Path == path {
		return node
	}
	for _, child := range node.GetChildren() {
		if match := findFileReference(child, path); match != nil {
			return match
		}
	}
	return nil
}
