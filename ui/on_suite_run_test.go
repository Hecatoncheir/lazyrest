package ui

import (
	"errors"
	"net/http"
	"testing"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
)

func TestRequestTaskState(t *testing.T) {
	tests := []struct {
		name        string
		response    runner.Response
		err         error
		phase       Phase
		outcome     Outcome
		footerState footer.IndicatorState
	}{
		{
			name:        "success",
			response:    runner.Response{StatusCode: http.StatusOK},
			phase:       PhaseReady,
			outcome:     OutcomeSuccess,
			footerState: footer.IndicatorSuccess,
		},
		{
			name:        "unsuccessful response",
			response:    runner.Response{StatusCode: http.StatusBadRequest},
			phase:       PhaseReady,
			outcome:     OutcomeFailure,
			footerState: footer.IndicatorFailure,
		},
		{
			name:        "execution error",
			err:         errors.New("request failed"),
			phase:       PhaseFailed,
			outcome:     OutcomeFailure,
			footerState: footer.IndicatorFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := requestTaskState(test.response, test.err)
			if request.Phase != test.phase || request.Outcome != test.outcome {
				t.Fatalf("unexpected request state: %+v", request)
			}
			state := State{Request: request}
			if got := footerIndicatorState(state); got != test.footerState {
				t.Fatalf("unexpected footer state: got %v, want %v", got, test.footerState)
			}
		})
	}
}

func TestSelectingSuiteResetsRequestOutcome(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	application.Model.update(func(state *State) {
		state.Request = TaskState{Phase: PhaseReady, Outcome: OutcomeSuccess}
	})

	onSuiteSelect(application)(parserhttp.HttpSuite{Name: "Next request"})

	request := application.Model.Snapshot().Request
	if request != (TaskState{}) {
		t.Fatalf("request state was not reset: %+v", request)
	}
}
