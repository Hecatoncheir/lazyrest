package ui

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchRequestFilesDebouncesWrites(t *testing.T) {
	// The burst below has to land inside one debounce window, and the writes
	// are only as fast as the machine is. With the production 150ms this test
	// fails on a loaded CI runner even though the debouncing is correct.
	previousDebounce := fileWatchDebounce
	fileWatchDebounce = time.Second
	t.Cleanup(func() { fileWatchDebounce = previousDebounce })

	root := t.TempDir()
	path := filepath.Join(root, "requests.http")
	if err := os.WriteFile(path, []byte("GET https://example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	changes := make(chan map[string]struct{}, 2)
	done := make(chan error, 1)
	go func() {
		done <- watchRequestFiles(ctx, root, nil, func() { close(ready) }, func(paths map[string]struct{}) {
			changes <- paths
		})
	}()
	select {
	case <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not become ready")
	}

	for index := range 3 {
		if err := os.WriteFile(path, []byte{byte('0' + index)}, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case changed := <-changes:
		if _, found := changed[path]; !found {
			t.Fatalf("changed request was not reported: %#v", changed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("request change was not reported")
	}
	select {
	case extra := <-changes:
		t.Fatalf("rapid writes were not debounced: %#v", extra)
	// The wait has an absolute floor rather than being a multiple of the
	// window. Scaled purely to the window, a debounce shortened to nothing
	// would leave too little time for the extra reports to arrive, and the
	// test would pass while reporting every write separately.
	case <-time.After(fileWatchDebounce + time.Second):
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop")
	}
}

func TestWatchRequestFilesFindsRequestsInNewDirectories(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	changes := make(chan map[string]struct{}, 1)
	go func() {
		_ = watchRequestFiles(ctx, root, nil, func() { close(ready) }, func(paths map[string]struct{}) {
			changes <- paths
		})
	}()
	<-ready

	directory := filepath.Join(root, "nested")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "new.hurl")
	if err := os.WriteFile(path, []byte("GET https://example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case changed := <-changes:
		if _, directoryFound := changed[directory]; !directoryFound {
			if _, fileFound := changed[path]; !fileFound {
				t.Fatalf("new request location was not reported: %#v", changed)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("new request directory was not reported")
	}
}
