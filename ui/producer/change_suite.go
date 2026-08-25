package producer

import (
	"context"
	"sync"
	"time"

	"github.com/Hecatoncheir/lazyrest/locale"
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

func (widget *Producer) formatProgressBar(current, total int64) string {
	return uiprogress.BodyLocalized(current, total, progressBarWidth, progressPulseWidth, widget.locale)
}

func formatIndeterminateProgressBar(frame int) string {
	return uiprogress.Indeterminate(frame, progressBarWidth, progressPulseWidth)
}

func runningRequestText(progress string) string {
	return localizedRunningRequestText(locale.English(), progress)
}

func localizedRunningRequestText(translator *locale.Translator, progress string) string {
	return translator.Text("running_request") + "\n" + progress
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
			text := localizedRunningRequestText(widget.locale, formatIndeterminateProgressBar(frame))
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
	initialText := localizedRunningRequestText(widget.locale, formatIndeterminateProgressBar(0))

	// Show loading state immediately
	element.Clear().
		SetText(initialText).
		SetWrap(true).
		SetTitleColor(widget.theme.TitleFocus).
		SetBorderColor(widget.theme.BorderFocus).
		SetBackgroundColor(widget.theme.BackgroundFocus)
	widget.updateTitle()
	widget.currentText = initialText

	// Set focus so the user sees it working
	element.SetFocusFunc(func() {
		widget.applyTheme(element, true)
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
				widget.setText(localizedRunningRequestText(widget.locale, widget.formatProgressBar(current, total)))
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
				widget.onRunFinished(response, err)
			}
			widget.addHistory(suite, response, err)
			text := renderExecutionResultWithLocale(suite, response, err, widget.bodyViewMode, widget.locale)

			widget.showCompletedResult(element, text)
		})
	}()
}

func (widget *Producer) showCompletedResult(element *tview.TextView, text string) {
	element.Clear().SetWrap(true)
	widget.applyTheme(element, element.HasFocus())
	widget.setText(text)
}
