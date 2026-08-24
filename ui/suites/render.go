package suites

import (
	"strings"

	"github.com/rivo/tview"
)

func (widget *Suites) render() {
	element := widget.Element.(*tview.List)
	element.Clear()

	title := "Suites"
	if widget.searchMode || widget.searchQuery != "" {
		title += " /" + widget.searchQuery
	}
	element.SetTitle(title)

	for _, diagnostic := range widget.diagnostics {
		element.AddItem("Warning: "+diagnostic.String(), "", 0, nil)
	}

	query := strings.ToLower(widget.searchQuery)
	for _, suite := range widget.suites {
		searchable := strings.ToLower(strings.Join([]string{suite.Name, suite.Method, suite.Uri, suite.Body}, "\n"))
		if query != "" && !strings.Contains(searchable, query) {
			continue
		}

		theme := widget.theme
		label := suite.Name
		if label == "" {
			label = suite.Method + " " + suite.Uri
		}
		element.AddItem(label, suite.Body, 0, func() {
			widget.onSuiteSelectCallback(suite)
		}).
			SetWrapAround(true).
			SetHighlightFullLine(true).
			SetMainTextColor(theme.SuiteForeground).
			SetSecondaryTextColor(theme.SuiteForeground).
			SetSelectedTextColor(theme.SuiteFocusForeground).
			SetSelectedBackgroundColor(theme.SuiteFocusBackground).
			SetBackgroundColor(theme.SuiteBackground)
	}
}
