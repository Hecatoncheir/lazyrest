package workspace

import (
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/Hecatoncheir/lazyrest/ui/tree"

	"github.com/rivo/tview"
)

func New() *Workspace {
	return &Workspace{}
}

type Workspace struct {
	Element    tview.Primitive
	suitesArea *tview.Flex
}

func (widget *Workspace) Build(
	parameters Parameters,
	treeWidget *tree.Tree,
	suitesWidget *suites.Suites,
	suiteWidget *suite.Suite,
	producerWidget *producer.Producer,
) tview.Primitive {
	suitesArea := tview.NewFlex().
		SetDirection(tview.FlexColumnCSS).
		AddItem(suitesWidget.Element, 0, 1, false).
		AddItem(suiteWidget.Element, 0, 1, false)
	suitesArea.SetBackgroundColor(parameters.Theme.Background)

	box := tview.NewFlex().
		SetDirection(tview.FlexRowCSS).
		AddItem(treeWidget.Element, 0, 2, true).
		AddItem(suitesArea, 0, 3, false).
		AddItem(producerWidget.Element, 0, 4, false)
	box.SetBackgroundColor(parameters.Theme.Background)
	widget.Element = box
	widget.suitesArea = suitesArea
	return box
}

func (widget *Workspace) ApplySettings(uiTheme theme.Theme) {
	widget.Element.(*tview.Flex).SetBackgroundColor(uiTheme.Background)
	widget.suitesArea.SetBackgroundColor(uiTheme.Background)
}
