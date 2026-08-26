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
		case widget.keybindings.Matches(keymap.SearchNext, event):
			widget.moveToSearchMatch(1)
			return nil
		case widget.keybindings.Matches(keymap.SearchPrevious, event):
			widget.moveToSearchMatch(-1)
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
	widget.rebuildSearchMatches()
	widget.scrollToSearchMatch()
	widget.updateTitle()
}

func (widget *Producer) rebuildSearchMatches() {
	widget.searchMatches = nil
	widget.searchIndex = -1
	query := strings.ToLower(widget.searchQuery)
	if query == "" {
		return
	}
	text := strings.ToLower(widget.currentText)
	row := 0
	for offset := 0; offset <= len(text)-len(query); {
		index := strings.Index(text[offset:], query)
		if index < 0 {
			break
		}
		index += offset
		row += strings.Count(text[offset:index], "\n")
		widget.searchMatches = append(widget.searchMatches, row)
		nextOffset := index + len(query)
		row += strings.Count(text[index:nextOffset], "\n")
		offset = nextOffset
	}
	if len(widget.searchMatches) > 0 {
		widget.searchIndex = 0
	}
}

func (widget *Producer) moveToSearchMatch(delta int) {
	if len(widget.searchMatches) == 0 {
		return
	}
	widget.searchIndex = (widget.searchIndex + delta + len(widget.searchMatches)) % len(widget.searchMatches)
	widget.scrollToSearchMatch()
	widget.updateTitle()
}

func (widget *Producer) scrollToSearchMatch() {
	if widget.searchIndex < 0 || widget.searchIndex >= len(widget.searchMatches) {
		return
	}
	widget.Element.(*tview.TextView).ScrollTo(widget.searchMatches[widget.searchIndex], 0)
}

func (widget *Producer) updateTitle() {
	element := widget.Element.(*tview.TextView)
	mode := widget.locale.Text("pretty")
	if widget.bodyViewMode == BodyViewRaw {
		mode = widget.locale.Text("raw")
	}
	title := widget.locale.Text("producer") + " [" + mode + "]"
	widget.historyDataMutex.RLock()
	historyVisible := widget.historyVisible
	historyIndex := widget.historyIndex
	historyLength := len(widget.history)
	widget.historyDataMutex.RUnlock()
	if historyVisible && historyLength > 0 {
		title += fmt.Sprintf(" %s %d/%d", widget.locale.Text("history"), historyIndex+1, historyLength)
	}
	if widget.searchMode || widget.searchQuery != "" {
		title += " /" + widget.searchQuery
		if widget.searchQuery != "" {
			title += fmt.Sprintf(" %d/%d", widget.searchIndex+1, len(widget.searchMatches))
		}
	}
	element.SetTitle(title)
}
