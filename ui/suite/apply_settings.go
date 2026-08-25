package suite

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func (widget *Suite) ApplySettings(uiTheme theme.Theme, translator *locale.Translator, bindings *keymap.Bindings) {
	widget.theme = uiTheme.Suite
	widget.syntax = uiTheme.Syntax
	widget.locale = translator
	widget.keybindings = bindings
	element := widget.Element.(*tview.TextView)
	applySuiteTheme(element, widget.theme, element.HasFocus())
	element.SetFocusFunc(func() { applySuiteTheme(element, widget.theme, true) })
	element.SetBlurFunc(func() { applySuiteTheme(element, widget.theme, false) })
	element.SetTitle(widget.locale.Text("suite"))
	if widget.suite.Name != "" || widget.suite.Uri != "" {
		widget.ChangeSuite(widget.suite)
	}
}

func applySuiteTheme(element *tview.TextView, uiTheme theme.SuiteTheme, focused bool) {
	element.SetTitleColor(uiTheme.Title).SetBackgroundColor(uiTheme.Background).SetBorderColor(uiTheme.Border)
	element.SetTextColor(uiTheme.Foreground)
	if focused {
		element.SetTitleColor(uiTheme.TitleFocus).SetBackgroundColor(uiTheme.BackgroundFocus).SetBorderColor(uiTheme.BorderFocus)
	}
}
