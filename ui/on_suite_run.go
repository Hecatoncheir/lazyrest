package ui

import (
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

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

func onRunFinished(application *Application) func(runner.Response, error) {
	return func(response runner.Response, err error) {
		application.Model.update(func(state *State) {
			request := requestTaskState(response, err)
			request.Current = state.Request.Current
			request.Total = state.Request.Total
			request.HasProgress = state.Request.HasProgress
			state.Request = request
		})
		application.refreshStatus()
	}
}

func requestTaskState(response runner.Response, err error) TaskState {
	if err != nil {
		return TaskState{Phase: PhaseFailed, Error: err.Error(), Outcome: OutcomeFailure}
	}
	if !response.IsSuccessful() {
		return TaskState{Phase: PhaseReady, Outcome: OutcomeFailure}
	}
	return TaskState{Phase: PhaseReady, Outcome: OutcomeSuccess}
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
