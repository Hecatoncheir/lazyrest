package tree

import (
	"lazyrest/finder"

	"github.com/rivo/tview"
)

type Tree struct {
	Element              tview.Primitive
	onSelectFileCallback OnSelectFileCallbackType
}

func New() Tree {
	widget := Tree{}
	return widget
}

func (widget *Tree) Build(parameters Parameters) tview.Primitive {
	onSelectFileCallback := parameters.OnSelectFileCallback
	widget.onSelectFileCallback = onSelectFileCallback

	theme := parameters.Theme.Tree
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

	rootDirectotyPath := parameters.RootDirectoryPath
	extensions := parameters.FilesExtension
	directoriesWithFiles, err := finder.FindFilesInDirectory(
		rootDirectotyPath,
		extensions,
	)
	if err != nil {
		widget := buildNoFilesFound(
			rootDirectotyPath,
			theme,
		)
		widget.Box = box
		return widget
	}

	element := buildTree(
		directoriesWithFiles,
		theme,
		onSelectFileCallback,
	)
	element.Box = box

	widget.Element = element

	return element
}
