package suites

import (
	"context"
	"sync"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"
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
	syntax                syntax.Palette
	methods               syntax.MethodPalette
	rows                  []listRow
	suites                []http.HttpSuite
	diagnosticCount       int
	searchQuery           string
	searchMode            bool
	onEscapeCallback      OnEscapeCallbackType
	onSuiteSelectCallback OnSuiteSelectCallbackType
	parseOptions          http.ParseOptions
	loadMutex             sync.Mutex
	loadID                uint64
	cancelLoad            context.CancelFunc
	keybindings           *keymap.Bindings
	locale                *locale.Translator
}

func (widget *Suites) Build(parameters Parameters) tview.Primitive {
	if parameters.Keybindings == nil {
		parameters.Keybindings = keymap.Default()
	}
	if parameters.Locale == nil {
		parameters.Locale = locale.English()
	}
	widget.onEscapeCallback = parameters.OnEscapeCallback
	widget.onSuiteSelectCallback = parameters.OnSuiteSelectCallbackType
	widget.parseOptions = parameters.ParseOptions
	widget.keybindings = parameters.Keybindings
	widget.locale = parameters.Locale

	theme := parameters.Theme.Suites
	widget.theme = theme
	widget.syntax = parameters.Theme.Syntax
	widget.methods = parameters.Theme.Methods

	element := tview.NewList()
	// Both the row and its preview carry markup, so the list must not escape
	// them; every part of the text is escaped or highlighted here instead.
	element.SetUseStyleTags(true, true)
	element.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		widget.applySelectionMarkup(index)
	})

	box := tview.NewBox().
		SetBorder(true).
		SetBackgroundColor(theme.Background).
		SetBorderColor(theme.Border).
		SetTitle(widget.locale.Text("suites")).
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
