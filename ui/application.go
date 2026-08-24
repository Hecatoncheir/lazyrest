package ui

import (
	"context"
	"sync"

	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/Hecatoncheir/lazyrest/ui/workspace"

	"github.com/rivo/tview"
)

func NewApplication() *Application {
	return &Application{}
}

type Application struct {
	Element       *tview.Application
	Pages         *tview.Pages
	Model         *Model
	HttpFilesTree *tree.Tree
	Suites        *suites.Suites
	Suite         *suite.Suite
	Producer      *producer.Producer
	Workspace     *workspace.Workspace
	Footer        *footer.Footer
	Diagnostics   *tview.TextView
	Help          *tview.TextView

	config          Config
	theme           theme.Theme
	loadEnvironment func(string, environment.Config) (environment.Environment, error)
	scanFiles       func(context.Context) tree.ScanResult
	previousFocus   tview.Primitive
	startOnce       sync.Once
}

func (widget *Application) Build() *tview.Application {
	application := tview.NewApplication().
		EnableMouse(true)
	widget.Element = application
	return application
}
