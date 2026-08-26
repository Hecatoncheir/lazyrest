package ui

import (
	"maps"
	"slices"

	"github.com/Hecatoncheir/lazyrest/environment"
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

		go func() {
			selectedEnvironment := environment.Environment{
				Name:            application.config.EnvironmentName,
				Values:          maps.Clone(application.config.ParseOptions.Variables),
				SecretVariables: append([]string(nil), application.config.ParseOptions.SecretVariables...),
			}
			if selectedEnvironment.Values == nil {
				selectedEnvironment.Values = map[string]string{}
			}
			loadedEnvironment, environmentError := application.loadEnvironment(
				application.Model.Snapshot().RootDirectoryPath,
				application.config.Environment,
			)
			if environmentError == nil {
				for key, value := range loadedEnvironment.Values {
					selectedEnvironment.Values[key] = value
				}
				selectedEnvironment.SecretVariables = append(selectedEnvironment.SecretVariables, loadedEnvironment.SecretVariables...)
				slices.Sort(selectedEnvironment.SecretVariables)
				selectedEnvironment.SecretVariables = slices.Compact(selectedEnvironment.SecretVariables)
				if loadedEnvironment.Name != "" {
					selectedEnvironment.Name = loadedEnvironment.Name
				}
			}
			scanResult := application.scanFiles(ctx)
			if !treeWidget.IsCurrentReload(reloadID) {
				return
			}

			application.Element.QueueUpdateDraw(func() {
				if !treeWidget.FinishReload(reloadID) {
					return
				}
				if environmentError == nil {
					application.Suites.SetParseOptions(parserhttp.ParseOptions{
						Variables:       selectedEnvironment.Values,
						SecretVariables: selectedEnvironment.SecretVariables,
					})
				}
				application.Model.update(func(state *State) {
					state.Startup = taskState(environmentError)
					state.Files = taskState(scanResult.Err)
					if environmentError == nil {
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
