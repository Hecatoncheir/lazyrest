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

		// handle Ctrl+h/j/k/l for navigation
		if event.Modifiers() == tcell.ModCtrl {
			focused := applicationElement.GetFocus()
			if focused == nil {
				return event
			}

			var target tview.Primitive
			switch event.Rune() {
			case 'h': // Left
				if focused == application.Suite.Element {
					target = application.Suites.Element
				} else if focused == application.Footer.Element {
					target = application.HttpFilesTree.Element
				}
			case 'l': // Right
				if focused == application.Suites.Element {
					target = application.Suite.Element
				} else if isWorkspaceElement(focused, application) {
					target = application.Footer.Element
				}
			case 'k': // Up
				if focused == application.Suite.Element || focused == application.Suites.Element {
					target = application.HttpFilesTree.Element
				} else if focused == application.Producer.Element {
					target = application.Suites.Element
				}
			case 'j': // Down
				if focused == application.HttpFilesTree.Element {
					target = application.Suites.Element
				} else if focused == application.Suite.Element || focused == application.Suites.Element {
					target = application.Producer.Element
				}
			}

			if target != nil {
				applicationElement.SetFocus(target)
				return event
			}
		}

		return event
	}
}

func isWorkspaceElement(focused tview.Primitive, app *Application) bool {
	return focused == app.HttpFilesTree.Element ||
		focused == app.Suites.Element ||
		focused == app.Suite.Element ||
		focused == app.Producer.Element
}
