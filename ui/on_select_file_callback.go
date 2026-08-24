package ui

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
)

func onSelectFileCallback(applicationWidget *Application) tree.OnSelectFileCallbackType {
	return func(file finder.File) {
		applicationWidget.Footer.SelectFile(file)

		applicationWidget.Suites.ChangeSuitesFromFile(file)
		suitesElement := applicationWidget.Suites.Element

		applicationElement := applicationWidget.Element
		applicationElement.SetFocus(suitesElement)
	}
}
