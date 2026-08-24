package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type onInputCallbackType func(event *tcell.EventKey) *tcell.EventKey

func onInputCallback(application *Application) onInputCallbackType {
	applicationElement := application.Element
	return func(event *tcell.EventKey) *tcell.EventKey {
		if application.HttpFilesTree.IsSearching() || application.Suites.IsSearching() || application.Producer.IsSearching() {
			return event
		}
		// handle 'q' to quit
		if event.Rune() == 'q' {
			application.Producer.CancelActive()
			applicationElement.Stop()
			return event
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
