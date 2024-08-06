package suite

import (
	"lazyrest/parser/http"

	"github.com/gdamore/tcell/v2"
)

type OnEscapeCallbackType func()

type OnRunCallbackType func(suite http.HttpSuite)

func onInputCallback(widget *Suite) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			widget.onEscapeCallback()
		case tcell.KeyEnter:
			widget.onRunCallback(widget.suite)
		}
		return event
	}
}
