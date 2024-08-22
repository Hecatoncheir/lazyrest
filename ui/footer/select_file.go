package footer

import (
	"lazyrest/finder"

	"github.com/rivo/tview"
)

func (widget *Footer) SelectFile(file finder.File) {
	element := widget.Element.(*tview.Flex)
	element.Clear()

	parameters := widget.Parameters
	footerTheme := parameters.Theme.Footer

	rootDirectoryDirectoryElement, rootDirectoryDirectoryElementSize := buildRootDirectoryElement(parameters.RootDirectoryPath, footerTheme)
	element.AddItem(
		rootDirectoryDirectoryElement,
		rootDirectoryDirectoryElementSize,
		0,
		false,
	)

	selectedFileNameTheme := footerTheme.RootDirectoryPath
	rootDirectoryArrowRightElement, rootDirectoryArrowRightElementSize := buildArrowRightElement(
		selectedFileNameTheme.ArrowBackground,
		selectedFileNameTheme.ArrowForeground,
	)
	element.AddItem(
		rootDirectoryArrowRightElement,
		rootDirectoryArrowRightElementSize,
		0,
		false,
	)

	selectedFileElement, selectedFileElementSize := buildSelectedFileElement(file, selectedFileNameTheme)
	element.AddItem(
		selectedFileElement,
		selectedFileElementSize,
		0,
		false,
	)

	selectedFileElementArrowRightElement, selectedFileElementArrowRightElementSize := buildArrowRightElement(
		selectedFileNameTheme.ArrowBackground,
		selectedFileNameTheme.ArrowForeground,
	)
	element.AddItem(
		selectedFileElementArrowRightElement,
		selectedFileElementArrowRightElementSize,
		0,
		false,
	)
}
