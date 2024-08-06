package producer

import (
	"lazyrest/parser/http"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func New() Producer {
	widget := Producer{}
	return widget
}

type Producer struct {
	Element          tview.Primitive
	theme            theme.ProducerTheme
	suite            http.HttpSuite
	onEscapeCallback OnEscapeCallbackType
}

func (widget *Producer) Build(parameters Parameters) tview.Primitive {
	widget.onEscapeCallback = parameters.OnEscapeCallback
	theme := parameters.Theme.Producer
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
		SetTitle("Producer").
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
