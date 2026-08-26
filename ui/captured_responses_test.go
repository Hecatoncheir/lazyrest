package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/locale"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestRenderCapturedResponsesUsesSafeRelativeSummaries(t *testing.T) {
	root := t.TempDir()
	text := renderCapturedResponses([]parserhttp.CapturedResponse{{
		SourceFilePath: filepath.Join(root, "requests", "auth.http"),
		Name:           "login\nrequest",
		Status:         "200 OK",
		HeaderCount:    3,
		BodyBytes:      128,
	}}, root, locale.English())

	for _, expected := range []string{"requests/auth.http", "login request", "200 OK", "3 headers", "128 body bytes"} {
		if !strings.Contains(text, expected) {
			t.Errorf("rendered captures %q do not contain %q", text, expected)
		}
	}
	if strings.Contains(text, root) {
		t.Fatalf("rendered captures exposed the absolute project root: %q", text)
	}
}

func TestCommandPaletteOpensCapturedResponsesOverlay(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	palette := application.CommandPalette
	index := -1
	for item := 0; item < palette.GetItemCount(); item++ {
		main, _ := palette.GetItemText(item)
		if main == application.config.Locale.Text("captured_responses") {
			index = item
			break
		}
	}
	if index < 0 {
		t.Fatal("captured responses command is missing")
	}
	palette.SetCurrentItem(index)
	palette.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(primitive tview.Primitive) {
		application.Element.SetFocus(primitive)
	})

	if application.Model.CurrentOverlay() != OverlayCaptured || application.Element.GetFocus() != application.Captured {
		t.Fatal("captured responses overlay was not opened and focused")
	}
	if !strings.Contains(application.Captured.GetText(false), application.config.Locale.Text("no_captured_responses")) {
		t.Fatalf("unexpected empty captured responses view: %q", application.Captured.GetText(false))
	}
}
