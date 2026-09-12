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
		overlay := application.Model.CurrentOverlay()
		if overlay == OverlayHistory {
			if application.confirmHistoryClear {
				return join(hint(keymap.ClearHistory, "hint_confirm"), hint(keymap.Back, "hint_cancel"))
			}
			if len(application.Producer.HistorySummaries()) == 0 {
				return hint(keymap.Back, "hint_close")
			}
			return join(hint(keymap.ClearHistory, "hint_clear"), hint(keymap.Back, "hint_close"))
		}
		if overlay == OverlayCaptured {
			if application.confirmCapturedClear {
				return join(hint(keymap.ClearCaptured, "hint_confirm"), hint(keymap.Back, "hint_cancel"))
			}
			if len(application.Producer.CapturedResponses()) == 0 {
				return hint(keymap.Back, "hint_close")
			}
			return join(hint(keymap.ClearCaptured, "hint_clear"), hint(keymap.Back, "hint_close"))
		}
		if overlay == OverlayCommandPalette {
			if application.commandSearchMode || application.commandQuery != "" {
				return join(hint(keymap.Open, "hint_select"), hint(keymap.Back, "hint_clear"))
			}
			return join(hint(keymap.Search, "hint_filter"), hint(keymap.Open, "hint_select"), hint(keymap.Back, "hint_close"))
		}
		if overlay == OverlayHelp {
			return join(hint(keymap.Help, "hint_toggle"), hint(keymap.Back, "hint_close"))
		}
		if overlay == OverlayDiagnostics {
			return join(hint(keymap.Diagnostics, "hint_toggle"), hint(keymap.Back, "hint_close"))
		}
		if overlay == OverlayThemePicker || overlay == OverlayEnvironmentPicker {
			return join(hint(keymap.Open, "hint_select"), hint(keymap.Back, "hint_close"))
		}
		if overlay == OverlaySaveResponse {
			if application.saveOverwritePath != "" {
				return join(hint(keymap.Open, "hint_confirm"), hint(keymap.Back, "hint_cancel"))
			}
			return join(hint(keymap.Open, "hint_save"), hint(keymap.Back, "hint_close"))
		}
		return hint(keymap.Back, "hint_close")
	}
	if application.HttpFilesTree.IsSearching() || application.Suites.IsSearching() || application.Producer.IsSearching() {
		return hint(keymap.SearchFinish, "hint_finish")
	}
	onboarding := application.Model != nil && func() bool {
		state := application.Model.Snapshot()
		return state.Request.Phase == PhaseIdle && state.Request.Outcome == OutcomeNone
	}()

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
	if width >= 120 || (onboarding && width >= 80) {
		contextual = join(contextual, hint(keymap.Help, "hint_help"), hint(keymap.CommandPalette, "hint_commands"))
	}
	return contextual
}
