package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestReloadConfigurationAppliesLanguageThemeAndKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	application := BuildApplication(t.TempDir(), Config{ConfigPath: path})
	contents := `language: ru
history: full
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
	if application.Producer.HistoryMode() != producer.HistoryFull {
		t.Fatal("history mode was not reloaded")
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

func TestThemePickerAppliesPreset(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	before := application.theme.Background
	commandPalette := application.CommandPalette
	themePicker := application.ThemePicker

	application.openOverlay(OverlayThemePicker)
	if application.Model.CurrentOverlay() != OverlayThemePicker || application.ThemePicker == nil {
		t.Fatal("theme picker was not opened")
	}
	application.closeOverlay()
	if application.Element.GetFocus() != application.HttpFilesTree.Element {
		t.Fatalf("focus was not restored to active panel: %T", application.Element.GetFocus())
	}
	application.selectThemePreset("monokai")
	if application.theme.Background == before || application.config.Theme.Background != application.theme.Background {
		t.Fatal("selected theme was not applied consistently")
	}
	treeView := application.HttpFilesTree.Element.(*tview.TreeView)
	if treeView.GetBackgroundColor() != application.theme.Tree.BackgroundFocus {
		t.Fatalf("focused panel kept stale background: got %v, want %v", treeView.GetBackgroundColor(), application.theme.Tree.BackgroundFocus)
	}
	if application.Element.GetFocus() != application.HttpFilesTree.Element {
		t.Fatal("active panel lost focus after applying theme")
	}
	if application.Pages.GetBackgroundColor() != application.theme.Background || application.Layout.Element.GetBackgroundColor() != application.theme.Background {
		t.Fatal("top-level containers kept stale theme")
	}
	if application.Workspace.Element.(*tview.Flex).GetBackgroundColor() != application.theme.Background {
		t.Fatal("workspace container kept stale theme")
	}
	if application.CommandPalette != commandPalette || application.ThemePicker != themePicker {
		t.Fatal("theme selection rebuilt overlays and disturbed focus state")
	}
	if application.Footer == nil {
		t.Fatal("footer was not preserved after applying theme")
	}
}
