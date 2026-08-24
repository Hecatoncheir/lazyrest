package ui

func onReloadFiles(application *Application) func() {
	return func() {
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
							state.Request = TaskState{}
						}
					}
				})
				treeWidget.ApplyScanResult(result)
				if selectionRemoved {
					application.Suites.Clear()
					application.Suite.Clear()
					application.Footer.DeselectFile()
				}
				application.refreshDiagnostics()
			})
		}()
	}
}
