package footer

import (
	"github.com/rivo/tview"
)

func New() Footer {
	element := Footer{}
	return element
}

type Footer struct {
	Element tview.Primitive
}

func (widget *Footer) Build(parameters Parameters) tview.Primitive {
	rootDirectoryPath := buildRootDirectoryPath(parameters)

	layout := tview.NewFlex().
		AddItem(rootDirectoryPath, 0, 1, true)

	theme := parameters.Theme.Footer
	layout.Box = tview.NewBox().
		SetBackgroundColor(theme.Background)

	widget.Element = layout

	return layout
}
