package layout

import (
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/Hecatoncheir/lazyrest/ui/workspace"

	"github.com/rivo/tview"
)

func New() *Layout {
	return &Layout{}
}

type Layout struct {
	Element *tview.Flex
}

func (element *Layout) Build(
	parameters Parameters,
	workspaceWidget *workspace.Workspace,
	footerWidget *footer.Footer,
) tview.Primitive {
	workspaceElement := workspaceWidget.Element
	footerElement := footerWidget.Element
	box := tview.NewFlex().
		SetDirection(tview.FlexColumnCSS).
		AddItem(workspaceElement, 0, 1, false).
		AddItem(footerElement, 1, 1, false)
	box.SetBackgroundColor(parameters.Theme.Background)
	element.Element = box
	return box
}

func (element *Layout) ApplySettings(uiTheme theme.Theme) {
	element.Element.SetBackgroundColor(uiTheme.Background)
}
