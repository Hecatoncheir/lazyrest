package ui

func onProducerEscape(application *Application) func() {
	applicationElement := application.Element
	return func() {
		application.Producer.CancelActive()
		application.Model.update(func(state *State) {
			state.Request = TaskState{}
		})
		application.refreshStatus()
		suiteElement := application.Suite.Element
		applicationElement.SetFocus(suiteElement)
	}
}
