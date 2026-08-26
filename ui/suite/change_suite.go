package suite

import (
	"strings"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"

	"github.com/rivo/tview"
)

func (widget *Suite) ChangeSuite(suite http.HttpSuite) {
	widget.suite = suite
	element := widget.Element.(*tview.TextView)
	text := widget.render(suite)
	theme := widget.theme
	element.
		Clear().
		SetText(text).
		SetWrap(true).
		SetInputCapture(onInputCallback(widget))
	applySuiteTheme(element, theme, element.HasFocus())
	element.SetFocusFunc(func() { applySuiteTheme(element, widget.theme, true) })
	element.SetBlurFunc(func() { applySuiteTheme(element, widget.theme, false) })
}

// render draws the request, colouring the body by the format it declares. The
// text is markup, so every part of it has to be escaped or highlighted.
func (widget *Suite) render(suite http.HttpSuite) string {
	var out strings.Builder
	out.WriteString(widget.locale.Text("request") + ": ")
	out.WriteString(tview.Escape(strings.TrimSpace(suite.Method + " " + suite.Redact(suite.Uri))))

	body := suite.Redact(suite.Body)
	if body == "" {
		return out.String()
	}

	label := widget.locale.Text("body")
	if suite.BodyType != "" {
		label += "(" + suite.BodyType + ")"
	}
	out.WriteString("\n" + tview.Escape(label) + ":\n")
	out.WriteString(syntax.Highlight(body, syntax.LanguageForBodyType(suite.BodyType), widget.syntax))
	return out.String()
}

func (widget *Suite) Clear() {
	widget.suite = http.HttpSuite{}
	if widget.Element == nil {
		return
	}
	widget.Element.(*tview.TextView).Clear()
}
