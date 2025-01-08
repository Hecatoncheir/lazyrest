package suites

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OnEscapeCallbackType func()

func onInputCallback(widget *Suites) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			widget.onEscapeCallback()
		}

		switch event.Rune() {
		case 'j':
			element := widget.Element.(*tview.List)
			currentItem := element.GetCurrentItem()
			element.SetCurrentItem(currentItem + 1)
		case 'k':
			element := widget.Element.(*tview.List)
			currentItem := element.GetCurrentItem()
			element.SetCurrentItem(currentItem - 1)
		}

		return event
	}
}
