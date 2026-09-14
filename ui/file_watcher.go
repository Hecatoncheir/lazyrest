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

const fileWatchDebounce = 150 * time.Millisecond

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

func watchRequestFiles(ctx context.Context, root string, ignored []string, onReady func(), onChange func(map[string]struct{})) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() { _ = watcher.Close() }()

	ignore := make(map[string]struct{}, len(defaultWatchIgnoredDirectories)+len(ignored))
	for name := range defaultWatchIgnoredDirectories {
		ignore[name] = struct{}{}
	}
	for _, name := range ignored {
		if base := filepath.Base(filepath.Clean(name)); base != "." && base != string(filepath.Separator) {
			ignore[base] = struct{}{}
		}
	}
	watchedDirectories := make(map[string]struct{})
	addTree := func(path string) error {
		return filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if current == path {
					return walkErr
				}
				return nil
			}
			if !entry.IsDir() {
				return nil
			}
			if current != path {
				if _, skip := ignore[entry.Name()]; skip {
					return filepath.SkipDir
				}
			}
			if err := watcher.Add(current); err != nil {
				if current == path {
					return err
				}
				return nil
			}
			watchedDirectories[filepath.Clean(current)] = struct{}{}
			return nil
		})
	}
	if err := addTree(root); err != nil {
		return err
	}
	if onReady != nil {
		onReady()
	}

	pending := make(map[string]struct{})
	var timer *time.Timer
	var timerChannel <-chan time.Time
	flush := func() {
		if len(pending) == 0 {
			return
		}
		changed := make(map[string]struct{}, len(pending))
		for path := range pending {
			changed[path] = struct{}{}
		}
		clear(pending)
		onChange(changed)
	}
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			cleanedEventPath := filepath.Clean(event.Name)
			if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					_ = addTree(event.Name)
					pending[cleanedEventPath] = struct{}{}
				}
			}
			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
				if _, wasDirectory := watchedDirectories[cleanedEventPath]; wasDirectory {
					delete(watchedDirectories, cleanedEventPath)
					pending[cleanedEventPath] = struct{}{}
				}
			}
			if isRequestFile(event.Name) || filepath.Base(event.Name) == ".gitignore" {
				pending[cleanedEventPath] = struct{}{}
			}
			if len(pending) == 0 {
				continue
			}
			if timer == nil {
				timer = time.NewTimer(fileWatchDebounce)
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(fileWatchDebounce)
			}
			timerChannel = timer.C
		case <-timerChannel:
			flush()
			timerChannel = nil
		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
		}
	}
}

func isRequestFile(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return extension == ".http" || extension == ".hurl" || extension == ".socket"
}
