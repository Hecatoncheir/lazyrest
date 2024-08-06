package ui

import "lazyrest/parser/http"

func onSuiteSelect(application *Application) func(suite http.HttpSuite) {
	return func(suite http.HttpSuite) {
		suiteWidget := application.Suite
		suiteWidget.ChangeSuite(suite)

		applicationElement := application.Element
		suiteElement := application.Suite.Element
		applicationElement.SetFocus(suiteElement)
	}
}
