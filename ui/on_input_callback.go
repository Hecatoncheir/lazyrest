package ui

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type onInputCallbackType func(event *tcell.EventKey) *tcell.EventKey

func onInputCallback(application *Application) onInputCallbackType {
	applicationElement := application.Element
	return func(event *tcell.EventKey) *tcell.EventKey {
		bindings := application.config.Keybindings
		if bindings == nil {
			bindings = keymap.Default()
		}
		if application.Model != nil {
			overlay := application.Model.CurrentOverlay()
			if overlay != OverlayNone {
				if application.handleViewportInput(event) {
					return nil
				}
				switch {
				case bindings.Matches(keymap.Quit, event):
					if event.Key() == tcell.KeyCtrlC {
						stopApplication(application)
					} else {
						application.closeOverlay()
					}
					return nil
				case bindings.Matches(keymap.Back, event):
					application.closeOverlay()
					return nil
				case bindings.Matches(keymap.CommandPalette, event):
					if overlay == OverlayCommandPalette {
						application.closeOverlay()
					} else {
						application.openOverlay(OverlayCommandPalette)
					}
					return nil
				case bindings.Matches(keymap.ReloadConfig, event):
					application.closeOverlay()
					application.reloadConfiguration()
					return nil
				case bindings.Matches(keymap.Help, event):
					if overlay == OverlayHelp {
						application.closeOverlay()
					} else {
						application.openOverlay(OverlayHelp)
					}
					return nil
				case bindings.Matches(keymap.Diagnostics, event):
					if overlay == OverlayDiagnostics {
						application.closeOverlay()
					} else {
						application.openOverlay(OverlayDiagnostics)
					}
					return nil
				case bindings.Matches(keymap.MoveDown, event):
					return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
				case bindings.Matches(keymap.MoveUp, event):
					return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
				}
				return event
			}
		}
		if application.HttpFilesTree.IsSearching() || application.Suites.IsSearching() || application.Producer.IsSearching() {
			application.resetViewportSequence()
			return event
		}
		if application.handleViewportInput(event) {
			return nil
		}
		switch {
		case bindings.Matches(keymap.CommandPalette, event):
			application.openOverlay(OverlayCommandPalette)
			return nil
		case bindings.Matches(keymap.ReloadConfig, event):
			application.reloadConfiguration()
			return nil
		case bindings.Matches(keymap.Help, event):
			application.openOverlay(OverlayHelp)
			return nil
		case bindings.Matches(keymap.Diagnostics, event):
			application.openOverlay(OverlayDiagnostics)
			return nil
		}
		// handle 'q' to quit
		if bindings.Matches(keymap.Quit, event) {
			stopApplication(application)
			return nil
		}

		if bindings.Matches(keymap.FocusLeft, event) || bindings.Matches(keymap.FocusDown, event) ||
			bindings.Matches(keymap.FocusUp, event) || bindings.Matches(keymap.FocusRight, event) {
			focused := applicationElement.GetFocus()
			if focused == nil {
				return event
			}

			var target tview.Primitive
			switch {
			case bindings.Matches(keymap.FocusLeft, event):
				switch focused {
				case application.Suites.Element, application.Suite.Element:
					target = application.HttpFilesTree.Element
				case application.Producer.Element:
					target = application.Suite.Element
				}
			case bindings.Matches(keymap.FocusDown, event):
				if focused == application.Suites.Element {
					target = application.Suite.Element
				}
			case bindings.Matches(keymap.FocusUp, event):
				if focused == application.Suite.Element {
					target = application.Suites.Element
				}
			case bindings.Matches(keymap.FocusRight, event):
				switch focused {
				case application.HttpFilesTree.Element:
					target = application.Suites.Element
				case application.Suite.Element, application.Suites.Element:
					target = application.Producer.Element
				}
			}

			if target != nil {
				applicationElement.SetFocus(target)
			}
			return nil
		}

		return event
	}
}

func stopApplication(application *Application) {
	application.stopFooterProgress()
	application.Producer.CancelActive()
	application.Suites.CancelLoad()
	application.HttpFilesTree.CancelReload()
	application.Element.Stop()
}
