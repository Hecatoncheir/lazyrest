package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestCommandPaletteOpensEnvironmentPicker(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, environment.DefaultPublicFile), []byte(`{"staging":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	application := BuildApplication(root, Config{})
	index := commandPaletteItem(application, application.config.Locale.Text("choose_environment"))
	if index < 0 {
		t.Fatal("choose environment command is missing")
	}
	application.CommandPalette.SetCurrentItem(index)
	application.CommandPalette.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(primitive tview.Primitive) {
		application.Element.SetFocus(primitive)
	})

	if application.Model.CurrentOverlay() != OverlayEnvironmentPicker || application.Element.GetFocus() != application.EnvironmentPicker {
		t.Fatal("environment picker was not opened and focused")
	}
	items := map[string]bool{}
	for item := 0; item < application.EnvironmentPicker.GetItemCount(); item++ {
		name, _ := application.EnvironmentPicker.GetItemText(item)
		items[name] = true
	}
	if !items[application.config.Locale.Text("base_environment")] || !items["staging"] {
		t.Fatalf("environment picker has unexpected items: %v", items)
	}
}

func commandPaletteItem(application *Application, name string) int {
	for item := 0; item < application.CommandPalette.GetItemCount(); item++ {
		main, _ := application.CommandPalette.GetItemText(item)
		if main == name {
			return item
		}
	}
	return -1
}
