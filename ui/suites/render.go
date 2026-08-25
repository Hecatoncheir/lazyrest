package suites

import (
	"strings"

	"github.com/rivo/tview"
)

func (widget *Suites) render() {
	element := widget.Element.(*tview.List)
	element.Clear()

	title := widget.locale.Text("suites")
	if widget.searchMode || widget.searchQuery != "" {
		title += " /" + widget.searchQuery
	}
	if widget.diagnosticCount > 0 {
		title += " — " + widget.locale.PluralDiagnostics(widget.diagnosticCount) + " [d]"
	}
	element.SetTitle(title)

	query := strings.ToLower(widget.searchQuery)
	visibleSuites := 0
	for _, suite := range widget.suites {
		searchable := strings.ToLower(strings.Join([]string{suite.Name, suite.Method, suite.Uri, suite.Body}, "\n"))
		if query != "" && !strings.Contains(searchable, query) {
			continue
		}
		visibleSuites++

		theme := widget.theme
		label := suite.Redact(suite.Name)
		if label == "" {
			label = suite.Method + " " + suite.Redact(suite.Uri)
		}
		element.AddItem(label, suite.Redact(suite.Body), 0, func() {
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
	if visibleSuites == 0 && widget.diagnosticCount > 0 && query == "" {
		element.AddItem(widget.locale.Text("no_requests"), "", 0, nil)
	}
}
