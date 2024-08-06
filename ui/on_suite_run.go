package ui

import "lazyrest/parser/http"

func onSuiteRun(application *Application) func(suite http.HttpSuite) {
	applicationElement := application.Element
	return func(suite http.HttpSuite) {
		producer := application.Producer
		producerElement := producer.Element
		applicationElement.SetFocus(producerElement)
		producer.ChangeSuite(suite)
	}
}
