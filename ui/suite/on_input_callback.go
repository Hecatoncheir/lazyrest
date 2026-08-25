package suite

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/parser/http"

	"github.com/gdamore/tcell/v2"
)

type OnEscapeCallbackType func()

type OnRunCallbackType func(suite http.HttpSuite)

func onInputCallback(widget *Suite) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case widget.keybindings.Matches(keymap.Back, event):
			widget.onEscapeCallback()
			return nil
		case widget.keybindings.Matches(keymap.Run, event):
			widget.onRunCallback(widget.suite)
			return nil
		}
		return event
	}
}
