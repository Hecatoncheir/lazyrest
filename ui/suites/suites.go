package suites

import (
	"context"
	"sync"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func New() *Suites {
	return &Suites{}
}

func (widget *Suites) IsSearching() bool {
	return widget.searchMode
}

func (widget *Suites) SetParseOptions(options http.ParseOptions) {
	widget.parseOptions = options
}

type Suites struct {
	Element               tview.Primitive
	theme                 theme.SuitesTheme
	suites                []http.HttpSuite
	diagnosticCount       int
	selectedSuite         http.HttpSuite
	searchQuery           string
	searchMode            bool
	onEscapeCallback      OnEscapeCallbackType
	onSuiteSelectCallback OnSuiteSelectCallbackType
	parseOptions          http.ParseOptions
	loadMutex             sync.Mutex
	loadID                uint64
	cancelLoad            context.CancelFunc
}

func (widget *Suites) Build(parameters Parameters) tview.Primitive {
	widget.onEscapeCallback = parameters.OnEscapeCallback
	widget.onSuiteSelectCallback = parameters.OnSuiteSelectCallbackType
	widget.parseOptions = parameters.ParseOptions

	theme := parameters.Theme.Suites
	widget.theme = theme

	element := tview.NewList()

	box := tview.NewBox().
		SetBorder(true).
		SetBackgroundColor(theme.Background).
		SetBorderColor(theme.Border).
		SetTitle("Suites").
		SetInputCapture(onInputCallback(widget))

	box.SetFocusFunc(func() {
		box.
			SetTitleColor(theme.TitleFocus).
			SetBackgroundColor(theme.BackgroundFocus).
			SetBorderColor(theme.BorderFocus)
	})

	box.SetBlurFunc(func() {
		box.
			SetTitleColor(theme.Title).
			SetBackgroundColor(theme.Background).
			SetBorderColor(theme.Border)
	})

	element.Box = box

	widget.Element = element
	return element
}
