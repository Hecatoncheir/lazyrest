package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type onInputCallbackType func(event *tcell.EventKey) *tcell.EventKey

func onInputCallback(application *Application) onInputCallbackType {
	applicationElement := application.Element
	return func(event *tcell.EventKey) *tcell.EventKey {
		if application.Model != nil {
			overlay := application.Model.CurrentOverlay()
			if overlay != OverlayNone {
				switch event.Key() {
				case tcell.KeyCtrlC:
					stopApplication(application)
					return nil
				case tcell.KeyEsc:
					application.closeOverlay()
					return nil
				}
				switch event.Rune() {
				case 'q':
					application.closeOverlay()
					return nil
				case '?':
					if overlay == OverlayHelp {
						application.closeOverlay()
					} else {
						application.openOverlay(OverlayHelp)
					}
					return nil
				case 'd':
					if overlay == OverlayDiagnostics {
						application.closeOverlay()
					} else {
						application.openOverlay(OverlayDiagnostics)
					}
					return nil
				case 'j':
					return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
				case 'k':
					return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
				}
				return event
			}
		}
		if application.HttpFilesTree.IsSearching() || application.Suites.IsSearching() || application.Producer.IsSearching() {
			return event
		}
		switch event.Rune() {
		case '?':
			application.openOverlay(OverlayHelp)
			return nil
		case 'd':
			application.openOverlay(OverlayDiagnostics)
			return nil
		}
		// handle 'q' to quit
		if event.Rune() == 'q' || event.Key() == tcell.KeyCtrlC {
			stopApplication(application)
			return nil
		}

		// Terminals commonly report Ctrl+h as Backspace.
		navigationKey := event.Key()
		isNavigationKey := event.Modifiers()&tcell.ModCtrl != 0
		if navigationKey == tcell.KeyBackspace {
			navigationKey = tcell.KeyCtrlH
			isNavigationKey = true
		}

		// Handle Ctrl+h/j/k/l for navigation.
		if isNavigationKey {
			focused := applicationElement.GetFocus()
			if focused == nil {
				return event
			}

			var target tview.Primitive
			switch navigationKey {
			case tcell.KeyCtrlH:
				if focused == application.Suites.Element || focused == application.Suite.Element {
					target = application.HttpFilesTree.Element
				} else if focused == application.Producer.Element {
					target = application.Suite.Element
				}
			case tcell.KeyCtrlJ:
				if focused == application.Suites.Element {
					target = application.Suite.Element
				}
			case tcell.KeyCtrlK:
				if focused == application.Suite.Element {
					target = application.Suites.Element
				}
			case tcell.KeyCtrlL:
				if focused == application.HttpFilesTree.Element {
					target = application.Suites.Element
				} else if focused == application.Suite.Element || focused == application.Suites.Element {
					target = application.Producer.Element
				}
			default:
				return event
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
	application.Producer.CancelActive()
	application.Suites.CancelLoad()
	application.HttpFilesTree.CancelReload()
	application.Element.Stop()
}
