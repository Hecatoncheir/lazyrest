package tree

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildFileNode(file finder.File, theme theme.TreeTheme) *tview.TreeNode {
	fileNode := tview.NewTreeNode(file.Name).
		SetColor(theme.Node.Foreground).
		SetSelectable(true).
		SetReference(file)
	return fileNode
}
