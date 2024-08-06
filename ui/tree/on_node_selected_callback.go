package tree

import (
	"lazyrest/finder"

	"github.com/rivo/tview"
)

type (
	OnSelectFileCallbackType func(file finder.File)
	callbackType             func(node *tview.TreeNode)
)

func onNodeSelectedCallback(callback OnSelectFileCallbackType) callbackType {
	return func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		if _, isDirectory := reference.(finder.Directory); isDirectory {
			node.SetExpanded(!node.IsExpanded())
		}
		if file, isFile := reference.(finder.File); isFile {
			callback(file)
		}
	}
}
