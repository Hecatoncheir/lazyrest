package ui

func onSuitesEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		footerWidget := application.Footer
		footerWidget.DeselectFile()

		httpFilesTreeElement := application.HttpFilesTree.Element
		applicationElement.SetFocus(httpFilesTreeElement)
	}
}
