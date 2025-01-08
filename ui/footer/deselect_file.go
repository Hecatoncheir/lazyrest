package footer

import (
	"github.com/rivo/tview"
)

func (widget *Footer) DeselectFile() {
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

	rootDirectoryPathTheme := footerTheme.RootDirectoryPath
	rootDirectoryArrowRightElement, rootDirectoryArrowRightElementSize := buildArrowRightElement(
		rootDirectoryPathTheme.ArrowBackground,
		rootDirectoryPathTheme.ArrowForeground,
	)
	element.AddItem(
		rootDirectoryArrowRightElement,
		rootDirectoryArrowRightElementSize,
		0,
		false,
	)
}
