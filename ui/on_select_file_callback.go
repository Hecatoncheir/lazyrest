package ui

import (
	"lazyrest/finder"
	"lazyrest/ui/tree"
)

func onSelectFileCallback(applicationWidget *Application) tree.OnSelectFileCallbackType {
	return func(file finder.File) {
		suitesWidget := applicationWidget.Suites
		suitesWidget.ChangeSuitesFromFile(file)
		suitesElement := suitesWidget.Element

		applicationElement := applicationWidget.Element
		applicationElement.SetFocus(suitesElement)
	}
}
