package ui

func onSuitesEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		application.Model.update(func(state *State) {
			state.SelectedFile = nil
			state.SelectedSuite = nil
			state.Suites = nil
			state.Diagnostics = nil
			state.Parser = TaskState{}
		})
		application.Suites.Clear()
		application.Suite.Clear()
		footerWidget := application.Footer
		footerWidget.DeselectFile()
		application.refreshDiagnostics()

		httpFilesTreeElement := application.HttpFilesTree.Element
		applicationElement.SetFocus(httpFilesTreeElement)
	}
}
