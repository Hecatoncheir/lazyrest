package ui

import "slices"

const maxHistoryErrors = 20

func (application *Application) recordHistoryError(err error) {
	if application == nil || application.Model == nil || err == nil {
		return
	}
	message := err.Error()
	application.Model.update(func(state *State) {
		if slices.Contains(state.HistoryErrors, message) {
			return
		}
		state.HistoryErrors = append(state.HistoryErrors, message)
		if len(state.HistoryErrors) > maxHistoryErrors {
			state.HistoryErrors = append([]string(nil), state.HistoryErrors[len(state.HistoryErrors)-maxHistoryErrors:]...)
		}
	})
}

func diagnosticsCount(state State) int {
	return len(state.Diagnostics) + len(state.HistoryErrors)
}
