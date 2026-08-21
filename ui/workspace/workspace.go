package workspace

import (
	"lazyrest/ui/producer"
	"lazyrest/ui/suite"
	"lazyrest/ui/suites"
	"lazyrest/ui/tree"

	"github.com/rivo/tview"
)

func New() Workspace {
	widget := Workspace{}
	return widget
}

type Workspace struct {
	Element tview.Primitive
}

func (widget *Workspace) Build(
	parameters Parameters,
	treeWidget tree.Tree,
	suitesWidget suites.Suites,
	suiteWidget suite.Suite,
	producerWidget producer.Producer,
) tview.Primitive {
	suitesArea := tview.NewFlex().
		SetDirection(tview.FlexColumnCSS).
		AddItem(suitesWidget.Element, 0, 1, false).
		AddItem(suiteWidget.Element, 0, 1, false)

	box := tview.NewFlex().
		SetDirection(tview.FlexRowCSS).
		AddItem(treeWidget.Element, 0, 2, true).
		AddItem(suitesArea, 0, 3, false).
		AddItem(producerWidget.Element, 0, 4, false)
	widget.Element = box
	return box
}
