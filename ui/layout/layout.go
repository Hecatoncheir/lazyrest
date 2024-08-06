package layout

import (
	"lazyrest/ui/footer"
	"lazyrest/ui/workspace"

	"github.com/rivo/tview"
)

func New() Layout {
	element := Layout{}
	return element
}

type Layout struct{}

func (element *Layout) Build(
	parameters Parameters,
	workspaceWidget workspace.Workspace,
	footerWidget footer.Footer,
) tview.Primitive {
	workspaceElement := workspaceWidget.Element
	footerElement := footerWidget.Element
	box := tview.NewFlex().
		SetDirection(tview.FlexColumnCSS).
		AddItem(workspaceElement, 0, 1, false).
		AddItem(footerElement, 1, 1, false)
	return box
}
