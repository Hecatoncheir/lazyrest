package ui

func onSuitesEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		httpFilesTreeElement := application.HttpFilesTree.Element
		applicationElement.SetFocus(httpFilesTreeElement)
	}
}
