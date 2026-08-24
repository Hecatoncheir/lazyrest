package ui

import (
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/Hecatoncheir/lazyrest/ui/workspace"

	"github.com/rivo/tview"
)

func NewApplication() *Application {
	return &Application{}
}

type Application struct {
	Element       *tview.Application
	HttpFilesTree *tree.Tree
	Suites        *suites.Suites
	Suite         *suite.Suite
	Producer      *producer.Producer
	Workspace     *workspace.Workspace
	Footer        *footer.Footer
}

func (widget *Application) Build() *tview.Application {
	application := tview.NewApplication().
		EnableMouse(true)
	widget.Element = application
	return application
}
