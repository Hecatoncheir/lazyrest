package suites

import (
	"strings"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"
	"github.com/rivo/tview"
)

// bodyPreview renders the body of a request as one highlighted line, since a
// list row cannot show more than that.
func (widget *Suites) bodyPreview(suite http.HttpSuite) string {
	body := strings.Join(strings.Fields(suite.Redact(suite.Body)), " ")
	if body == "" {
		return ""
	}
	return syntax.Highlight(body, syntax.LanguageForBodyType(suite.BodyType), widget.syntax)
}

// listRow holds both forms of a row: one with the method coloured, and a plain
// one for the row carrying the selection, whose own style would otherwise be
// painted under the colour.
type listRow struct {
	markup string
	plain  string
}

func (widget *Suites) rowFor(suite http.HttpSuite) listRow {
	name := suite.Redact(suite.Name)
	if name == "" {
		name = suite.Redact(suite.Uri)
	}
	// An unnamed request already reads as "METHOD uri", so the method is not
	// put in front of it twice.
	name = strings.TrimPrefix(name, suite.Method+" ")
	if suite.Method == "" {
		escaped := tview.Escape(name)
		return listRow{markup: escaped, plain: escaped}
	}
	return listRow{
		markup: syntax.Method(suite.Method, widget.methods) + " " + tview.Escape(name),
		plain:  tview.Escape(strings.TrimSpace(suite.Method + " " + name)),
	}
}

// applySelectionMarkup drops the markup from the row that carries the
// selection. The list paints its selection style under the text, so a coloured
// method would keep its own colour against that background.
func (widget *Suites) applySelectionMarkup(current int) {
	element, ok := widget.Element.(*tview.List)
	if !ok {
		return
	}
	for index, row := range widget.rows {
		if index >= element.GetItemCount() {
			break
		}
		text := row.markup
		if index == current {
			text = row.plain
		}
		_, secondary := element.GetItemText(index)
		element.SetItemText(index, text, secondary)
	}
}

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
	widget.rows = widget.rows[:0]
	visibleSuites := 0
	for _, suite := range widget.suites {
		searchable := strings.ToLower(strings.Join([]string{suite.Name, suite.Method, suite.Uri, suite.Body}, "\n"))
		if query != "" && !strings.Contains(searchable, query) {
			continue
		}
		visibleSuites++

		theme := widget.theme
		row := widget.rowFor(suite)
		widget.rows = append(widget.rows, row)
		element.AddItem(row.markup, widget.bodyPreview(suite), 0, func() {
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
	widget.applySelectionMarkup(element.GetCurrentItem())
}
