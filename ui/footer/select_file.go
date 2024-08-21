package footer

import (
	"lazyrest/finder"

	"github.com/rivo/tview"
)

func (widget *Footer) SelectFile(file finder.File) {
	element := widget.Element.(*tview.Flex)
	element.RemoveItem(widget.selectedFileElement)

	parameters := widget.Parameters
	selectedFileNameTheme := parameters.Theme.Footer.RootDirectoryPath

	element.RemoveItem(widget.rootDirectoryArrowRightElement)
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
}
