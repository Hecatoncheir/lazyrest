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
	widget.focused = element.HasFocus()
	applySuitesBoxTheme(box, widget.theme, widget.focused)
	box.SetFocusFunc(func() {
		widget.focused = true
		applySuitesBoxTheme(box, widget.theme, true)
		widget.updateTitle()
	})
	box.SetBlurFunc(func() {
		widget.focused = false
		applySuitesBoxTheme(box, widget.theme, false)
		widget.updateTitle()
	})
	widget.render()
}

func applySuitesBoxTheme(box *tview.Box, uiTheme theme.SuitesTheme, focused bool) {
	box.SetTitleColor(uiTheme.Title).SetBackgroundColor(uiTheme.Background).SetBorderColor(uiTheme.Border)
	if focused {
		box.SetTitleColor(uiTheme.TitleFocus).SetBackgroundColor(uiTheme.BackgroundFocus).SetBorderColor(uiTheme.BorderFocus)
	}
}
