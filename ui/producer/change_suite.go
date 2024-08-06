package producer

import (
	"fmt"
	"lazyrest/parser/http"
	"lazyrest/runner"

	"github.com/rivo/tview"
)

func (widget *Producer) ChangeSuite(suite http.HttpSuite) {
	text := ""

	runner := runner.NewFromSuite(suite)
	response, err := runner.Execute()
	if err != nil {
		text = fmt.Sprintf(
			"Response error:\n" +
				"%v\n" +
				err.Error(),
		)
	} else {
		text = fmt.Sprintf(
			"Response:\n"+
				"Body:\n"+
				"%v\n\n"+
				"%v\n",
			response.Body,
			response.ToMiniString(),
		)
	}

	widget.suite = suite
	element := widget.Element.(*tview.TextView)
	theme := widget.theme
	element.
		Clear().
		SetText(text).
		SetWrap(true).
		SetTitleColor(theme.TitleFocus).
		SetBorderColor(theme.BorderFocus).
		SetBackgroundColor(theme.BackgroundFocus).
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
