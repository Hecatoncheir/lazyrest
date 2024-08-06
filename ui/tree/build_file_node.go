package tree

import (
	"lazyrest/finder"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildFileNode(file finder.File, theme theme.TreeTheme) *tview.TreeNode {
	fileNode := tview.NewTreeNode(file.Name).
		SetColor(theme.Node.Foreground).
		SetSelectable(true).
		SetReference(file)
	return fileNode
}
