package ui

import (
	"context"

	"github.com/Hecatoncheir/lazyrest/environment"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
)

// Start brings the application up. The environment and the file tree are both
// read in the background so the panes are usable while they load.
func (application *Application) Start() {
	application.startOnce.Do(func() {
		application.startFileWatcher()
		treeWidget := application.HttpFilesTree
		ctx, reloadID := treeWidget.StartReload()
		treeWidget.ShowLoading()

		application.Model.update(func(state *State) {
			state.Startup = TaskState{Phase: PhaseLoading}
			state.Files = TaskState{Phase: PhaseLoading}
		})
		application.refreshDiagnostics()

		environmentLoadID := application.startEnvironmentLoad()
		go application.loadStartup(ctx, reloadID, environmentLoadID)
	})
}

// startup is what the background load produced, held together so the drawing
// step reads as one thing rather than five.
type startup struct {
	environment      environment.Environment
	environmentError error
	// currentEnvironment is false when the user switched environments while
	// this load was still running, in which case its result is stale.
	currentEnvironment bool
	scan               tree.ScanResult
}

func (application *Application) loadStartup(ctx context.Context, reloadID, environmentLoadID uint64) {
	loaded := startup{}
	loaded.environment, loaded.environmentError = application.loadMergedEnvironment(
		application.Model.Snapshot().RootDirectoryPath,
		application.config.Environment,
		application.config.EnvironmentName,
	)
	loaded.scan = application.scanFiles(ctx)

	treeWidget := application.HttpFilesTree
	if !treeWidget.IsCurrentReload(reloadID) {
		return
	}
	application.Element.QueueUpdateDraw(func() {
		if !treeWidget.FinishReload(reloadID) {
			return
		}
		loaded.currentEnvironment = application.isCurrentEnvironmentLoad(environmentLoadID)
		application.applyStartup(loaded)
	})
}

func (application *Application) applyStartup(loaded startup) {
	if loaded.environmentUsable() {
		application.Suites.SetParseOptions(parserhttp.ParseOptions{
			Variables:       loaded.environment.Values,
			SecretVariables: loaded.environment.SecretVariables,
		})
	}
	application.Model.update(func(state *State) {
		state.Startup = loaded.startupState()
		state.Files = taskState(loaded.scan.Err)
		if loaded.environmentUsable() {
			state.EnvironmentName = loaded.environment.Name
		}
		if loaded.scan.Err == nil {
			state.Directory = loaded.scan.Directory
		}
	})
	application.HttpFilesTree.ApplyScanResult(loaded.scan)
	application.refreshDiagnostics()
}

func (loaded startup) environmentUsable() bool {
	return loaded.environmentError == nil && loaded.currentEnvironment
}

// startupState reports a stale load as ready rather than failed: whatever went
// wrong belonged to an environment the user has already moved on from.
func (loaded startup) startupState() TaskState {
	if !loaded.currentEnvironment {
		return TaskState{Phase: PhaseReady}
	}
	return taskState(loaded.environmentError)
}

func taskState(err error) TaskState {
	if err != nil {
		return TaskState{Phase: PhaseFailed, Error: err.Error()}
	}
	return TaskState{Phase: PhaseReady}
}
