package ui

func onSuitesEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		application.Suites.CancelLoad()
		footerWidget := application.Footer
		footerWidget.DeselectFile()

		httpFilesTreeElement := application.HttpFilesTree.Element
		applicationElement.SetFocus(httpFilesTreeElement)
	}
}
