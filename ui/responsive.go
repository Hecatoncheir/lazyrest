package ui

import (
	"fmt"
	"strings"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (application *Application) updateResponsiveUI(screen tcell.Screen) {
	if screen == nil || application.Workspace == nil || application.Footer == nil {
		return
	}
	width, height := screen.Size()
	focused := application.focusedMainPane()
	application.Workspace.Resize(width, focused)
	application.Footer.Resize(width)
	application.Footer.UpdateHints(application.footerHints(width, focused))
	for _, overlay := range application.overlayFrames {
		overlay.Resize(width, height)
	}
}

func (application *Application) focusedMainPane() tview.Primitive {
	for _, primitive := range []tview.Primitive{
		application.HttpFilesTree.Element,
		application.Suites.Element,
		application.Suite.Element,
		application.Producer.Element,
	} {
		if primitive != nil && primitive.HasFocus() {
			return primitive
		}
	}
	return nil
}

func (application *Application) footerHints(width int, focused tview.Primitive) string {
	bindings := application.config.Keybindings
	translator := application.config.Locale
	if bindings == nil || translator == nil {
		return ""
	}

	hint := func(action keymap.Action, label string) string {
		key, _, _ := strings.Cut(bindings.Describe(action), " / ")
		return fmt.Sprintf("%s %s", key, translator.Text(label))
	}
	join := func(items ...string) string { return strings.Join(items, " · ") }

	if application.Model != nil && application.Model.CurrentOverlay() != OverlayNone {
		return hint(keymap.Back, "hint_close")
	}
	if application.HttpFilesTree.IsSearching() || application.Suites.IsSearching() || application.Producer.IsSearching() {
		return hint(keymap.SearchFinish, "hint_finish")
	}

	var contextual string
	switch focused {
	case application.HttpFilesTree.Element:
		contextual = join(hint(keymap.Open, "hint_open"), hint(keymap.Search, "hint_search"))
		if width >= 100 {
			contextual = join(contextual, hint(keymap.Reload, "hint_reload"))
		}
	case application.Suites.Element:
		contextual = join(hint(keymap.Open, "hint_select"), hint(keymap.Search, "hint_search"))
	case application.Suite.Element:
		contextual = join(hint(keymap.Run, "hint_run"), hint(keymap.Back, "hint_back"))
	case application.Producer.Element:
		contextual = join(hint(keymap.ToggleBody, "hint_view"), hint(keymap.ToggleHeaders, "hint_headers"))
		if width >= 80 {
			contextual = join(contextual, hint(keymap.ToggleRequest, "hint_request"))
		}
	default:
		contextual = hint(keymap.Help, "hint_help")
	}
	if width >= 120 {
		contextual = join(contextual, hint(keymap.Help, "hint_help"), hint(keymap.CommandPalette, "hint_commands"))
	}
	return contextual
}
