package suites

import (
	"github.com/gdamore/tcell/v2"
)

type OnEscapeCallbackType func()

func onInputCallback(widget *Suites) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			widget.onEscapeCallback()
		}
		return event
	}
}
