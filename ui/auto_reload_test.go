package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Hecatoncheir/lazyrest/finder"
)

func TestTUIAutomaticallyReparsesTheSelectedRequestFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "requests.http")
	if err := os.WriteFile(path, []byte("GET https://one.example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	application := BuildApplication(root, Config{})
	_, _ = runTestApplication(t, application)
	application.Start()
	waitFor(t, "initial file discovery", func() bool {
		return application.Model.Snapshot().Files.Phase == PhaseReady
	})
	application.fileWatcherMutex.Lock()
	watcherReady := application.fileWatcherReady
	application.fileWatcherMutex.Unlock()
	select {
	case <-watcherReady:
	case <-time.After(2 * time.Second):
		t.Fatal("file watcher did not become ready")
	}
	application.Element.QueueUpdateDraw(func() {
		application.loadFile(finder.File{Name: filepath.Base(path), Path: path}, true)
	})
	waitFor(t, "initial request parsing", func() bool {
		state := application.Model.Snapshot()
		return len(state.Suites) == 1 && state.Suites[0].Uri == "https://one.example.test"
	})

	if err := os.WriteFile(path, []byte("GET https://two.example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "automatic request reparse", func() bool {
		state := application.Model.Snapshot()
		return state.Files.Phase == PhaseReady && len(state.Suites) == 1 && state.Suites[0].Uri == "https://two.example.test"
	})
}
