package suite

import (
	"lazyrest/parser/http"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func New() Suite {
	widget := Suite{}
	return widget
}

type Suite struct {
	Element          tview.Primitive
	theme            theme.SuiteTheme
	suite            http.HttpSuite
	onEscapeCallback OnEscapeCallbackType
	onRunCallback    OnRunCallbackType
}

func (widget *Suite) Build(parameters Parameters) tview.Primitive {
	widget.onEscapeCallback = parameters.OnEscapeCallback
	widget.onRunCallback = parameters.OnRunCallback

	theme := parameters.Theme.Suite
	widget.theme = theme

	element := tview.NewTextView()
	element.
		SetTextColor(theme.Foreground).
		SetBackgroundColor(theme.Background)

	element.SetFocusFunc(func() {
		element.
			SetTitleColor(theme.TitleFocus).
			SetBackgroundColor(theme.BackgroundFocus).
			SetBorderColor(theme.BorderFocus)
	})

	element.SetBlurFunc(func() {
		element.
			SetTitleColor(theme.Title).
			SetBackgroundColor(theme.Background).
			SetBorderColor(theme.Border)
	})

	box := tview.NewBox().
		SetBorder(true).
		SetBorderColor(theme.Border).
		SetBackgroundColor(theme.Background).
		SetTitle("Suite").
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
