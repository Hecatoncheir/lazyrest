package tree

import (
	"context"
	"fmt"
	"sync"

	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

type Tree struct {
	Element              tview.Primitive
	onSelectFileCallback OnSelectFileCallbackType
	searchMode           bool
	searchQuery          string
	searchMatches        []*tview.TreeNode
	searchIndex          int
	rootDirectoryPath    string
	filesExtension       []string
	theme                theme.TreeTheme
	onReloadCallback     func()
	loading              bool
	reloading            bool
	reloadMutex          sync.Mutex
	reloadID             uint64
	cancelReload         context.CancelFunc
}

func New() *Tree {
	return &Tree{}
}

func (widget *Tree) IsSearching() bool {
	return widget.searchMode
}

func (widget *Tree) OpenCurrentFile() bool {
	element, ok := widget.Element.(*tview.TreeView)
	if !ok {
		return false
	}
	node := element.GetCurrentNode()
	if node == nil {
		return false
	}
	file, ok := node.GetReference().(finder.File)
	if !ok || widget.onSelectFileCallback == nil {
		return false
	}
	widget.onSelectFileCallback(file)
	return true
}

func (widget *Tree) Build(parameters Parameters) tview.Primitive {
	onSelectFileCallback := parameters.OnSelectFileCallback
	widget.onSelectFileCallback = onSelectFileCallback
	widget.onReloadCallback = parameters.OnReloadCallback
	widget.rootDirectoryPath = parameters.RootDirectoryPath
	widget.filesExtension = append([]string(nil), parameters.FilesExtension...)

	theme := parameters.Theme.Tree
	widget.theme = theme
	box := tview.NewBox().
		SetBorder(true).
		SetBorderColor(theme.Border).
		SetBorderPadding(0, 0, 0, 0).
		SetBackgroundColor(theme.Background).
		SetTitle("Files").
		SetTitleColor(theme.Title).
		SetTitleAlign(1)

	box.SetFocusFunc(func() {
		box.
			SetTitleColor(theme.TitleFocus).
			SetBackgroundColor(theme.BackgroundFocus).
			SetBorderColor(theme.BorderFocus)
	})

	box.SetBlurFunc(func() {
		box.
			SetTitleColor(theme.Title).
			SetBackgroundColor(theme.Background).
			SetBorderColor(theme.Border)
	})

	element := tview.NewTreeView()
	element.Box = box
	element.SetInputCapture(onInputCallback(widget))
	widget.Element = element
	widget.ShowLoading()

	return element
}

type ScanResult struct {
	Directory finder.Directory
	Err       error
}

func (widget *Tree) StartReload() (context.Context, uint64) {
	widget.reloadMutex.Lock()
	defer widget.reloadMutex.Unlock()
	if widget.cancelReload != nil {
		widget.cancelReload()
	}
	widget.reloadID++
	ctx, cancel := context.WithCancel(context.Background())
	widget.cancelReload = cancel
	return ctx, widget.reloadID
}

func (widget *Tree) IsCurrentReload(reloadID uint64) bool {
	widget.reloadMutex.Lock()
	defer widget.reloadMutex.Unlock()
	return widget.reloadID == reloadID
}

func (widget *Tree) FinishReload(reloadID uint64) bool {
	widget.reloadMutex.Lock()
	defer widget.reloadMutex.Unlock()
	if widget.reloadID != reloadID {
		return false
	}
	if widget.cancelReload != nil {
		widget.cancelReload()
		widget.cancelReload = nil
	}
	return true
}

func (widget *Tree) CancelReload() {
	widget.reloadMutex.Lock()
	defer widget.reloadMutex.Unlock()
	widget.reloadID++
	if widget.cancelReload != nil {
		widget.cancelReload()
		widget.cancelReload = nil
	}
}

func (widget *Tree) Scan(ctx context.Context) ScanResult {
	directory, err := finder.FindFilesInDirectoryContext(ctx, widget.rootDirectoryPath, widget.filesExtension)
	return ScanResult{Directory: directory, Err: err}
}

func (widget *Tree) ShowLoading() {
	widget.loading = true
	widget.reloading = false
	element := widget.Element.(*tview.TreeView)
	root := tview.NewTreeNode("Loading files...").
		SetSelectable(false).
		SetColor(widget.theme.Node.Foreground)
	element.SetRoot(root).SetCurrentNode(root)
	widget.updateTitle()
}

func (widget *Tree) ShowReloading() {
	widget.loading = false
	widget.reloading = true
	widget.updateTitle()
}

func (widget *Tree) ApplyScanResult(result ScanResult) {
	element := widget.Element.(*tview.TreeView)
	selectedPath := currentFilePath(element.GetCurrentNode())
	widget.loading = false
	widget.reloading = false

	if result.Err != nil {
		root := tview.NewTreeNode(fmt.Sprintf("Unable to scan directory: %v", result.Err)).
			SetSelectable(false).
			SetColor(widget.theme.Node.Foreground)
		element.SetRoot(root).SetCurrentNode(root)
		widget.updateTitle()
		return
	}

	newTree := buildTree(result.Directory, widget.theme, widget.onSelectFileCallback)
	root := newTree.GetRoot()
	element.SetRoot(root).SetSelectedFunc(onNodeSelectedCallback(widget.onSelectFileCallback))
	if selectedPath != "" {
		if selectedNode := findFileNode(root, selectedPath); selectedNode != nil {
			element.SetCurrentNode(selectedNode)
		} else {
			element.SetCurrentNode(root)
		}
	} else {
		element.SetCurrentNode(root)
	}
	widget.updateSearch()
}

func currentFilePath(node *tview.TreeNode) string {
	if node == nil {
		return ""
	}
	file, ok := node.GetReference().(finder.File)
	if !ok {
		return ""
	}
	return file.Path
}

func findFileNode(node *tview.TreeNode, path string) *tview.TreeNode {
	if node == nil {
		return nil
	}
	if currentFilePath(node) == path {
		return node
	}
	for _, child := range node.GetChildren() {
		if match := findFileNode(child, path); match != nil {
			node.SetExpanded(true)
			return match
		}
	}
	return nil
}
