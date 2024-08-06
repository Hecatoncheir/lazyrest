package ui

import (
	"lazyrest/ui/footer"
	"lazyrest/ui/producer"
	"lazyrest/ui/suite"
	"lazyrest/ui/suites"
	"lazyrest/ui/tree"
	"lazyrest/ui/workspace"

	"github.com/rivo/tview"
)

func NewApplication() Application {
	widget := Application{}
	return widget
}

type Application struct {
	Element       *tview.Application
	HttpFilesTree tree.Tree
	Suites        suites.Suites
	Suite         suite.Suite
	Producer      producer.Producer
	Workspace     workspace.Workspace
	Footer        footer.Footer
}

func (widget *Application) Build() *tview.Application {
	application := tview.NewApplication().
		EnableMouse(true)
	widget.Element = application
	return application
}
