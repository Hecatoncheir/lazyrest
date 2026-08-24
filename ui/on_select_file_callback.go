package ui

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
)

func onSelectFileCallback(applicationWidget *Application) tree.OnSelectFileCallbackType {
	return func(file finder.File) {
		applicationWidget.Footer.SelectFile(file)

		suitesWidget := applicationWidget.Suites
		ctx, loadID := suitesWidget.StartLoad()
		suitesWidget.ShowLoading(file)
		suitesElement := applicationWidget.Suites.Element

		applicationElement := applicationWidget.Element
		applicationElement.SetFocus(suitesElement)

		go func() {
			result := suitesWidget.LoadSuitesFromFile(ctx, file)
			if !suitesWidget.IsCurrentLoad(loadID) {
				return
			}
			applicationElement.QueueUpdateDraw(func() {
				if !suitesWidget.FinishLoad(loadID) {
					return
				}
				suitesWidget.ApplyLoadResult(result)
			})
		}()
	}
}
