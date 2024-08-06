package tree

import (
	"lazyrest/finder"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildDirectoryNode(directory finder.Directory, theme theme.TreeTheme) *tview.TreeNode {
	directoryNode := tview.NewTreeNode(directory.Name).
		SetColor(theme.NodeDirectory.Foreground).
		SetSelectable(true).
		SetReference(directory)
	return directoryNode
}
