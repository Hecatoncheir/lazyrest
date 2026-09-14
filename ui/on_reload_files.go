package ui

import (
	"path/filepath"

	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
)

func onReloadFiles(application *Application) func() {
	return func() {
		application.reloadFiles(nil)
	}
}

// reloadFiles rescans the tree in the background and puts the result on screen
// when it is still the current scan.
func (application *Application) reloadFiles(changedFiles map[string]struct{}) {
	treeWidget := application.HttpFilesTree
	ctx, reloadID := treeWidget.StartReload()
	treeWidget.ShowReloading()
	application.Model.update(func(state *State) {
		state.Files = TaskState{Phase: PhaseLoading}
	})
	application.refreshDiagnostics()

	go func() {
		result := treeWidget.Scan(ctx)
		if !treeWidget.IsCurrentReload(reloadID) {
			return
		}
		application.Element.QueueUpdateDraw(func() {
			if treeWidget.FinishReload(reloadID) {
				application.applyScan(result, changedFiles)
			}
		})
	}()
}

// applyScan shows a finished scan: the new tree, and whatever the change did to
// the file that was open.
func (application *Application) applyScan(result tree.ScanResult, changedFiles map[string]struct{}) {
	kept := application.rememberSelection(result)
	application.HttpFilesTree.ApplyScanResult(result)

	if kept.removed {
		application.Suites.Clear()
		application.Suite.Clear()
		application.Footer.DeselectFile()
	}
	if kept.needsReloading(changedFiles) {
		application.loadFile(finder.File{Name: filepath.Base(kept.path), Path: kept.path}, false)
	}
	application.refreshDiagnostics()
}

// keptFile is what a scan left of the file that was open.
type keptFile struct {
	path    string
	removed bool
}

// needsReloading reports whether the open file has to be read again: because
// the watcher named it, or because this reload came from somewhere else and
// nothing says the file is unchanged.
func (kept keptFile) needsReloading(changedFiles map[string]struct{}) bool {
	if kept.path == "" || kept.removed {
		return false
	}
	if changedFiles == nil {
		return true
	}
	_, named := changedFiles[kept.path]
	return named
}

// rememberSelection records the scan and forgets the open file when the scan no
// longer finds it.
func (application *Application) rememberSelection(result tree.ScanResult) keptFile {
	var kept keptFile
	application.Model.update(func(state *State) {
		state.Files = taskState(result.Err)
		if result.Err != nil {
			return
		}
		state.Directory = result.Directory
		if state.SelectedFile == nil {
			return
		}
		if directoryContainsFile(result.Directory, state.SelectedFile.Path) {
			kept.path = state.SelectedFile.Path
			return
		}
		kept.removed = true
		application.forgetSelection(state)
	})
	return kept
}

// forgetSelection clears everything that belonged to a file the scan no longer
// finds. A request still in flight keeps its own state, because it is still in
// flight.
func (application *Application) forgetSelection(state *State) {
	state.SelectedFile = nil
	state.SelectedSuite = nil
	state.Suites = nil
	state.Diagnostics = nil
	state.Parser = TaskState{}
	if !application.Producer.IsRunning() {
		state.Request = TaskState{}
	}
}
