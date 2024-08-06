package tree

import (
	"lazyrest/finder"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildTree(
	rootDirectory finder.Directory,
	theme theme.TreeTheme,
	callback OnSelectFileCallbackType,
) *tview.TreeView {
	rootNode := buildDirectoryTree(rootDirectory, theme)

	tree := tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)

	tree.
		SetSelectedFunc(onNodeSelectedCallback(callback))

	return tree
}
