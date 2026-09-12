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
	Element         tview.Primitive
	suitesArea      *tview.Flex
	treeElement     tview.Primitive
	suitesElement   tview.Primitive
	suiteElement    tview.Primitive
	producerElement tview.Primitive
	mode            layoutMode
	focus           focusArea
}

const (
	compactLayoutWidth = 80
	wideLayoutWidth    = 120
)

type layoutMode uint8

const (
	layoutWide layoutMode = iota
	layoutMedium
	layoutCompact
)

type focusArea uint8

const (
	focusFiles focusArea = iota
	focusRequests
	focusResponse
)

func (widget *Workspace) Build(
	parameters Parameters,
	treeWidget *tree.Tree,
	suitesWidget *suites.Suites,
	suiteWidget *suite.Suite,
	producerWidget *producer.Producer,
) tview.Primitive {
	suitesArea := tview.NewFlex().
		SetDirection(tview.FlexColumnCSS).
		AddItem(suitesWidget.Element, 0, 2, false).
		AddItem(suiteWidget.Element, 0, 1, false)
	suitesArea.SetBackgroundColor(parameters.Theme.Background)

	box := tview.NewFlex().SetDirection(tview.FlexRowCSS)
	box.SetBackgroundColor(parameters.Theme.Background)
	widget.Element = box
	widget.suitesArea = suitesArea
	widget.treeElement = treeWidget.Element
	widget.suitesElement = suitesWidget.Element
	widget.suiteElement = suiteWidget.Element
	widget.producerElement = producerWidget.Element
	widget.mode = layoutWide
	widget.focus = focusFiles
	widget.render()
	return box
}

// Resize chooses a layout that keeps every visible pane useful. Medium
// terminals show the current workflow step and its neighbour; compact
// terminals show only the active step.
func (widget *Workspace) Resize(width int, focused tview.Primitive) {
	mode := layoutWide
	if width < compactLayoutWidth {
		mode = layoutCompact
	} else if width < wideLayoutWidth {
		mode = layoutMedium
	}

	focus := widget.focus
	switch focused {
	case widget.treeElement:
		focus = focusFiles
	case widget.suitesElement, widget.suiteElement:
		focus = focusRequests
	case widget.producerElement:
		focus = focusResponse
	}
	if widget.mode == mode && widget.focus == focus {
		return
	}
	widget.mode = mode
	widget.focus = focus
	widget.render()
}

func (widget *Workspace) render() {
	box := widget.Element.(*tview.Flex)
	box.Clear()
	switch widget.mode {
	case layoutCompact:
		switch widget.focus {
		case focusFiles:
			box.AddItem(widget.treeElement, 0, 1, true)
		case focusRequests:
			box.AddItem(widget.suitesArea, 0, 1, true)
		case focusResponse:
			box.AddItem(widget.producerElement, 0, 1, true)
		}
	case layoutMedium:
		if widget.focus == focusFiles {
			box.AddItem(widget.treeElement, 0, 2, true).
				AddItem(widget.suitesArea, 0, 3, false)
		} else {
			box.AddItem(widget.suitesArea, 0, 3, widget.focus == focusRequests).
				AddItem(widget.producerElement, 0, 5, widget.focus == focusResponse)
		}
	default:
		box.AddItem(widget.treeElement, 0, 2, widget.focus == focusFiles).
			AddItem(widget.suitesArea, 0, 3, widget.focus == focusRequests).
			AddItem(widget.producerElement, 0, 4, widget.focus == focusResponse)
	}
}

func (widget *Workspace) ApplySettings(uiTheme theme.Theme) {
	widget.Element.(*tview.Flex).SetBackgroundColor(uiTheme.Background)
	widget.suitesArea.SetBackgroundColor(uiTheme.Background)
}
