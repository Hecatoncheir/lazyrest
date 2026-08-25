package tree

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildTree(
	rootDirectory finder.Directory,
	theme theme.TreeTheme,
	callback OnSelectFileCallbackType,
	translator *locale.Translator,
) *tview.TreeView {
	rootNode := buildDirectoryTree(rootDirectory, theme)
	for _, warning := range rootDirectory.Warnings {
		rootNode.AddChild(tview.NewTreeNode(translator.Format("warning", warning)).
			SetSelectable(false).
			SetColor(theme.Node.Foreground))
	}

	tree := tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)

	tree.
		SetSelectedFunc(onNodeSelectedCallback(callback))

	return tree
}
