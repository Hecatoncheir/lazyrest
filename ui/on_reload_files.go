package ui

func onReloadFiles(application *Application) func() {
	return func() {
		treeWidget := application.HttpFilesTree
		ctx, reloadID := treeWidget.StartReload()
		treeWidget.ShowReloading()

		go func() {
			result := treeWidget.Scan(ctx)
			if !treeWidget.IsCurrentReload(reloadID) {
				return
			}
			application.Element.QueueUpdateDraw(func() {
				if !treeWidget.FinishReload(reloadID) {
					return
				}
				treeWidget.ApplyScanResult(result)
			})
		}()
	}
}
