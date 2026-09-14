package suites

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OnEscapeCallbackType func()

// onInputCallback decides what a key press means in the request list. While the
// list is searching every key belongs to the query; otherwise the key opens a
// request, runs one of the list's commands, or is passed on.
func onInputCallback(widget *Suites) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if widget.searchMode {
			widget.searchInput(event)
			return nil
		}
		if widget.keybindings.Matches(keymap.Back, event) {
			widget.onEscapeCallback()
			return nil
		}
		if widget.keybindings.Matches(keymap.Open, event) {
			// tview selects an item on Enter, so the configured key is
			// translated into the one the widget already understands.
			return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		}
		if event.Key() == tcell.KeyEnter {
			// A bare Enter that was not the configured key selects nothing.
			return nil
		}
		if widget.command(event) {
			return nil
		}
		return event
	}
}

// searchInput edits the query. A key that is neither text nor the end of the
// search leaves the query alone, but is still swallowed: the list is searching,
// not navigating.
func (widget *Suites) searchInput(event *tcell.EventKey) {
	switch {
	case widget.keybindings.Matches(keymap.SearchFinish, event):
		widget.searchMode = false
	case event.Key() == tcell.KeyBackspace, event.Key() == tcell.KeyBackspace2:
		widget.searchQuery = withoutLastRune(widget.searchQuery)
	case event.Rune() != 0:
		widget.searchQuery += string(event.Rune())
	default:
		return
	}
	widget.render()
}

// withoutLastRune drops one character rather than one byte, so a backspace
// removes a whole letter in any alphabet.
func withoutLastRune(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}
	return string(runes[:len(runes)-1])
}

// command runs what the key asks of the list, and reports whether the key asked
// anything at all.
func (widget *Suites) command(event *tcell.EventKey) bool {
	switch {
	case widget.keybindings.Matches(keymap.Search, event):
		widget.searchMode = true
		widget.searchQuery = ""
		widget.render()
	case widget.keybindings.Matches(keymap.MoveDown, event):
		widget.moveSelection(1)
	case widget.keybindings.Matches(keymap.MoveUp, event):
		widget.moveSelection(-1)
	default:
		return false
	}
	return true
}

func (widget *Suites) moveSelection(delta int) {
	element := widget.Element.(*tview.List)
	element.SetCurrentItem(element.GetCurrentItem() + delta)
}
