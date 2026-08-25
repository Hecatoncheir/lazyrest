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
	applyProducerTheme(element, widget.theme, element.HasFocus())
	element.SetFocusFunc(func() { applyProducerTheme(element, widget.theme, true) })
	element.SetBlurFunc(func() { applyProducerTheme(element, widget.theme, false) })
	if !widget.IsRunning() && len(widget.history) > 0 {
		entry := widget.history[widget.historyIndex]
		widget.setText(renderExecutionResultWithLocale(entry.Suite, entry.Response, entry.Err, widget.bodyViewMode, widget.locale))
	}
	widget.updateTitle()
}

func applyProducerTheme(element *tview.TextView, uiTheme theme.ProducerTheme, focused bool) {
	element.SetTitleColor(uiTheme.Title).SetBackgroundColor(uiTheme.Background).SetBorderColor(uiTheme.Border)
	element.SetTextColor(uiTheme.Foreground)
	if focused {
		element.SetTitleColor(uiTheme.TitleFocus).SetBackgroundColor(uiTheme.BackgroundFocus).SetBorderColor(uiTheme.BorderFocus)
	}
}
