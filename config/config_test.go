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

func TestLoadFilesMergesLayersInPriorityOrder(t *testing.T) {
	directory := t.TempDir()
	user := filepath.Join(directory, "user.yml")
	project := filepath.Join(directory, "project.yml")
	explicit := filepath.Join(directory, "explicit.yml")
	files := map[string]string{
		user:     "language: ru\ntheme:\n  accent: '#111111'\n",
		project:  "languages:\n  ru:\n    files: Проект\ntheme:\n  accent: '#222222'\n",
		explicit: "language: es\nkeybindings:\n  help: ['h']\n",
	}
	for path, contents := range files {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	settings, err := LoadFiles([]string{user, project, explicit})
	if err != nil {
		t.Fatal(err)
	}
	if settings.Locale.Language() != "es" || settings.Document.Theme.Accent != "#222222" {
		t.Fatalf("layers were not merged in order: %+v", settings.Document)
	}
	if !settings.Keybindings.Matches(keymap.Help, tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)) {
		t.Fatal("explicit keybinding did not win")
	}
}

func TestGenerateDoesNotOverwriteExistingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lazyrest", "config.yml")
	if err := Generate(path); err != nil {
		t.Fatal(err)
	}
	if err := Generate(path); err == nil {
		t.Fatal("generate must not overwrite an existing config")
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("generated config is invalid: %v", err)
	}
}
