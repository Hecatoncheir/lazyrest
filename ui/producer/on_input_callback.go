package producer

import (
	"fmt"
	"strings"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OnEscapeCallbackType func()

// onInputCallback decides what a key press means in the response pane. While
// the pane is searching every key belongs to the query; otherwise the key is
// one of the pane's commands, or none of them and passed on.
func onInputCallback(widget *Producer) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if widget.searchMode {
			widget.searchInput(event)
			return nil
		}
		if widget.command(event) {
			return nil
		}
		return event
	}
}

// searchInput edits the query. A key that is neither text nor the end of the
// search leaves the query as it is, but is still swallowed: the pane is
// searching, not navigating.
func (widget *Producer) searchInput(event *tcell.EventKey) {
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
	widget.updateSearch()
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

// command runs what the key asks of the pane, and reports whether the key
// asked anything at all.
func (widget *Producer) command(event *tcell.EventKey) bool {
	matches := func(action keymap.Action) bool { return widget.keybindings.Matches(action, event) }

	switch {
	case matches(keymap.Back):
		widget.onEscapeCallback()
	case matches(keymap.Search):
		widget.startSearch()
	case matches(keymap.SearchNext):
		widget.moveToSearchMatch(1)
	case matches(keymap.SearchPrevious):
		widget.moveToSearchMatch(-1)
	case matches(keymap.HistoryPrevious):
		widget.showHistory(-1)
	case matches(keymap.HistoryNext):
		widget.showHistory(1)
	case matches(keymap.ToggleBody):
		widget.toggleBodyView()
	case matches(keymap.ToggleHeaders):
		widget.toggleHeaders()
	case matches(keymap.ToggleRequest):
		widget.toggleRequestDetails()
	case matches(keymap.RerunRequest):
		call(widget.onRerunRequest)
	case matches(keymap.CopyResponseBody):
		call(widget.onCopyBody)
	case matches(keymap.CopyResponse):
		call(widget.onCopyResponse)
	case matches(keymap.CopyAsCurl):
		call(widget.onCopyAsCurl)
	case matches(keymap.SaveResponse):
		call(widget.onSaveResponse)
	case matches(keymap.SaveFullResponse):
		call(widget.onSaveFullResponse)
	default:
		return false
	}
	return true
}

func (widget *Producer) startSearch() {
	widget.searchMode = true
	widget.searchQuery = ""
	widget.updateSearch()
}

// call runs a callback the application may have left unset.
func call(callback func()) {
	if callback != nil {
		callback()
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
	title := widget.locale.Text("producer") + " [" + mode
	if widget.showHeaders {
		title += " H+"
	} else {
		title += " H−"
	}
	if widget.showRequest {
		title += " R+"
	} else {
		title += " R−"
	}
	title += "]"
	if widget.focused {
		title = "▶ " + title
	}
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
