package producer

import (
	"context"
	"errors"
	"slices"
	"strings"
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
	widget.searchMode = false
	widget.searchQuery = ""
	widget.searchMatches = nil
	widget.searchIndex = -1
	widget.historyDataMutex.Lock()
	widget.historyVisible = false
	widget.resultAvailable = false
	widget.requestAvailable = true
	widget.historyDataMutex.Unlock()
	element := widget.Element.(*tview.TextView)

	// What an earlier request answered is filled in now rather than while the
	// file is read, because it depends on what has been run.
	if unresolved := http.ResolveResponseReferences(&suite, &widget.responses); len(unresolved) > 0 {
		widget.historyDataMutex.Lock()
		widget.suite = cloneRequestSuite(suite)
		widget.historyDataMutex.Unlock()
		widget.CancelActive()
		// The run has to be reported as finished even though nothing was sent,
		// or the footer keeps animating a request that will never arrive.
		if widget.onRunFinished != nil {
			widget.onRunFinished(runner.Response{}, errors.New(strings.Join(unresolved, "; ")))
		}
		widget.showCompletedResult(element, widget.renderUnresolved(unresolved))
		widget.updateTitle()
		return
	}
	widget.historyDataMutex.Lock()
	widget.suite = cloneRequestSuite(suite)
	widget.historyDataMutex.Unlock()

	ctx, runID := widget.StartRun()
	runnerConfig := widget.runnerConfiguration()
	initialText := localizedRunningRequestText(widget.locale, formatIndeterminateProgressBar(0))

	// Show loading state immediately
	element.Clear().
		SetText(initialText).
		SetWrap(true).
		SetTitleColor(widget.theme.TitleFocus).
		SetBorderColor(widget.theme.BorderFocus)
	element.SetBackgroundColor(widget.theme.BackgroundFocus)
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
		r := runner.NewFromSuiteWithConfig(suite, runnerConfig)

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
			suite.SecretValues = append(suite.SecretValues, http.SensitiveResponseValues(response.Body, response.Header)...)
			widget.recordResponse(suite, response, err)
			widget.addHistory(suite, response, err)
			text := widget.renderResult(suite, response, err)

			widget.showCompletedResult(element, text)
		})
	}()
}

// recordResponse keeps the answer of a named request so that the requests
// after it can refer to what it returned.
func (widget *Producer) recordResponse(suite http.HttpSuite, response runner.Response, err error) {
	if err != nil || suite.Name == "" {
		return
	}
	widget.responses.Record(suite, http.ResponseValue{
		Body:   response.Body,
		Header: response.Header,
		Status: redactSecrets(response.Code, suite.SecretValues),
	})
}

func (widget *Producer) renderUnresolved(unresolved []string) string {
	slices.Sort(unresolved)
	var text strings.Builder
	text.WriteString("[red]" + widget.locale.Text("response_error") + ":[white]\n")
	for _, reference := range unresolved {
		text.WriteString(tview.Escape("- " + reference + "\n"))
	}
	return text.String()
}

func (widget *Producer) showCompletedResult(element *tview.TextView, text string) {
	element.Clear().SetWrap(true)
	widget.applyTheme(element, element.HasFocus())
	widget.setText(text)
}
