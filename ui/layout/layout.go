package layout

import (
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/workspace"

	"github.com/rivo/tview"
)

func New() *Layout {
	return &Layout{}
}

type Layout struct{}

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
	return box
}
