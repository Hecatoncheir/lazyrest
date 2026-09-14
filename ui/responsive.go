package ui

import (
	"fmt"
	"strings"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
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
	panes := []tview.Primitive{
		application.HttpFilesTree.Element,
		application.Suites.Element,
		application.Suite.Element,
		application.Producer.Element,
	}
	if application.Stream != nil {
		panes = append(panes, application.Stream.Element)
	}
	for _, primitive := range panes {
		if primitive != nil && primitive.HasFocus() {
			return primitive
		}
	}
	return nil
}

// hint writes one footer hint: the key a binding uses, then what it does.
type hint struct {
	bindings   *keymap.Bindings
	translator *locale.Translator
}

func (write hint) of(action keymap.Action, label string) string {
	key, _, _ := strings.Cut(write.bindings.Describe(action), " / ")
	return fmt.Sprintf("%s %s", key, write.translator.Text(label))
}

func (write hint) join(items ...string) string {
	return strings.Join(items, " · ")
}

// footerHints says what the keys do right now. An overlay answers first, then a
// search in progress, then the focused pane.
func (application *Application) footerHints(width int, focused tview.Primitive) string {
	bindings := application.config.Keybindings
	translator := application.config.Locale
	if bindings == nil || translator == nil {
		return ""
	}
	write := hint{bindings: bindings, translator: translator}

	if overlay, open := application.overlayOnScreen(); open {
		return application.overlayHints(overlay, write)
	}
	if application.isSearching() {
		return write.of(keymap.SearchFinish, "hint_finish")
	}
	return application.paneHints(width, focused, write)
}

func (application *Application) overlayHints(overlay Overlay, write hint) string {
	switch overlay {
	case OverlayHistory:
		return clearableHints(write, keymap.ClearHistory,
			application.confirmHistoryClear, len(application.Producer.HistorySummaries()) == 0)
	case OverlayCaptured:
		return clearableHints(write, keymap.ClearCaptured,
			application.confirmCapturedClear, len(application.Producer.CapturedResponses()) == 0)
	case OverlayCommandPalette:
		if application.commandSearchMode || application.commandQuery != "" {
			return write.join(write.of(keymap.Open, "hint_select"), write.of(keymap.Back, "hint_clear"))
		}
		return write.join(
			write.of(keymap.Search, "hint_filter"),
			write.of(keymap.Open, "hint_select"),
			write.of(keymap.Back, "hint_close"),
		)
	case OverlayHelp:
		return write.join(write.of(keymap.Help, "hint_toggle"), write.of(keymap.Back, "hint_close"))
	case OverlayDiagnostics:
		return write.join(write.of(keymap.Diagnostics, "hint_toggle"), write.of(keymap.Back, "hint_close"))
	case OverlayThemePicker, OverlayEnvironmentPicker:
		return write.join(write.of(keymap.Open, "hint_select"), write.of(keymap.Back, "hint_close"))
	case OverlaySaveResponse:
		if application.saveOverwritePath != "" {
			return write.join(write.of(keymap.Open, "hint_confirm"), write.of(keymap.Back, "hint_cancel"))
		}
		return write.join(write.of(keymap.Open, "hint_save"), write.of(keymap.Back, "hint_close"))
	default:
		return write.of(keymap.Back, "hint_close")
	}
}

// clearableHints serves the windows that list something and can empty it. They
// differ only in which action clears them.
func clearableHints(write hint, clear keymap.Action, confirming, empty bool) string {
	if confirming {
		return write.join(write.of(clear, "hint_confirm"), write.of(keymap.Back, "hint_cancel"))
	}
	if empty {
		return write.of(keymap.Back, "hint_close")
	}
	return write.join(write.of(clear, "hint_clear"), write.of(keymap.Back, "hint_close"))
}

// paneHints says what the focused pane offers, and adds the two bindings worth
// advertising whenever the footer has room for them.
func (application *Application) paneHints(width int, focused tview.Primitive, write hint) string {
	help := write.of(keymap.Help, "hint_help")
	contextual := application.contextualHints(width, focused, write, help)

	if !application.hasRoomForGlobalHints(width) {
		return contextual
	}
	// The fallback already offers help; appending it again is how the hint used
	// to appear twice.
	if contextual != help {
		contextual = write.join(contextual, help)
	}
	return write.join(contextual, write.of(keymap.CommandPalette, "hint_commands"))
}

// hasRoomForGlobalHints is wider before the first request has run, when the
// footer is the only place that says help and the palette exist.
func (application *Application) hasRoomForGlobalHints(width int) bool {
	if width >= 120 {
		return true
	}
	if width < 80 || application.Model == nil {
		return false
	}
	state := application.Model.Snapshot()
	return state.Request.Phase == PhaseIdle && state.Request.Outcome == OutcomeNone
}

func (application *Application) contextualHints(
	width int,
	focused tview.Primitive,
	write hint,
	help string,
) string {
	switch {
	case focused == application.HttpFilesTree.Element:
		hints := write.join(write.of(keymap.Open, "hint_open"), write.of(keymap.Search, "hint_search"))
		if width >= 100 {
			hints = write.join(hints, write.of(keymap.Reload, "hint_reload"))
		}
		return hints
	case focused == application.Suites.Element:
		return write.join(write.of(keymap.Open, "hint_select"), write.of(keymap.Search, "hint_search"))
	case focused == application.Suite.Element:
		return write.join(write.of(keymap.Run, "hint_run"), write.of(keymap.Back, "hint_back"))
	case focused == application.Producer.Element:
		hints := write.join(write.of(keymap.ToggleBody, "hint_view"), write.of(keymap.ToggleHeaders, "hint_headers"))
		if width >= 80 {
			hints = write.join(hints, write.of(keymap.ToggleRequest, "hint_request"))
		}
		return hints
	case application.Stream != nil && focused == application.Stream.Element:
		hints := write.join(write.of(keymap.StreamFollow, "hint_follow"), write.of(keymap.StreamSend, "hint_send"))
		if width >= 80 {
			hints = write.join(hints, write.of(keymap.StreamClear, "hint_clear"))
		}
		return hints
	default:
		return help
	}
}
