package ui

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
)

func onSelectFileCallback(applicationWidget *Application) tree.OnSelectFileCallbackType {
	return func(file finder.File) {
		applicationWidget.Model.update(func(state *State) {
			selectedFile := file
			state.SelectedFile = &selectedFile
			state.SelectedSuite = nil
			state.Suites = nil
			state.Diagnostics = nil
			state.Parser = TaskState{Phase: PhaseLoading}
			state.Request = TaskState{}
		})
		applicationWidget.Footer.SelectFile(file)
		applicationWidget.Suite.Clear()
		applicationWidget.refreshDiagnostics()

		suitesWidget := applicationWidget.Suites
		ctx, loadID := suitesWidget.StartLoad()
		suitesWidget.ShowLoading(file)
		suitesElement := applicationWidget.Suites.Element

		applicationElement := applicationWidget.Element
		applicationElement.SetFocus(suitesElement)

		go func() {
			result := suitesWidget.LoadSuitesFromFile(ctx, file)
			if !suitesWidget.IsCurrentLoad(loadID) {
				return
			}
			applicationElement.QueueUpdateDraw(func() {
				if !suitesWidget.FinishLoad(loadID) {
					return
				}
				applicationWidget.Model.update(func(state *State) {
					state.Parser = taskState(result.Err)
					if result.Err == nil {
						state.Suites = append([]http.HttpSuite(nil), result.Suites...)
						state.Diagnostics = append([]http.Diagnostic(nil), result.Diagnostics...)
					} else {
						state.Suites = nil
						state.Diagnostics = nil
					}
				})
				suitesWidget.ApplyLoadResult(result)
				applicationWidget.refreshDiagnostics()
			})
		}()
	}
}
