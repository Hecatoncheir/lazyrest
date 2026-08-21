package producer

import (
	"fmt"
	"lazyrest/parser/http"
	"lazyrest/runner"
	"strings"

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
				text = fmt.Sprintf("[red]Response error:[white]\n%v", err.Error())
			} else {
				// Format Request
				var reqBuilder strings.Builder
				reqBuilder.WriteString("[yellow]Request:[white]\n")
				reqBuilder.WriteString(fmt.Sprintf("%s %s\n", suite.Method, suite.Uri))
				for k, v := range suite.Header {
					reqBuilder.WriteString(fmt.Sprintf("%s: %s\n", k, v))
				}
				if suite.Body != "" {
					reqBuilder.WriteString("\n[yellow]Body:[white]\n")
					reqBuilder.WriteString(suite.Body)
				}

				// Format Response
				var respColor string
				if strings.HasPrefix(response.Code, "2") {
					respColor = "green"
				} else if strings.HasPrefix(response.Code, "3") {
					respColor = "yellow"
				} else if strings.HasPrefix(response.Code, "4") || strings.HasPrefix(response.Code, "5") {
					respColor = "red"
				} else {
					respColor = "white"
				}

				reqPart := reqBuilder.String()
				sep := "\n" + strings.Repeat("─", 40) + "\n" // Using unicode dash for better look
				respPart := fmt.Sprintf("[%s]Response:[white]\n%s\n\n%s",
					respColor, response.ToMiniString(), response.Body)

				text = reqPart + sep + respPart
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
