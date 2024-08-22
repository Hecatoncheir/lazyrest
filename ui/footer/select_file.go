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
	selectedFileNameTheme := footerTheme.RootDirectoryPath

	rootDirectoryDirectoryElement, rootDirectoryDirectoryElementSize := buildRootDirectoryElement(parameters.RootDirectoryPath, footerTheme)
	element.AddItem(
		rootDirectoryDirectoryElement,
		rootDirectoryDirectoryElementSize,
		0,
		false,
	)

	rootDirectoryArrowRightElement, rootDirectoryArrowRightElementSize := buildArrowRightElement(
		selectedFileNameTheme.ArrowBackground,
		selectedFileNameTheme.ArrowForeground,
	)
	widget.rootDirectoryArrowRightElement = rootDirectoryArrowRightElement
	element.AddItem(
		rootDirectoryArrowRightElement,
		rootDirectoryArrowRightElementSize,
		0,
		false,
	)

	selectedFileElement, selectedFileElementSize := buildSelectedFileElement(file, selectedFileNameTheme)
	widget.selectedFileElement = selectedFileElement
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
	widget.selectedFileElementArrowRightElement = selectedFileElementArrowRightElement
	element.AddItem(
		selectedFileElementArrowRightElement,
		selectedFileElementArrowRightElementSize,
		0,
		false,
	)
}
