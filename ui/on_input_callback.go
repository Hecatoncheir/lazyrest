package ui

import (
	"github.com/gdamore/tcell/v2"
)

type onInputCallbackType func(event *tcell.EventKey) *tcell.EventKey

func onInputCallback(application *Application) onInputCallbackType {
	applicationElement := application.Element
	return func(event *tcell.EventKey) *tcell.EventKey {
		// switch event.Key() {
		// case tcell.KeyEsc:
		// 	httpFilesTreeElement := application.HttpFilesTree.Element
		// 	applicationElement.SetFocus(httpFilesTreeElement)
		// }

		switch event.Rune() {
		case 'q':
			applicationElement.Stop()
		}
		return event
	}
}
