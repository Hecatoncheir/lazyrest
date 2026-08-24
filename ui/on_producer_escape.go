package ui

func onProducerEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		application.Producer.CancelActive()
		suiteElement := application.Suite.Element
		applicationElement.SetFocus(suiteElement)
	}
}
