package tree

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func (widget *Tree) ApplySettings(uiTheme theme.Theme, translator *locale.Translator, bindings *keymap.Bindings) {
	widget.theme = uiTheme.Tree
	widget.locale = translator
	widget.keybindings = bindings
	element := widget.Element.(*tview.TreeView)
	box := element.Box
	applyTreeBoxTheme(box, widget.theme, element.HasFocus())
	box.SetFocusFunc(func() { applyTreeBoxTheme(box, widget.theme, true) })
	box.SetBlurFunc(func() { applyTreeBoxTheme(box, widget.theme, false) })
	applyNodeTheme(element.GetRoot(), widget.theme)
	widget.updateTitle()
}

func applyTreeBoxTheme(box *tview.Box, uiTheme theme.TreeTheme, focused bool) {
	box.SetTitleColor(uiTheme.Title).SetBackgroundColor(uiTheme.Background).SetBorderColor(uiTheme.Border)
	if focused {
		box.SetTitleColor(uiTheme.TitleFocus).SetBackgroundColor(uiTheme.BackgroundFocus).SetBorderColor(uiTheme.BorderFocus)
	}
}

func applyNodeTheme(node *tview.TreeNode, uiTheme theme.TreeTheme) {
	if node == nil {
		return
	}
	if _, ok := node.GetReference().(finder.Directory); ok {
		node.SetColor(uiTheme.NodeDirectory.Foreground)
	} else {
		node.SetColor(uiTheme.Node.Foreground)
	}
	for _, child := range node.GetChildren() {
		applyNodeTheme(child, uiTheme)
	}
}
