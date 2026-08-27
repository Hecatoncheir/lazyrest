package config

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/gdamore/tcell/v2"
)

func TestProjectHistoryPathIsStableAndScopedByProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	firstRoot := t.TempDir()
	secondRoot := t.TempDir()

	first, err := ProjectHistoryPath(firstRoot)
	if err != nil {
		t.Fatal(err)
	}
	firstAgain, err := ProjectHistoryPath(filepath.Join(firstRoot, "."))
	if err != nil {
		t.Fatal(err)
	}
	second, err := ProjectHistoryPath(secondRoot)
	if err != nil {
		t.Fatal(err)
	}
	if first != firstAgain {
		t.Fatalf("equivalent project roots produced different history paths: %q and %q", first, firstAgain)
	}
	link := filepath.Join(t.TempDir(), "project-link")
	if err := os.Symlink(firstRoot, link); err == nil {
		throughLink, linkErr := ProjectHistoryPath(link)
		if linkErr != nil {
			t.Fatal(linkErr)
		}
		if throughLink != first {
			t.Fatalf("symlinked project root produced different history path: %q and %q", first, throughLink)
		}
	}
	if first == second {
		t.Fatalf("different projects share history path %q", first)
	}
	wantDirectory := filepath.Join(home, ".config", "lazyrest", "history")
	if filepath.Dir(first) != wantDirectory {
		t.Fatalf("history directory %q, want %q", filepath.Dir(first), wantDirectory)
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}\.json$`).MatchString(filepath.Base(first)) {
		t.Fatalf("history file does not use a stable project id: %q", first)
	}
}

func TestHistoryPathStillLocatesLegacySharedHistory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := HistoryPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".config", "lazyrest", "history.json")
	if path != want {
		t.Fatalf("legacy history path %q, want %q", path, want)
	}
}

func TestLoadCombinedConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	contents := `language: ru
languages:
  ru:
    files: Мои файлы
keybindings:
  command_palette: [":", "ctrl+p"]
theme:
  preset: nord
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
	if settings.Theme.Background == tcell.ColorDefault || settings.Theme.Background == theme.NewDefault().Background {
		t.Fatal("theme preset was not loaded")
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

func TestLoadRefusesAnUnknownKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	// "keybinding" is a typo for "keybindings" and used to be dropped without
	// a word, leaving the defaults in place.
	contents := "language: en\nkeybinding:\n  quit: [\"x\"]\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("an unknown key was accepted")
	}
	if !strings.Contains(err.Error(), "keybinding") {
		t.Fatalf("the error does not name the key: %v", err)
	}
}

func TestLoadAcceptsAnEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	settings, err := Load(path)
	if err != nil {
		t.Fatalf("an empty file was refused: %v", err)
	}
	if settings.Keybindings == nil {
		t.Fatal("the defaults were not applied")
	}
}

func TestLoadFilesAddsUpIgnoreLists(t *testing.T) {
	directory := t.TempDir()
	user := filepath.Join(directory, "user.yml")
	project := filepath.Join(directory, "project.yml")
	if err := os.WriteFile(user, []byte("ignore:\n  - fixtures\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A project says what else to skip without losing what the user chose.
	if err := os.WriteFile(project, []byte("ignore:\n  - generated\n  - fixtures\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	settings, err := LoadFiles([]string{user, project})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(settings.Ignore, []string{"fixtures", "generated"}) {
		t.Fatalf("unexpected ignore list: %#v", settings.Ignore)
	}
}

func TestHistoryModeDefaultsToMetadataAndLayersCanOverrideIt(t *testing.T) {
	directory := t.TempDir()
	defaults, err := Load(filepath.Join(directory, "missing.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !defaults.HistoryMetadata || defaults.Document.History != HistoryMetadata {
		t.Fatalf("history default is not metadata-only: %+v", defaults.Document)
	}

	user := filepath.Join(directory, "user.yml")
	project := filepath.Join(directory, "project.yml")
	if err := os.WriteFile(user, []byte("history: full\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	full, err := Load(user)
	if err != nil {
		t.Fatal(err)
	}
	if full.HistoryMetadata || full.Document.History != HistoryFull {
		t.Fatalf("full history mode was not loaded: %+v", full.Document)
	}
	if err := os.WriteFile(project, []byte("history: metadata\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	metadata, err := LoadFiles([]string{user, project})
	if err != nil {
		t.Fatal(err)
	}
	if !metadata.HistoryMetadata {
		t.Fatal("later metadata history layer did not override full mode")
	}
}

func TestLoadRejectsUnknownHistoryMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("history: everything\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "history") {
		t.Fatalf("unknown history mode was accepted: %v", err)
	}
}
