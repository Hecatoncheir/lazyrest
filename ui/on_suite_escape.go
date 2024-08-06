package ui

func onSuiteEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		suitesElement := application.Suites.Element
		applicationElement.SetFocus(suitesElement)
	}
}
