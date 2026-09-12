package producer

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func (widget *Producer) ApplySettings(uiTheme theme.Theme, translator *locale.Translator, bindings *keymap.Bindings) {
	widget.theme = uiTheme.Producer
	widget.syntax = uiTheme.Syntax
	widget.locale = translator
	widget.keybindings = bindings
	element := widget.Element.(*tview.TextView)
	widget.focused = element.HasFocus()
	widget.applyTheme(element, widget.focused)
	element.SetFocusFunc(func() {
		widget.focused = true
		widget.applyTheme(element, true)
		widget.updateTitle()
	})
	element.SetBlurFunc(func() {
		widget.focused = false
		widget.applyTheme(element, false)
		widget.updateTitle()
	})
	if entry, ok := widget.currentHistoryEntry(); !widget.IsRunning() && ok {
		widget.setText(widget.renderEntry(entry))
	}
	widget.updateTitle()
}

func (widget *Producer) applyTheme(element *tview.TextView, focused bool) {
	applyProducerTheme(element, widget.theme, focused)
}

func applyProducerTheme(element *tview.TextView, uiTheme theme.ProducerTheme, focused bool) {
	element.SetTitleColor(uiTheme.Title).SetBorderColor(uiTheme.Border)
	element.SetTextColor(uiTheme.Foreground).SetBackgroundColor(uiTheme.Background)
	if focused {
		element.SetTitleColor(uiTheme.TitleFocus).SetBorderColor(uiTheme.BorderFocus)
		element.SetBackgroundColor(uiTheme.BackgroundFocus)
	}
}
