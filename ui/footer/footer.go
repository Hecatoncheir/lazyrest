package footer

import (
	"github.com/rivo/tview"
)

func New() Footer {
	element := Footer{}
	return element
}

type Footer struct {
	Parameters Parameters
	Element    tview.Primitive
}

func (widget *Footer) Build(parameters Parameters) tview.Primitive {
	widget.Parameters = parameters

	rootDirectoryPath := parameters.RootDirectoryPath
	footerTheme := parameters.Theme.Footer

	rootDirectoryElement, rootDirectoryElementSize := buildRootDirectoryElement(
		rootDirectoryPath,
		footerTheme,
	)

	rootDirectoryPathTheme := footerTheme.RootDirectoryPath
	rootDirectoryArrowRightElement, rootDirectoryArrowRightElementSize := buildArrowRightElement(
		rootDirectoryPathTheme.ArrowBackground,
		rootDirectoryPathTheme.ArrowForeground,
	)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRowCSS).
		AddItem(rootDirectoryElement, rootDirectoryElementSize, 0, false).
		AddItem(rootDirectoryArrowRightElement, rootDirectoryArrowRightElementSize, 0, false)

	theme := parameters.Theme.Footer
	layout.Box = tview.NewBox().
		SetBackgroundColor(theme.Background)

	widget.Element = layout

	return layout
}
