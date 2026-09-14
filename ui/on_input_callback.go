package ui

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type onInputCallbackType func(event *tcell.EventKey) *tcell.EventKey

// onInputCallback decides what a key press means. The order below is the whole
// rule: an overlay takes the key first, then a search in progress, then the
// bindings that work anywhere, and finally the ones that move focus between
// panes. Each step is a function of its own so that order stays readable.
func onInputCallback(application *Application) onInputCallbackType {
	return func(event *tcell.EventKey) *tcell.EventKey {
		bindings := application.bindings()

		if overlay, open := application.overlayOnScreen(); open {
			return application.overlayInput(overlay, bindings, event)
		}
		if application.isSearching() {
			application.resetViewportSequence()
			return event
		}
		if application.handleViewportInput(event) {
			return nil
		}
		if application.globalInput(bindings, event) {
			return nil
		}
		return application.focusInput(bindings, event)
	}
}

// bindings falls back to the defaults so the rest of the file never checks for
// a missing configuration.
func (application *Application) bindings() *keymap.Bindings {
	if application.config.Keybindings == nil {
		return keymap.Default()
	}
	return application.config.Keybindings
}

// overlayOnScreen reports the overlay on screen, if there is one.
func (application *Application) overlayOnScreen() (Overlay, bool) {
	if application.Model == nil {
		return OverlayNone, false
	}
	overlay := application.Model.CurrentOverlay()
	return overlay, overlay != OverlayNone
}

func (application *Application) isSearching() bool {
	return application.HttpFilesTree.IsSearching() ||
		application.Suites.IsSearching() ||
		application.Producer.IsSearching()
}

// isTextInputOverlay reports whether the overlay is one the user types into,
// where only keys that cannot be part of text may act globally.
func isTextInputOverlay(overlay Overlay) bool {
	return overlay == OverlaySaveResponse || overlay == OverlaySendFrame
}

// overlayInput gives the key to whatever is on top of the panes.
func (application *Application) overlayInput(
	overlay Overlay,
	bindings *keymap.Bindings,
	event *tcell.EventKey,
) *tcell.EventKey {
	if isTextInputOverlay(overlay) {
		return application.textFieldInput(bindings, event)
	}
	if application.handleClearConfirmation(overlay, event) {
		return nil
	}
	if overlay == OverlayCommandPalette && application.handleCommandPaletteInput(event) {
		return nil
	}
	if application.handleViewportInput(event) {
		return nil
	}
	if application.overlayCommand(overlay, bindings, event) {
		return nil
	}
	return application.overlayMovement(bindings, event)
}

// textFieldInput keeps an overlay that is a text field able to receive what is
// typed. Without this a global binding on a printable key wins over the field:
// `q` closes the overlay, `:` opens the command palette, and a frame or a path
// holding either simply cannot be entered.
func (application *Application) textFieldInput(
	bindings *keymap.Bindings,
	event *tcell.EventKey,
) *tcell.EventKey {
	switch {
	case event.Key() == tcell.KeyCtrlC:
		stopApplication(application)
		return nil
	case bindings.Matches(keymap.Back, event):
		application.closeOverlay()
		return nil
	}
	return event
}

// overlayCommand handles the keys that act on the overlay itself. It reports
// whether the key was one of them.
func (application *Application) overlayCommand(
	overlay Overlay,
	bindings *keymap.Bindings,
	event *tcell.EventKey,
) bool {
	switch {
	case bindings.Matches(keymap.Quit, event):
		// Ctrl+C leaves the application; the letter only leaves the overlay.
		if event.Key() == tcell.KeyCtrlC {
			stopApplication(application)
		} else {
			application.closeOverlay()
		}
	case bindings.Matches(keymap.Back, event):
		application.closeOverlay()
	case bindings.Matches(keymap.CommandPalette, event):
		application.toggleOverlay(overlay, OverlayCommandPalette)
	case bindings.Matches(keymap.Help, event):
		application.toggleOverlay(overlay, OverlayHelp)
	case bindings.Matches(keymap.Diagnostics, event):
		application.toggleOverlay(overlay, OverlayDiagnostics)
	case bindings.Matches(keymap.ReloadConfig, event):
		application.closeOverlay()
		application.reloadConfiguration()
	default:
		return false
	}
	return true
}

// toggleOverlay closes the overlay when its own key is pressed again, and
// switches to another one otherwise.
func (application *Application) toggleOverlay(current, wanted Overlay) {
	if current == wanted {
		application.closeOverlay()
		return
	}
	application.openOverlay(wanted)
}

// overlayMovement turns the configured movement keys into the arrows a tview
// list understands, and leaves every other key to the overlay.
func (application *Application) overlayMovement(
	bindings *keymap.Bindings,
	event *tcell.EventKey,
) *tcell.EventKey {
	switch {
	case bindings.Matches(keymap.MoveDown, event):
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case bindings.Matches(keymap.MoveUp, event):
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	}
	return event
}

// globalInput handles the bindings that work whenever no overlay is open. It
// reports whether the key was one of them.
func (application *Application) globalInput(bindings *keymap.Bindings, event *tcell.EventKey) bool {
	switch {
	case bindings.Matches(keymap.CommandPalette, event):
		application.openOverlay(OverlayCommandPalette)
	case bindings.Matches(keymap.Help, event):
		application.openOverlay(OverlayHelp)
	case bindings.Matches(keymap.Diagnostics, event):
		application.openOverlay(OverlayDiagnostics)
	case bindings.Matches(keymap.ReloadConfig, event):
		application.reloadConfiguration()
	case bindings.Matches(keymap.Quit, event):
		stopApplication(application)
	default:
		return false
	}
	return true
}

// focusInput moves focus between panes, and passes the key on when it is not a
// focus key at all.
func (application *Application) focusInput(
	bindings *keymap.Bindings,
	event *tcell.EventKey,
) *tcell.EventKey {
	if !isFocusKey(bindings, event) {
		return event
	}
	focused := application.Element.GetFocus()
	if focused == nil {
		return event
	}
	if target := application.focusTarget(bindings, event, focused); target != nil {
		application.Element.SetFocus(target)
	}
	return nil
}

func isFocusKey(bindings *keymap.Bindings, event *tcell.EventKey) bool {
	return bindings.Matches(keymap.FocusLeft, event) ||
		bindings.Matches(keymap.FocusDown, event) ||
		bindings.Matches(keymap.FocusUp, event) ||
		bindings.Matches(keymap.FocusRight, event)
}

// focusTarget names the pane a focus key leads to, or nil when the key leads
// nowhere from where focus already is.
func (application *Application) focusTarget(
	bindings *keymap.Bindings,
	event *tcell.EventKey,
	focused tview.Primitive,
) tview.Primitive {
	// The response slot holds Producer's view, or the stream pane while a
	// connection is open. Naming Producer here would focus a primitive that is
	// not on screen.
	response := application.responseElement()

	switch {
	case bindings.Matches(keymap.FocusLeft, event):
		switch focused {
		case application.Suites.Element, application.Suite.Element:
			return application.HttpFilesTree.Element
		case response:
			return application.Suite.Element
		}
	case bindings.Matches(keymap.FocusRight, event):
		switch focused {
		case application.HttpFilesTree.Element:
			return application.Suites.Element
		case application.Suite.Element, application.Suites.Element:
			return response
		}
	case bindings.Matches(keymap.FocusDown, event):
		if focused == application.Suites.Element {
			return application.Suite.Element
		}
	case bindings.Matches(keymap.FocusUp, event):
		if focused == application.Suite.Element {
			return application.Suites.Element
		}
	}
	return nil
}

// stopApplication ends every background task before the screen goes away, so
// nothing is left writing to a stopped application.
func stopApplication(application *Application) {
	application.stopFooterProgress()
	application.Producer.CancelActive()
	application.Suites.CancelLoad()
	application.HttpFilesTree.CancelReload()
	application.stopFileWatcher()
	application.Element.Stop()
}
