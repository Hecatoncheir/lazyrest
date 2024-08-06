package suites

import (
	"lazyrest/parser/http"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func New() Suites {
	element := Suites{}
	return element
}

type Suites struct {
	Element               tview.Primitive
	theme                 theme.SuitesTheme
	suites                []http.HttpSuite
	selectedSuite         http.HttpSuite
	onEscapeCallback      OnEscapeCallbackType
	onSuiteSelectCallback OnSuiteSelectCallbackType
}

func (widget *Suites) Build(parameters Parameters) tview.Primitive {
	widget.onEscapeCallback = parameters.OnEscapeCallback
	widget.onSuiteSelectCallback = parameters.OnSuiteSelectCallbackType

	theme := parameters.Theme.Suites
	widget.theme = theme

	element := tview.NewList()

	box := tview.NewBox().
		SetBorder(true).
		SetBackgroundColor(theme.Background).
		SetBorderColor(theme.Border).
		SetTitle("Suites").
		SetInputCapture(onInputCallback(widget))

	box.SetFocusFunc(func() {
		box.
			SetTitleColor(theme.TitleFocus).
			SetBackgroundColor(theme.BackgroundFocus).
			SetBorderColor(theme.BorderFocus)
	})

	box.SetBlurFunc(func() {
		box.
			SetTitleColor(theme.Title).
			SetBackgroundColor(theme.Background).
			SetBorderColor(theme.Border)
	})

	element.Box = box

	widget.Element = element
	return element
}
