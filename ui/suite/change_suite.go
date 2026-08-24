package suite

import (
	"fmt"
	"github.com/Hecatoncheir/lazyrest/parser/http"

	"github.com/rivo/tview"
)

func (widget *Suite) ChangeSuite(suite http.HttpSuite) {
	widget.suite = suite
	element := widget.Element.(*tview.TextView)
	text := fmt.Sprintf(
		"Request: %v %v\n"+
			"Body(%v): %v",
		suite.Method,
		suite.Redact(suite.Uri),
		suite.BodyType,
		suite.Redact(suite.Body),
	)
	theme := widget.theme
	element.
		Clear().
		SetText(text).
		SetWrap(true).
		SetTitleColor(theme.TitleFocus).
		SetBorderColor(theme.BorderFocus).
		SetInputCapture(onInputCallback(widget))

	element.
		SetBackgroundColor(theme.BackgroundFocus)

	element.SetFocusFunc(func() {
		element.
			SetBackgroundColor(theme.BackgroundFocus)
	})

	element.SetBlurFunc(func() {
		element.
			SetBackgroundColor(theme.Background)
	})
}

func (widget *Suite) Clear() {
	widget.suite = http.HttpSuite{}
	if widget.Element == nil {
		return
	}
	widget.Element.(*tview.TextView).Clear()
}
