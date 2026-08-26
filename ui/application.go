package ui

import (
	"context"
	"sync"

	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/layout"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/Hecatoncheir/lazyrest/ui/workspace"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewApplication() *Application {
	return &Application{}
}

type Application struct {
	Element           *tview.Application
	Pages             *tview.Pages
	Model             *Model
	HttpFilesTree     *tree.Tree
	Suites            *suites.Suites
	Suite             *suite.Suite
	Producer          *producer.Producer
	Workspace         *workspace.Workspace
	Layout            *layout.Layout
	Footer            *footer.Footer
	Diagnostics       *tview.TextView
	Help              *tview.TextView
	Captured          *tview.TextView
	History           *tview.List
	CommandPalette    *tview.List
	ThemePicker       *tview.List
	EnvironmentPicker *tview.List
	SaveResponse      *tview.InputField

	config            Config
	theme             theme.Theme
	loadEnvironment   func(string, environment.Config) (environment.Environment, error)
	scanFiles         func(context.Context) tree.ScanResult
	previousFocus     tview.Primitive
	screen            tcell.Screen
	pendingExport     *producer.ResponseExport
	saveFullResponse  bool
	saveOverwritePath string
	pendingViewKeys   []*tcell.EventKey
	pendingViewFocus  tview.Primitive
	startOnce         sync.Once

	footerProgressMutex  sync.Mutex
	footerProgressCancel context.CancelFunc
	footerProgressLabel  string

	environmentLoadMutex sync.Mutex
	environmentLoadID    uint64
}

func (widget *Application) startEnvironmentLoad() uint64 {
	widget.environmentLoadMutex.Lock()
	defer widget.environmentLoadMutex.Unlock()
	widget.environmentLoadID++
	return widget.environmentLoadID
}

func (widget *Application) isCurrentEnvironmentLoad(loadID uint64) bool {
	widget.environmentLoadMutex.Lock()
	defer widget.environmentLoadMutex.Unlock()
	return widget.environmentLoadID == loadID
}

func (widget *Application) Build() *tview.Application {
	application := tview.NewApplication().
		EnableMouse(true)
	widget.Element = application
	return application
}
