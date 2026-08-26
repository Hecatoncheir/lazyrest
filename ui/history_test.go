package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestCommandPaletteOpensEmptyHistoryWindow(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	palette := application.CommandPalette
	index := -1
	for item := 0; item < palette.GetItemCount(); item++ {
		main, _ := palette.GetItemText(item)
		if main == application.config.Locale.Text("history_window") {
			index = item
			break
		}
	}
	if index < 0 {
		t.Fatal("history command is missing")
	}
	palette.SetCurrentItem(index)
	palette.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(primitive tview.Primitive) {
		application.Element.SetFocus(primitive)
	})

	if application.Model.CurrentOverlay() != OverlayHistory || application.Element.GetFocus() != application.History {
		t.Fatal("history window was not opened and focused")
	}
	main, _ := application.History.GetItemText(0)
	if main != application.config.Locale.Text("no_history_entries") {
		t.Fatalf("unexpected empty history item: %q", main)
	}
}

func TestHistoryEntryRenderingUsesSafeMetadata(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	summary := producer.HistorySummary{
		Name:      "Create\nuser",
		Method:    "POST",
		Status:    "201 Created",
		Duration:  1500 * time.Microsecond,
		CreatedAt: time.Date(2026, time.August, 26, 12, 30, 0, 0, time.Local),
	}
	title := historyEntryTitle(summary, "Unnamed")
	details := historyEntryDetails(summary, t.TempDir(), application)
	for _, expected := range []string{"POST Create user", "201 Created", "1.5ms", "2026-08-26 12:30:00"} {
		if !strings.Contains(title+" "+details, expected) {
			t.Errorf("rendered history %q / %q does not contain %q", title, details, expected)
		}
	}
}
