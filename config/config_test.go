package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
)

func TestLoadCombinedConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	contents := `language: ru
languages:
  ru:
    files: Мои файлы
keybindings:
  command_palette: [":", "ctrl+p"]
theme:
  background: "#010203"
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	settings, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Locale.Text("files") != "Мои файлы" {
		t.Fatal("locale was not loaded")
	}
	if !settings.Keybindings.Matches(keymap.CommandPalette, tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone)) {
		t.Fatal("keybindings were not loaded")
	}
	if settings.Theme.Background == tcell.ColorDefault {
		t.Fatal("theme was not loaded")
	}
}
