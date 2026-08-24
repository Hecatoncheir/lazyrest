package footer

import (
	"github.com/Hecatoncheir/lazyrest/finder"

	"github.com/rivo/tview"
)

type Footer struct {
	Parameters           Parameters
	Element              tview.Primitive
	rootDirectoryElement *tview.TextView
	suiteElement         *tview.TextView
	selectedFile         *finder.File
	suiteName            string
}

func New() *Footer {
	return &Footer{}
}

func (widget *Footer) Build(parameters Parameters) tview.Primitive {
	widget.Parameters = parameters

	layout := tview.NewFlex().SetDirection(tview.FlexColumnCSS)

	theme := parameters.Theme.Footer
	layout.Box = tview.NewBox().
		SetBackgroundColor(theme.Background)

	widget.Element = layout
	widget.render()

	return layout
}

func (widget *Footer) UpdateSuite(suiteName string) {
	widget.suiteName = suiteName
	widget.render()
}

func (widget *Footer) UpdateRootDirectory(path string) {
	widget.Parameters.RootDirectoryPath = path
	widget.render()
}
