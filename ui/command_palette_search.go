package ui

import (
	"unicode"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
)

func (application *Application) handleCommandPaletteInput(event *tcell.EventKey) bool {
	bindings := application.config.Keybindings
	if bindings == nil {
		bindings = keymap.Default()
	}
	if application.commandSearchMode || application.commandQuery != "" {
		if bindings.Matches(keymap.Back, event) {
			application.resetCommandPaletteSearch()
			return true
		}
	}
	if !application.commandSearchMode {
		if bindings.Matches(keymap.Search, event) {
			application.commandSearchMode = true
			application.renderCommandPalette()
			return true
		}
		return false
	}

	switch event.Key() {
	case tcell.KeyEnter:
		if application.commandMatches == 0 {
			return true
		}
		return false
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		query := []rune(application.commandQuery)
		if len(query) > 0 {
			application.commandQuery = string(query[:len(query)-1])
		}
		application.renderCommandPalette()
		return true
	}
	if event.Rune() == 0 || unicode.IsControl(event.Rune()) || event.Modifiers()&(tcell.ModCtrl|tcell.ModAlt|tcell.ModMeta) != 0 {
		return false
	}
	application.commandQuery += string(event.Rune())
	application.renderCommandPalette()
	return true
}
