package producer

import (
	"fmt"
	"lazyrest/parser/http"
	"lazyrest/runner"

	"github.com/rivo/tview"
)

func (widget *Producer) ChangeSuite(suite http.HttpSuite) {
	widget.suite = suite
	element := widget.Element.(*tview.TextView)
	theme := widget.theme

	// Show loading state immediately
	element.Clear().
		SetText("Running request...").
		SetWrap(true).
		SetTitleColor(theme.TitleFocus).
		SetBorderColor(theme.BorderFocus).
		SetBackgroundColor(theme.BackgroundFocus)

	// Set focus so the user sees it working
	element.SetFocusFunc(func() {
		element.SetBackgroundColor(theme.BackgroundFocus)
	})

	// Run in background
	go func() {
		runner := runner.NewFromSuite(suite)
		response, err := runner.Execute()

		// Update UI safely
		widget.app.QueueUpdateDraw(func() {
			var text string
			if err != nil {
				text = fmt.Sprintf("Response error:\n%v\n%v", "error", err.Error())
			} else {
				text = fmt.Sprintf("Response:\nBody:\n%v\n\n%v\n", response.Body, response.ToMiniString())
			}

			element.
				Clear().
				SetText(text).
				SetWrap(true).
				SetTitleColor(theme.TitleFocus).
				SetBorderColor(theme.BorderFocus).
				SetBackgroundColor(theme.BackgroundFocus)

			element.
				SetBackgroundColor(theme.BackgroundFocus)
		})
	}()
}
