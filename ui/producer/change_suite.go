package producer

import (
	"fmt"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"

	"github.com/rivo/tview"
)

func formatProgressBar(current, total int64) string {
	if total <= 0 {
		return fmt.Sprintf("Loading... %d bytes", current)
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
	widget.searchMode = false
	widget.searchQuery = ""
	widget.historyVisible = false
	ctx, runID := widget.StartRun()
	element := widget.Element.(*tview.TextView)
	theme := widget.theme

	// Show loading state immediately
	element.Clear().
		SetText("Running request...").
		SetWrap(true).
		SetTitleColor(theme.TitleFocus).
		SetBorderColor(theme.BorderFocus).
		SetBackgroundColor(theme.BackgroundFocus)
	widget.updateTitle()
	widget.currentText = "Running request..."

	// Set focus so the user sees it working
	element.SetFocusFunc(func() {
		element.SetBackgroundColor(theme.BackgroundFocus)
	})

	// Run in background
	go func() {
		r := runner.NewFromSuiteWithConfig(suite, widget.runnerConfig)

		// Start executing with progress callback
		response, err := r.Execute(ctx, func(current, total int64) {
			widget.app.QueueUpdateDraw(func() {
				if !widget.IsCurrentRun(runID) {
					return
				}
				widget.setText(fmt.Sprintf("Running request...\n%s", formatProgressBar(current, total)))
			})
		})

		// Update UI safely
		widget.app.QueueUpdateDraw(func() {
			if !widget.FinishRun(runID) {
				return
			}
			widget.addHistory(suite, response, err)
			text := renderExecutionResultWithMode(suite, response, err, widget.bodyViewMode)

			element.
				Clear().
				SetWrap(true).
				SetTitleColor(theme.TitleFocus).
				SetBorderColor(theme.BorderFocus).
				SetBackgroundColor(theme.BackgroundFocus)

			element.
				SetBackgroundColor(theme.BackgroundFocus)
			widget.setText(text)
		})
	}()
}
