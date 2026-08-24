package ui

import "github.com/Hecatoncheir/lazyrest/parser/http"

func onSuiteRun(application *Application) func(suite http.HttpSuite) {
	applicationElement := application.Element
	return func(suite http.HttpSuite) {
		producerElement := application.Producer.Element
		applicationElement.SetFocus(producerElement)
		application.Producer.ChangeSuite(suite)
	}
}
