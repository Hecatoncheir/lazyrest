package footer

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

type Footer struct {
	Parameters           Parameters
	Element              tview.Primitive
	rootDirectoryElement *tview.TextView
	suiteElement         *tview.TextView
	environmentElement   *tview.TextView
	statusElement        *tview.TextView
	selectedFile         *finder.File
	suiteName            string
	status               string
	indicatorState       IndicatorState
}

type IndicatorState uint8

const (
	IndicatorDefault IndicatorState = iota
	IndicatorSuccess
	IndicatorFailure
)

func New() *Footer {
	return &Footer{}
}

func (widget *Footer) Build(parameters Parameters) tview.Primitive {
	widget.Parameters = parameters

	layout := tview.NewFlex().SetDirection(tview.FlexRowCSS)

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

func (widget *Footer) UpdateStatus(status string) {
	widget.status = status
	widget.render()
}

func (widget *Footer) UpdateIndicatorState(state IndicatorState) {
	if widget.indicatorState == state {
		return
	}
	widget.indicatorState = state
	widget.render()
}

func (widget *Footer) indicatorTheme() theme.FooterIndicatorTheme {
	footerTheme := widget.Parameters.Theme.Footer
	switch widget.indicatorState {
	case IndicatorSuccess:
		return footerTheme.SuiteSuccess
	case IndicatorFailure:
		return footerTheme.SuiteFailure
	default:
		return theme.FooterIndicatorTheme{
			Background: footerTheme.SuiteBackground,
			Foreground: footerTheme.SuiteForeground,
		}
	}
}
