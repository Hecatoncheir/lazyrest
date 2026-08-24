package suites

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OnEscapeCallbackType func()

func onInputCallback(widget *Suites) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if widget.searchMode {
			switch event.Key() {
			case tcell.KeyEsc, tcell.KeyEnter:
				widget.searchMode = false
				widget.render()
				return nil
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				query := []rune(widget.searchQuery)
				if len(query) > 0 {
					widget.searchQuery = string(query[:len(query)-1])
				}
				widget.render()
				return nil
			}
			if event.Rune() != 0 {
				widget.searchQuery += string(event.Rune())
				widget.render()
			}
			return nil
		}

		switch event.Key() {
		case tcell.KeyEsc:
			widget.onEscapeCallback()
			return nil
		}

		switch event.Rune() {
		case '/':
			widget.searchMode = true
			widget.searchQuery = ""
			widget.render()
			return nil
		case 'j':
			element := widget.Element.(*tview.List)
			currentItem := element.GetCurrentItem()
			element.SetCurrentItem(currentItem + 1)
			return nil
		case 'k':
			element := widget.Element.(*tview.List)
			currentItem := element.GetCurrentItem()
			element.SetCurrentItem(currentItem - 1)
			return nil
		}

		return event
	}
}
