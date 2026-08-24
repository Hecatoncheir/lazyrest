package tree

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func onInputCallback(widget *Tree) func(event *tcell.EventKey) *tcell.EventKey {
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

		switch event.Rune() {
		case '/':
			widget.searchMode = true
			widget.searchQuery = ""
			widget.updateSearch()
			return nil
		case 'n':
			widget.moveToMatch(1)
			return nil
		case 'N':
			widget.moveToMatch(-1)
			return nil
		}
		return event
	}
}

func (widget *Tree) updateSearch() {
	element, ok := widget.Element.(*tview.TreeView)
	if !ok {
		return
	}
	title := "Files"
	if widget.searchMode || widget.searchQuery != "" {
		title += " /" + widget.searchQuery
	}
	element.SetTitle(title)
	widget.searchMatches = nil
	widget.searchIndex = 0
	if widget.searchQuery == "" {
		return
	}
	query := strings.ToLower(widget.searchQuery)
	collectMatchingNodes(element.GetRoot(), query, &widget.searchMatches)
	if len(widget.searchMatches) > 0 {
		element.SetCurrentNode(widget.searchMatches[0])
	}
}

func collectMatchingNodes(node *tview.TreeNode, query string, matches *[]*tview.TreeNode) bool {
	matched := strings.Contains(strings.ToLower(node.GetText()), query)
	for _, child := range node.GetChildren() {
		if collectMatchingNodes(child, query, matches) {
			matched = true
			node.SetExpanded(true)
		}
	}
	if matched {
		if _, ok := node.GetReference().(finder.File); ok {
			*matches = append(*matches, node)
		}
	}
	return matched
}

func (widget *Tree) moveToMatch(delta int) {
	if len(widget.searchMatches) == 0 {
		return
	}
	widget.searchIndex = (widget.searchIndex + delta + len(widget.searchMatches)) % len(widget.searchMatches)
	widget.Element.(*tview.TreeView).SetCurrentNode(widget.searchMatches[widget.searchIndex])
}
