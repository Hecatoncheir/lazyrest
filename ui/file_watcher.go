package ui

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// fileWatchDebounce is a variable rather than a constant so a test can widen
// the window. A test that writes a burst and expects one report depends on the
// whole burst landing inside the window, which 150ms cannot guarantee on a
// loaded machine.
var fileWatchDebounce = 150 * time.Millisecond

var defaultWatchIgnoredDirectories = map[string]struct{}{
	".git": {}, ".hg": {}, ".svn": {}, ".cache": {}, ".venv": {}, ".tox": {},
	"node_modules": {}, "vendor": {}, "target": {}, "dist": {}, "build": {},
}

func (application *Application) startFileWatcher() {
	application.fileWatcherMutex.Lock()
	if application.fileWatcherCancel != nil {
		application.fileWatcherMutex.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	ready := make(chan struct{})
	application.fileWatcherCancel = cancel
	application.fileWatcherDone = done
	application.fileWatcherReady = ready
	application.fileWatcherMutex.Unlock()

	go func() {
		defer close(done)
		_ = watchRequestFiles(ctx, application.Model.Snapshot().RootDirectoryPath, application.config.Ignore, func() {
			close(ready)
		}, func(paths map[string]struct{}) {
			application.Element.QueueUpdateDraw(func() {
				application.reloadFiles(paths)
			})
		})
	}()
}

func (application *Application) stopFileWatcher() {
	application.fileWatcherMutex.Lock()
	cancel := application.fileWatcherCancel
	done := application.fileWatcherDone
	application.fileWatcherCancel = nil
	application.fileWatcherDone = nil
	application.fileWatcherReady = nil
	application.fileWatcherMutex.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	if done != nil {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	}
}

// requestWatcher is the state watching a tree needs: the watcher itself, the
// directory names to skip, the directories currently watched, and the changes
// waiting to be reported.
type requestWatcher struct {
	watcher *fsnotify.Watcher
	ignore  map[string]struct{}
	watched map[string]struct{}
	pending map[string]struct{}
}

func newRequestWatcher(ignored []string) (*requestWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &requestWatcher{
		watcher: watcher,
		ignore:  ignoredDirectoryNames(ignored),
		watched: make(map[string]struct{}),
		pending: make(map[string]struct{}),
	}, nil
}

// ignoredDirectoryNames collects the directory names never to descend into:
// the built in ones and whatever the configuration added. Only the last part
// of a configured path is used, since that is what a walk compares against.
func ignoredDirectoryNames(ignored []string) map[string]struct{} {
	names := make(map[string]struct{}, len(defaultWatchIgnoredDirectories)+len(ignored))
	for name := range defaultWatchIgnoredDirectories {
		names[name] = struct{}{}
	}
	for _, name := range ignored {
		base := filepath.Base(filepath.Clean(name))
		if base == "." || base == string(filepath.Separator) {
			continue
		}
		names[base] = struct{}{}
	}
	return names
}

func (watch *requestWatcher) close() { _ = watch.watcher.Close() }

// addTree watches path and every directory under it. A failure on path itself
// is fatal, because nothing would be watched; a failure deeper down is not,
// because the rest of the tree still is.
func (watch *requestWatcher) addTree(path string) error {
	return filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return watch.rootOnly(current, path, walkErr)
		}
		if !entry.IsDir() {
			return nil
		}
		if current != path {
			if _, skip := watch.ignore[entry.Name()]; skip {
				return filepath.SkipDir
			}
		}
		if err := watch.watcher.Add(current); err != nil {
			return watch.rootOnly(current, path, err)
		}
		watch.watched[filepath.Clean(current)] = struct{}{}
		return nil
	})
}

func (watch *requestWatcher) rootOnly(current, root string, err error) error {
	if current == root {
		return err
	}
	return nil
}

// note records what an event changed, if anything worth reporting. It reports
// whether there is now something to report.
func (watch *requestWatcher) note(event fsnotify.Event) bool {
	const interesting = fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename
	if event.Op&interesting == 0 {
		return false
	}
	path := filepath.Clean(event.Name)

	if event.Op&fsnotify.Create != 0 && isDirectory(event.Name) {
		// A directory that appeared has to be watched too, or the files
		// created inside it later are never seen.
		_ = watch.addTree(event.Name)
		watch.pending[path] = struct{}{}
	}
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		if _, wasDirectory := watch.watched[path]; wasDirectory {
			delete(watch.watched, path)
			watch.pending[path] = struct{}{}
		}
	}
	if isRequestFile(event.Name) || filepath.Base(event.Name) == ".gitignore" {
		watch.pending[path] = struct{}{}
	}
	return len(watch.pending) > 0
}

// take hands over what has changed and starts collecting again.
func (watch *requestWatcher) take() map[string]struct{} {
	if len(watch.pending) == 0 {
		return nil
	}
	changed := make(map[string]struct{}, len(watch.pending))
	for path := range watch.pending {
		changed[path] = struct{}{}
	}
	clear(watch.pending)
	return changed
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// debounce collapses a burst of events into one report. Editors write a file
// more than once when saving it, and a reload per write would be wasted work.
type debounce struct {
	timer   *time.Timer
	expired <-chan time.Time
}

// restart begins the window again, so a report follows the last event of a
// burst rather than the first.
func (d *debounce) restart(after time.Duration) {
	if d.timer == nil {
		d.timer = time.NewTimer(after)
		d.expired = d.timer.C
		return
	}
	if !d.timer.Stop() {
		// The timer had already fired; drain it so the restart is not reported
		// immediately.
		select {
		case <-d.timer.C:
		default:
		}
	}
	d.timer.Reset(after)
	d.expired = d.timer.C
}

func (d *debounce) settle() { d.expired = nil }

func (d *debounce) stop() {
	if d.timer != nil {
		d.timer.Stop()
	}
}

// watchRequestFiles reports request files that change under root, one report
// per burst of writes, until ctx is done.
func watchRequestFiles(
	ctx context.Context,
	root string,
	ignored []string,
	onReady func(),
	onChange func(map[string]struct{}),
) error {
	watch, err := newRequestWatcher(ignored)
	if err != nil {
		return err
	}
	defer watch.close()

	if err := watch.addTree(root); err != nil {
		return err
	}
	if onReady != nil {
		onReady()
	}

	var window debounce
	defer window.stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, open := <-watch.watcher.Events:
			if !open {
				return nil
			}
			if watch.note(event) {
				window.restart(fileWatchDebounce)
			}
		case <-window.expired:
			window.settle()
			if changed := watch.take(); changed != nil {
				onChange(changed)
			}
		case _, open := <-watch.watcher.Errors:
			if !open {
				return nil
			}
		}
	}
}

func isRequestFile(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return extension == ".http" || extension == ".hurl" || extension == ".socket"
}
