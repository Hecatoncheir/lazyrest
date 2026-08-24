package producer

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OnEscapeCallbackType func()

func onInputCallback(widget *Producer) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if widget.searchMode {
			switch event.Key() {
			case tcell.KeyEsc, tcell.KeyEnter:
				widget.searchMode = false
				widget.updateSearch()
				return nil
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

		switch event.Key() {
		case tcell.KeyEsc:
			widget.onEscapeCallback()
			return nil
		}
		switch event.Rune() {
		case '/':
			widget.searchMode = true
			widget.searchQuery = ""
			widget.updateSearch()
			return nil
		case '[':
			widget.showHistory(-1)
			return nil
		case ']':
			widget.showHistory(1)
			return nil
		}
		return event
	}
}

func (widget *Producer) updateSearch() {
	element := widget.Element.(*tview.TextView)
	title := "Producer"
	if widget.searchMode || widget.searchQuery != "" {
		title += " /" + widget.searchQuery
	}
	element.SetTitle(title)
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
