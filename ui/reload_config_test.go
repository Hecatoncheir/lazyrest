package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestReloadConfigurationAppliesLanguageThemeAndKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	application := BuildApplication(t.TempDir(), Config{ConfigPath: path})
	contents := `language: ru
keybindings:
  command_palette: ["x", ":"]
theme:
  panel_background: "#010203"
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	application.reloadConfiguration()

	if application.config.Locale.Language() != "ru" {
		t.Fatal("language was not reloaded")
	}
	if !application.config.Keybindings.Matches(keymap.CommandPalette, tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)) {
		t.Fatal("keybindings were not reloaded")
	}
	if got := application.HttpFilesTree.Element.(*tview.TreeView).GetTitle(); got != "Файлы — загрузка" {
		t.Fatalf("localized title was not applied: %q", got)
	}
	if application.theme.Tree.Background == theme.NewDefault().Tree.Background {
		t.Fatal("theme was not reloaded")
	}
}

func TestCommandPaletteOpensFromConfiguredKey(t *testing.T) {
	bindings, err := keymap.New(map[string][]string{"command_palette": {"x", ":"}})
	if err != nil {
		t.Fatal(err)
	}
	application := BuildApplication(t.TempDir(), Config{Keybindings: bindings})
	event := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	if returned := onInputCallback(application)(event); returned != nil {
		t.Fatal("palette event was not consumed")
	}
	if application.Model.CurrentOverlay() != OverlayCommandPalette || application.CommandPalette == nil {
		t.Fatal("command palette was not opened")
	}
}
