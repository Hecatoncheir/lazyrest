package suites

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OnEscapeCallbackType func()

func onInputCallback(widget *Suites) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if widget.searchMode {
			if widget.keybindings.Matches(keymap.SearchFinish, event) {
				widget.searchMode = false
				widget.render()
				return nil
			}
			switch event.Key() {
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

		if widget.keybindings.Matches(keymap.Back, event) {
			widget.onEscapeCallback()
			return nil
		}
		if widget.keybindings.Matches(keymap.Open, event) {
			return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		}
		if event.Key() == tcell.KeyEnter {
			return nil
		}

		switch {
		case widget.keybindings.Matches(keymap.Search, event):
			widget.searchMode = true
			widget.searchQuery = ""
			widget.render()
			return nil
		case widget.keybindings.Matches(keymap.MoveDown, event):
			element := widget.Element.(*tview.List)
			currentItem := element.GetCurrentItem()
			element.SetCurrentItem(currentItem + 1)
			return nil
		case widget.keybindings.Matches(keymap.MoveUp, event):
			element := widget.Element.(*tview.List)
			currentItem := element.GetCurrentItem()
			element.SetCurrentItem(currentItem - 1)
			return nil
		}

		return event
	}
}
