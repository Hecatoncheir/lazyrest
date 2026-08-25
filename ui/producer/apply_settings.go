package producer

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func (widget *Producer) ApplySettings(uiTheme theme.Theme, translator *locale.Translator, bindings *keymap.Bindings) {
	widget.theme = uiTheme.Producer
	widget.locale = translator
	widget.keybindings = bindings
	element := widget.Element.(*tview.TextView)
	widget.applyTheme(element, element.HasFocus())
	element.SetFocusFunc(func() { widget.applyTheme(element, true) })
	element.SetBlurFunc(func() { widget.applyTheme(element, false) })
	if !widget.IsRunning() && len(widget.history) > 0 {
		widget.setText(widget.renderEntry(widget.history[widget.historyIndex]))
	}
	widget.updateTitle()
}

func (widget *Producer) applyTheme(element *tview.TextView, focused bool) {
	applyProducerTheme(element, widget.theme, focused)
}

func applyProducerTheme(element *tview.TextView, uiTheme theme.ProducerTheme, focused bool) {
	element.SetTitleColor(uiTheme.Title).SetBackgroundColor(uiTheme.Background).SetBorderColor(uiTheme.Border)
	element.SetTextColor(uiTheme.Foreground)
	if focused {
		element.SetTitleColor(uiTheme.TitleFocus).SetBackgroundColor(uiTheme.BackgroundFocus).SetBorderColor(uiTheme.BorderFocus)
	}
}
