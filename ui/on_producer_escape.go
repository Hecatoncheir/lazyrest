package ui

func onProducerEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		suiteElement := application.Suite.Element
		applicationElement.SetFocus(suiteElement)
	}
}
