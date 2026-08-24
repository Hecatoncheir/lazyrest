package producer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"

	"github.com/rivo/tview"
)

const (
	progressBarWidth    = 20
	progressPulseWidth  = 5
	progressFramePeriod = 100 * time.Millisecond
)

func formatProgressBar(current, total int64) string {
	if total <= 0 {
		frame := int(current / 1024)
		return fmt.Sprintf("%s %d bytes", formatIndeterminateProgressBar(frame), current)
	}
	percentage := float64(current) / float64(total) * 100
	if percentage > 100 {
		percentage = 100
	}
	if percentage < 0 {
		percentage = 0
	}
	filled := int(percentage / 100 * progressBarWidth)
	bar := strings.Repeat("=", filled) + strings.Repeat("-", progressBarWidth-filled)
	return fmt.Sprintf("[%s] %.0f%%", bar, percentage)
}

func formatIndeterminateProgressBar(frame int) string {
	travel := progressBarWidth - progressPulseWidth
	cycle := travel * 2
	position := frame % cycle
	forward := position <= travel
	if !forward {
		position = cycle - position
	}

	bar := []byte(strings.Repeat("-", progressBarWidth))
	for index := range progressPulseWidth {
		bar[position+index] = '='
	}
	if forward {
		bar[position+progressPulseWidth-1] = '>'
	} else {
		bar[position] = '<'
	}
	return "[" + string(bar) + "]"
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
