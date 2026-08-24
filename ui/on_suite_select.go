package ui

import "github.com/Hecatoncheir/lazyrest/parser/http"

func onSuiteSelect(application *Application) func(suite http.HttpSuite) {
	return func(suite http.HttpSuite) {
		application.Model.update(func(state *State) {
			selectedSuite := suite
			state.SelectedSuite = &selectedSuite
		})
		suiteWidget := application.Suite
		suiteWidget.ChangeSuite(suite)

		applicationElement := application.Element
		suiteElement := application.Suite.Element
		applicationElement.SetFocus(suiteElement)

		application.Footer.UpdateSuite(suite.Redact(suite.Name))
	}
}
