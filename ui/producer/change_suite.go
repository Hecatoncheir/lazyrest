package producer

import (
	"fmt"
	"lazyrest/parser/http"
	"lazyrest/runner"

	"github.com/rivo/tview"
)

func formatProgressBar(current, total int64) string {
	if total <= 0 {
		return "Loading..."
	}
	percentage := float64(current) / float64(total) * 100
	width := 20
	filled := int(percentage / (100.0 / float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += "-"
		}
	}
	bar += fmt.Sprintf("] %.0f%%", percentage)
	return bar
}

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
		r := runner.NewFromSuite(suite)
		
		// Start executing with progress callback
		response, err := r.Execute(func(current, total int64) {
			widget.app.QueueUpdateDraw(func() {
				element.SetText(fmt.Sprintf("Running request...\n%s", formatProgressBar(current, total)))
			})
		})

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
