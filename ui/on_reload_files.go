package ui

import (
	"path/filepath"

	"github.com/Hecatoncheir/lazyrest/finder"
)

func onReloadFiles(application *Application) func() {
	return func() {
		application.reloadFiles(nil)
	}
}

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
			if !treeWidget.FinishReload(reloadID) {
				return
			}
			selectionRemoved := false
			var selectedFilePath string
			application.Model.update(func(state *State) {
				state.Files = taskState(result.Err)
				if result.Err == nil {
					state.Directory = result.Directory
					if state.SelectedFile != nil && !directoryContainsFile(result.Directory, state.SelectedFile.Path) {
						selectionRemoved = true
						state.SelectedFile = nil
						state.SelectedSuite = nil
						state.Suites = nil
						state.Diagnostics = nil
						state.Parser = TaskState{}
						if !application.Producer.IsRunning() {
							state.Request = TaskState{}
						}
					}
					if state.SelectedFile != nil {
						selectedFilePath = state.SelectedFile.Path
					}
				}
			})
			treeWidget.ApplyScanResult(result)
			if selectionRemoved {
				application.Suites.Clear()
				application.Suite.Clear()
				application.Footer.DeselectFile()
			}
			_, selectedChanged := changedFiles[selectedFilePath]
			if selectedFilePath != "" && !selectionRemoved && (changedFiles == nil || selectedChanged) {
				application.loadFile(finder.File{Name: filepath.Base(selectedFilePath), Path: selectedFilePath}, false)
			}
			application.refreshDiagnostics()
		})
	}()
}
