package footer

import (
	"github.com/rivo/tview"
)

type Footer struct {
	Parameters           Parameters
	Element              tview.Primitive
	rootDirectoryElement *tview.TextView
	suiteElement         *tview.TextView
}

func New() Footer {
	element := Footer{}
	return element
}

func (widget *Footer) Build(parameters Parameters) tview.Primitive {
	widget.Parameters = parameters

	rootDirectoryPath := parameters.RootDirectoryPath
	footerTheme := parameters.Theme.Footer

	rootDirectoryElement, rootDirectoryElementSize := buildRootDirectoryElement(
		rootDirectoryPath,
		footerTheme,
	)

	if tv, ok := rootDirectoryElement.(*tview.TextView); ok {
		widget.rootDirectoryElement = tv
	}

	rootDirectoryPathTheme := footerTheme.RootDirectoryPath
	rootDirectoryArrowRightElement, rootDirectoryArrowRightElementSize := buildArrowRightElement(
		rootDirectoryPathTheme.ArrowBackground,
		rootDirectoryPathTheme.Foreground,
	)

	suiteText := tview.NewTextView().
		SetTextAlign(1).
		SetTextColor(footerTheme.SuiteForeground).
		SetBackgroundColor(footerTheme.Background)

	if tv, ok := suiteText.(*tview.TextView); ok {
		widget.suiteElement = tv
	}

	layout := tview.NewFlex().
		SetDirection(tview.FlexColumnCSS).
		AddItem(rootDirectoryElement, rootDirectoryElementSize, 0, false).
		AddItem(rootDirectoryArrowRightElement, rootDirectoryArrowRightElementSize, 0, false).
		AddItem(suiteText, 0, 1, false)

	theme := parameters.Theme.Footer
	layout.Box = tview.NewBox().
		SetBackgroundColor(theme.Background)

	widget.Element = layout

	return layout
}

func (widget *Footer) UpdateSuite(suiteName string) {
	if widget.suiteElement != nil {
		widget.suiteElement.SetText(" | Suite: " + suiteName)
	}
}

func (widget *Footer) UpdateRootDirectory(path string) {
	if widget.rootDirectoryElement != nil {
		widget.rootDirectoryElement.SetText(path)
	}
}
