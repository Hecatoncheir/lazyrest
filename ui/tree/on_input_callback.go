package tree

import (
	"fmt"
	"strings"

	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/keymap"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// onInputCallback decides what a key press means in the file tree. While the
// tree is searching every key belongs to the query; otherwise the key opens a
// file, runs one of the tree's commands, or is passed on.
func onInputCallback(widget *Tree) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if widget.searchMode {
			widget.searchInput(event)
			return nil
		}
		if widget.keybindings.Matches(keymap.Open, event) {
			// tview opens a node on Enter, so the configured key is translated
			// into the one the widget already understands.
			return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		}
		if event.Key() == tcell.KeyEnter {
			// A bare Enter that was not the configured key opens nothing.
			return nil
		}
		if widget.command(event) {
			return nil
		}
		return event
	}
}

// searchInput edits the query. A key that is neither text nor the end of the
// search leaves the query alone, but is still swallowed: the tree is
// searching, not navigating.
func (widget *Tree) searchInput(event *tcell.EventKey) {
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

// command runs what the key asks of the tree, and reports whether the key asked
// anything at all.
func (widget *Tree) command(event *tcell.EventKey) bool {
	switch {
	case widget.keybindings.Matches(keymap.Search, event):
		widget.searchMode = true
		widget.searchQuery = ""
		widget.updateSearch()
	case widget.keybindings.Matches(keymap.SearchNext, event):
		widget.moveToMatch(1)
	case widget.keybindings.Matches(keymap.SearchPrevious, event):
		widget.moveToMatch(-1)
	case widget.keybindings.Matches(keymap.Reload, event):
		if widget.onReloadCallback != nil {
			widget.onReloadCallback()
		}
	default:
		return false
	}
	return true
}

func (widget *Tree) updateSearch() {
	element, ok := widget.Element.(*tview.TreeView)
	if !ok {
		return
	}
	widget.searchMatches = nil
	widget.searchIndex = 0
	if widget.searchQuery == "" {
		widget.updateTitle()
		return
	}
	query := strings.ToLower(widget.searchQuery)
	collectMatchingNodes(element.GetRoot(), query, &widget.searchMatches)
	widget.updateTitle()
	if len(widget.searchMatches) > 0 {
		element.SetCurrentNode(widget.searchMatches[0])
	}
}

func (widget *Tree) updateTitle() {
	element, ok := widget.Element.(*tview.TreeView)
	if !ok {
		return
	}
	title := widget.locale.Text("files")
	if widget.focused {
		title = "▶ " + title
	}
	if widget.loading {
		title += " — " + widget.locale.Text("loading")
	} else if widget.reloading {
		title += " — " + widget.locale.Text("reloading")
	}
	if widget.searchMode || widget.searchQuery != "" {
		title += " /" + widget.searchQuery
		if widget.searchQuery != "" {
			current := 0
			if len(widget.searchMatches) > 0 {
				current = widget.searchIndex + 1
			}
			title += fmt.Sprintf(" [%d/%d]", current, len(widget.searchMatches))
		}
	}
	element.SetTitle(title)
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
