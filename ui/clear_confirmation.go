package ui

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
)

func (application *Application) requestHistoryClear() {
	count := len(application.Producer.HistorySummaries())
	if count == 0 {
		return
	}
	if application.confirmHistoryClear {
		application.clearHistory()
		return
	}
	application.confirmHistoryClear = true
	application.refreshHistory()
	application.Footer.UpdateStatus(application.config.Locale.Format(
		"confirm_clear_history",
		application.config.Keybindings.Describe(keymap.ClearHistory),
		count,
		application.config.Keybindings.Describe(keymap.Back),
	))
}

func (application *Application) requestCapturedClear() {
	count := len(application.Producer.CapturedResponses())
	if count == 0 {
		return
	}
	if application.confirmCapturedClear {
		application.clearCapturedResponses()
		return
	}
	application.confirmCapturedClear = true
	application.refreshCapturedResponses()
	application.Footer.UpdateStatus(application.config.Locale.Format(
		"confirm_clear_captured",
		application.config.Keybindings.Describe(keymap.ClearCaptured),
		count,
		application.config.Keybindings.Describe(keymap.Back),
	))
}

func (application *Application) handleClearConfirmation(overlay Overlay, event *tcell.EventKey) bool {
	bindings := application.config.Keybindings
	if bindings == nil {
		bindings = keymap.Default()
	}
	if !bindings.Matches(keymap.Back, event) {
		return false
	}
	switch {
	case overlay == OverlayHistory && application.confirmHistoryClear:
		application.confirmHistoryClear = false
		application.refreshHistory()
	case overlay == OverlayCaptured && application.confirmCapturedClear:
		application.confirmCapturedClear = false
		application.refreshCapturedResponses()
	default:
		return false
	}
	application.Footer.UpdateStatus(application.config.Locale.Text("clear_cancelled"))
	return true
}

func (application *Application) resetClearConfirmations() {
	application.confirmHistoryClear = false
	application.confirmCapturedClear = false
}
