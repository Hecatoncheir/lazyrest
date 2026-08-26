package suites

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func (widget *Suites) ApplySettings(uiTheme theme.Theme, translator *locale.Translator, bindings *keymap.Bindings) {
	widget.theme = uiTheme.Suites
	widget.syntax = uiTheme.Syntax
	widget.methods = uiTheme.Methods
	widget.locale = translator
	widget.keybindings = bindings
	element := widget.Element.(*tview.List)
	box := element.Box
	applySuitesBoxTheme(box, widget.theme, element.HasFocus())
	box.SetFocusFunc(func() { applySuitesBoxTheme(box, widget.theme, true) })
	box.SetBlurFunc(func() { applySuitesBoxTheme(box, widget.theme, false) })
	widget.render()
}

func applySuitesBoxTheme(box *tview.Box, uiTheme theme.SuitesTheme, focused bool) {
	box.SetTitleColor(uiTheme.Title).SetBackgroundColor(uiTheme.Background).SetBorderColor(uiTheme.Border)
	if focused {
		box.SetTitleColor(uiTheme.TitleFocus).SetBackgroundColor(uiTheme.BackgroundFocus).SetBorderColor(uiTheme.BorderFocus)
	}
}
