package ui

import "github.com/Hecatoncheir/lazyrest/parser/http"

func onSuiteRun(application *Application) func(suite http.HttpSuite) {
	applicationElement := application.Element
	return func(suite http.HttpSuite) {
		application.Model.update(func(state *State) {
			state.Request = TaskState{Phase: PhaseLoading}
		})
		application.refreshStatus()
		producerElement := application.Producer.Element
		applicationElement.SetFocus(producerElement)
		application.Producer.ChangeSuite(suite)
	}
}

func onRunFinished(application *Application) func(error) {
	return func(err error) {
		application.Model.update(func(state *State) {
			state.Request = taskState(err)
		})
		application.refreshStatus()
	}
}

func onRunProgress(application *Application) func(current, total int64) {
	return func(current, total int64) {
		application.Model.update(func(state *State) {
			if state.Request.Phase != PhaseLoading {
				return
			}
			state.Request.Current = current
			state.Request.Total = total
			state.Request.HasProgress = true
		})
		application.refreshStatus()
	}
}
