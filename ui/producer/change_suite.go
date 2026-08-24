package producer

import (
	"context"
	"sync"
	"time"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	uiprogress "github.com/Hecatoncheir/lazyrest/ui/progress"

	"github.com/rivo/tview"
)

const (
	progressBarWidth    = 20
	progressPulseWidth  = 5
	progressFramePeriod = 100 * time.Millisecond
)

func formatProgressBar(current, total int64) string {
	return uiprogress.Body(current, total, progressBarWidth, progressPulseWidth)
}

func formatIndeterminateProgressBar(frame int) string {
	return uiprogress.Indeterminate(frame, progressBarWidth, progressPulseWidth)
}

func runningRequestText(progress string) string {
	return "Running request...\n" + progress
}

func (widget *Producer) animateProgress(ctx context.Context, runID uint64, done <-chan struct{}) {
	ticker := time.NewTicker(progressFramePeriod)
	defer ticker.Stop()

	frame := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			text := runningRequestText(formatIndeterminateProgressBar(frame))
			frame++
			widget.app.QueueUpdateDraw(func() {
				select {
				case <-done:
					return
				default:
				}
				if !widget.IsCurrentRun(runID) || !widget.IsRunning() {
					return
				}
				widget.setText(text)
			})
		}
	}
}

func (widget *Producer) ChangeSuite(suite http.HttpSuite) {
	widget.suite = suite
	widget.searchMode = false
	widget.searchQuery = ""
	widget.historyVisible = false
	ctx, runID := widget.StartRun()
	element := widget.Element.(*tview.TextView)
	theme := widget.theme
	initialText := runningRequestText(formatIndeterminateProgressBar(0))

	// Show loading state immediately
	element.Clear().
		SetText(initialText).
		SetWrap(true).
		SetTitleColor(theme.TitleFocus).
		SetBorderColor(theme.BorderFocus).
		SetBackgroundColor(theme.BackgroundFocus)
	widget.updateTitle()
	widget.currentText = initialText

	// Set focus so the user sees it working
	element.SetFocusFunc(func() {
		element.SetBackgroundColor(theme.BackgroundFocus)
	})

	animationDone := make(chan struct{})
	var stopAnimation sync.Once
	stopProgressAnimation := func() {
		stopAnimation.Do(func() {
			close(animationDone)
		})
	}
	go widget.animateProgress(ctx, runID, animationDone)

	// Run in background
	go func() {
		r := runner.NewFromSuiteWithConfig(suite, widget.runnerConfig)

		// Start executing with progress callback
		response, err := r.Execute(ctx, func(current, total int64) {
			stopProgressAnimation()
			widget.app.QueueUpdateDraw(func() {
				if !widget.IsCurrentRun(runID) {
					return
				}
				widget.setText(runningRequestText(formatProgressBar(current, total)))
				if widget.onProgress != nil {
					widget.onProgress(current, total)
				}
			})
		})
		stopProgressAnimation()

		// Update UI safely
		widget.app.QueueUpdateDraw(func() {
			if !widget.FinishRun(runID) {
				return
			}
			if widget.onRunFinished != nil {
				widget.onRunFinished(err)
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
