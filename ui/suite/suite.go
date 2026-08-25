package suite

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func New() *Suite {
	return &Suite{}
}

type Suite struct {
	Element          tview.Primitive
	theme            theme.SuiteTheme
	suite            http.HttpSuite
	onEscapeCallback OnEscapeCallbackType
	onRunCallback    OnRunCallbackType
	keybindings      *keymap.Bindings
	locale           *locale.Translator
	syntax           syntax.Palette
}

func (widget *Suite) Build(parameters Parameters) tview.Primitive {
	if parameters.Keybindings == nil {
		parameters.Keybindings = keymap.Default()
	}
	if parameters.Locale == nil {
		parameters.Locale = locale.English()
	}
	widget.onEscapeCallback = parameters.OnEscapeCallback
	widget.onRunCallback = parameters.OnRunCallback
	widget.keybindings = parameters.Keybindings
	widget.locale = parameters.Locale

	theme := parameters.Theme.Suite
	widget.theme = theme

	widget.syntax = parameters.Theme.Syntax

	element := tview.NewTextView()
	element.
		SetDynamicColors(true).
		SetTextColor(theme.Foreground).
		SetBackgroundColor(theme.Background)

	element.SetFocusFunc(func() {
		element.
			SetTitleColor(theme.TitleFocus).
			SetBackgroundColor(theme.BackgroundFocus).
			SetBorderColor(theme.BorderFocus)
	})

	element.SetBlurFunc(func() {
		element.
			SetTitleColor(theme.Title).
			SetBackgroundColor(theme.Background).
			SetBorderColor(theme.Border)
	})

	box := tview.NewBox().
		SetBorder(true).
		SetBorderColor(theme.Border).
		SetBackgroundColor(theme.Background).
		SetTitle(widget.locale.Text("suite")).
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
