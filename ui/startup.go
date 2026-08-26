package ui

import (
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
)

func (application *Application) Start() {
	application.startOnce.Do(func() {
		treeWidget := application.HttpFilesTree
		ctx, reloadID := treeWidget.StartReload()
		treeWidget.ShowLoading()
		application.Model.update(func(state *State) {
			state.Startup = TaskState{Phase: PhaseLoading}
			state.Files = TaskState{Phase: PhaseLoading}
		})
		application.refreshDiagnostics()
		environmentLoadID := application.startEnvironmentLoad()

		go func() {
			selectedEnvironment, environmentError := application.loadMergedEnvironment(
				application.Model.Snapshot().RootDirectoryPath,
				application.config.Environment,
				application.config.EnvironmentName,
			)
			scanResult := application.scanFiles(ctx)
			if !treeWidget.IsCurrentReload(reloadID) {
				return
			}

			application.Element.QueueUpdateDraw(func() {
				if !treeWidget.FinishReload(reloadID) {
					return
				}
				currentEnvironmentLoad := application.isCurrentEnvironmentLoad(environmentLoadID)
				if environmentError == nil && currentEnvironmentLoad {
					application.Suites.SetParseOptions(parserhttp.ParseOptions{
						Variables:       selectedEnvironment.Values,
						SecretVariables: selectedEnvironment.SecretVariables,
					})
				}
				application.Model.update(func(state *State) {
					if currentEnvironmentLoad {
						state.Startup = taskState(environmentError)
					} else {
						state.Startup = TaskState{Phase: PhaseReady}
					}
					state.Files = taskState(scanResult.Err)
					if environmentError == nil && currentEnvironmentLoad {
						state.EnvironmentName = selectedEnvironment.Name
					}
					if scanResult.Err == nil {
						state.Directory = scanResult.Directory
					}
				})
				treeWidget.ApplyScanResult(scanResult)
				application.refreshDiagnostics()
			})
		}()
	})
}

func taskState(err error) TaskState {
	if err != nil {
		return TaskState{Phase: PhaseFailed, Error: err.Error()}
	}
	return TaskState{Phase: PhaseReady}
}
