package producer

import (
	"fmt"
	"strings"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OnEscapeCallbackType func()

func onInputCallback(widget *Producer) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if widget.searchMode {
			if widget.keybindings.Matches(keymap.SearchFinish, event) {
				widget.searchMode = false
				widget.updateSearch()
				return nil
			}
			switch event.Key() {
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				query := []rune(widget.searchQuery)
				if len(query) > 0 {
					widget.searchQuery = string(query[:len(query)-1])
				}
				widget.updateSearch()
				return nil
			}
			if event.Rune() != 0 {
				widget.searchQuery += string(event.Rune())
				widget.updateSearch()
			}
			return nil
		}

		if widget.keybindings.Matches(keymap.Back, event) {
			widget.onEscapeCallback()
			return nil
		}
		switch {
		case widget.keybindings.Matches(keymap.Search, event):
			widget.searchMode = true
			widget.searchQuery = ""
			widget.updateSearch()
			return nil
		case widget.keybindings.Matches(keymap.HistoryPrevious, event):
			widget.showHistory(-1)
			return nil
		case widget.keybindings.Matches(keymap.HistoryNext, event):
			widget.showHistory(1)
			return nil
		case widget.keybindings.Matches(keymap.ToggleBody, event):
			widget.toggleBodyView()
			return nil
		case widget.keybindings.Matches(keymap.CopyResponseBody, event):
			if widget.onCopyBody != nil {
				widget.onCopyBody()
			}
			return nil
		case widget.keybindings.Matches(keymap.CopyResponse, event):
			if widget.onCopyResponse != nil {
				widget.onCopyResponse()
			}
			return nil
		case widget.keybindings.Matches(keymap.SaveResponse, event):
			if widget.onSaveResponse != nil {
				widget.onSaveResponse()
			}
			return nil
		case widget.keybindings.Matches(keymap.SaveFullResponse, event):
			if widget.onSaveFullResponse != nil {
				widget.onSaveFullResponse()
			}
			return nil
		}
		return event
	}
}

func (widget *Producer) updateSearch() {
	element := widget.Element.(*tview.TextView)
	widget.updateTitle()
	if widget.searchQuery == "" {
		return
	}
	index := strings.Index(strings.ToLower(widget.currentText), strings.ToLower(widget.searchQuery))
	if index < 0 {
		return
	}
	row := strings.Count(widget.currentText[:index], "\n")
	element.ScrollTo(row, 0)
}

func (widget *Producer) updateTitle() {
	element := widget.Element.(*tview.TextView)
	mode := widget.locale.Text("pretty")
	if widget.bodyViewMode == BodyViewRaw {
		mode = widget.locale.Text("raw")
	}
	title := widget.locale.Text("producer") + " [" + mode + "]"
	if widget.historyVisible && len(widget.history) > 0 {
		title += fmt.Sprintf(" %s %d/%d", widget.locale.Text("history"), widget.historyIndex+1, len(widget.history))
	}
	if widget.searchMode || widget.searchQuery != "" {
		title += " /" + widget.searchQuery
	}
	element.SetTitle(title)
}
