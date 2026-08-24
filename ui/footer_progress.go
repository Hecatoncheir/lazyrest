package ui

import (
	"context"
	"time"

	uiprogress "github.com/Hecatoncheir/lazyrest/ui/progress"
)

const (
	footerProgressWidth  = 12
	footerProgressPulse  = 3
	footerProgressPeriod = 100 * time.Millisecond
)

func (application *Application) showFooterProgress(label string) {
	application.footerProgressMutex.Lock()
	if application.footerProgressCancel != nil && application.footerProgressLabel == label {
		application.footerProgressMutex.Unlock()
		return
	}
	if application.footerProgressCancel != nil {
		application.footerProgressCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	application.footerProgressCancel = cancel
	application.footerProgressLabel = label
	application.footerProgressMutex.Unlock()

	application.Footer.UpdateStatus(label + " " + uiprogress.Indeterminate(0, footerProgressWidth, footerProgressPulse))
	go func() {
		ticker := time.NewTicker(footerProgressPeriod)
		defer ticker.Stop()
		frame := 1
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				status := label + " " + uiprogress.Indeterminate(frame, footerProgressWidth, footerProgressPulse)
				frame++
				application.Element.QueueUpdateDraw(func() {
					select {
					case <-ctx.Done():
						return
					default:
					}
					application.Footer.UpdateStatus(status)
				})
			}
		}
	}()
}

func (application *Application) stopFooterProgress() {
	application.footerProgressMutex.Lock()
	defer application.footerProgressMutex.Unlock()
	if application.footerProgressCancel != nil {
		application.footerProgressCancel()
		application.footerProgressCancel = nil
	}
	application.footerProgressLabel = ""
}
