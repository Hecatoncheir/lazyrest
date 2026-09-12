package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestCommandPaletteFiltersAndRunsTheSelectedCommand(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	application.openOverlay(OverlayCommandPalette)
	handler := onInputCallback(application)

	handler(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	for _, character := range "theme" {
		handler(tcell.NewEventKey(tcell.KeyRune, character, tcell.ModNone))
	}
	if application.CommandPalette.GetItemCount() != 1 {
		t.Fatalf("filtered command count %d, want 1", application.CommandPalette.GetItemCount())
	}
	label, _ := application.CommandPalette.GetItemText(0)
	if label != application.config.Locale.Text("choose_theme") {
		t.Fatalf("unexpected filtered command %q", label)
	}
	if title := application.CommandPalette.GetTitle(); !strings.Contains(title, "/theme (1/") {
		t.Fatalf("filter and match count are missing from title %q", title)
	}

	event := handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if event == nil {
		t.Fatal("enter was not forwarded to the filtered list")
	}
	application.CommandPalette.InputHandler()(event, func(primitive tview.Primitive) {
		application.Element.SetFocus(primitive)
	})
	if application.Model.CurrentOverlay() != OverlayThemePicker {
		t.Fatal("filtered command did not open the theme picker")
	}
}

func TestCommandPaletteSearchHasAGuidedEmptyStateAndLayeredEscape(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	application.openOverlay(OverlayCommandPalette)
	handler := onInputCallback(application)
	handler(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	for _, character := range "qqq" {
		handler(tcell.NewEventKey(tcell.KeyRune, character, tcell.ModNone))
	}

	message, _ := application.CommandPalette.GetItemText(0)
	if !strings.Contains(message, "qqq") || application.commandMatches != 0 {
		t.Fatalf("unexpected empty state %q with %d matches", message, application.commandMatches)
	}
	if application.Model.CurrentOverlay() != OverlayCommandPalette {
		t.Fatal("quit key closed the palette while it was part of a search query")
	}

	handler(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if application.Model.CurrentOverlay() != OverlayCommandPalette || application.commandQuery != "" {
		t.Fatal("first escape did not clear the filter while keeping the palette open")
	}
	handler(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if application.Model.CurrentOverlay() != OverlayNone {
		t.Fatal("second escape did not close the command palette")
	}
}

func TestThemePickerMarksTheActivePreset(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	application.openOverlay(OverlayThemePicker)

	marked := ""
	for index := 0; index < application.ThemePicker.GetItemCount(); index++ {
		label, _ := application.ThemePicker.GetItemText(index)
		if strings.HasPrefix(label, "✓ ") {
			if marked != "" {
				t.Fatalf("multiple active themes are marked: %q and %q", marked, label)
			}
			marked = label
		}
	}
	if marked != "✓ gruvbox" {
		t.Fatalf("active theme marker %q, want gruvbox", marked)
	}
}
