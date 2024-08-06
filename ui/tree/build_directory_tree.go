package tree

import (
	"lazyrest/finder"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildDirectoryTree(directory finder.Directory, theme theme.TreeTheme) *tview.TreeNode {
	directoryNode := buildDirectoryNode(directory, theme)

	for _, directory := range directory.Directories {
		node := buildDirectoryTree(directory, theme)
		directoryNode.AddChild(node)
	}

	for _, file := range directory.Files {
		node := buildFileNode(file, theme)
		directoryNode.AddChild(node)
	}

	return directoryNode
}
