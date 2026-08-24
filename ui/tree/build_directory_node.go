package tree

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildDirectoryNode(directory finder.Directory, theme theme.TreeTheme) *tview.TreeNode {
	directoryNode := tview.NewTreeNode(directory.Name).
		SetColor(theme.NodeDirectory.Foreground).
		SetSelectable(true).
		SetReference(directory)
	return directoryNode
}
