package keymap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestBindingsSupportMultipleKeys(t *testing.T) {
	bindings, err := New(map[string][]string{"quit": {"x", "ctrl+q"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyCtrlQ, 0, tcell.ModCtrl),
	} {
		if !bindings.Matches(Quit, event) {
			t.Fatalf("expected event %s to match quit", event.Name())
		}
	}
	if bindings.Matches(Quit, tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)) {
		t.Fatal("configured keys must replace the default action keys")
	}
}

func TestBindingsSupportControlKeysFromNonLatinLayouts(t *testing.T) {
	bindings, err := New(map[string][]string{
		"focus_left":      {"ctrl+h", "ctrl+р"},
		"focus_down":      {"ctrl+j", "ctrl+о"},
		"focus_up":        {"ctrl+k", "ctrl+л"},
		"focus_right":     {"ctrl+l", "ctrl+д"},
		"command_palette": {":", "ctrl+p", "ctrl+з"},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		action Action
		key    rune
	}{
		{FocusLeft, 'р'},
		{FocusDown, 'о'},
		{FocusUp, 'л'},
		{FocusRight, 'д'},
		{CommandPalette, 'з'},
	}
	for _, test := range tests {
		event := tcell.NewEventKey(tcell.KeyRune, test.key, tcell.ModCtrl)
		if !bindings.Matches(test.action, event) {
			t.Errorf("expected ctrl+%c to match %s", test.key, test.action)
		}
		if bindings.Matches(test.action, tcell.NewEventKey(tcell.KeyRune, test.key, tcell.ModNone)) {
			t.Errorf("expected unmodified %c not to match %s", test.key, test.action)
		}
	}
}

func TestLoadMergesConfiguredAndDefaultBindings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("keybindings:\n  help: ['h', '?']\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bindings, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bindings.Matches(Help, tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)) {
		t.Fatal("configured help binding was not loaded")
	}
	if !bindings.Matches(Quit, tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)) {
		t.Fatal("unconfigured default binding was not preserved")
	}
}

func TestLoadRejectsUnknownActionsAndKeys(t *testing.T) {
	for _, contents := range []string{
		"keybindings:\n  unknown: ['x']\n",
		"keybindings:\n  quit: ['ctrl+shift+x']\n",
	} {
		path := filepath.Join(t.TempDir(), "config.yml")
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatalf("expected invalid config error for %q", contents)
		}
	}
}

func TestLoadDefaultUsesHomeConfigDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDirectory := filepath.Join(home, ".config", "lazyrest")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configDirectory, "config.yml")
	if err := os.WriteFile(path, []byte("keybindings:\n  help: ['f1', '?']\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bindings, loadedPath, err := LoadDefault()
	if err != nil {
		t.Fatal(err)
	}
	if loadedPath != path {
		t.Fatalf("unexpected config path: got %q, want %q", loadedPath, path)
	}
	if !bindings.Matches(Help, tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone)) {
		t.Fatal("function key binding was not loaded")
	}
}

func TestBindingsRejectContextualConflicts(t *testing.T) {
	_, err := New(map[string][]string{"help": {"x"}, "reload": {"x"}})
	if err == nil {
		t.Fatal("expected conflicting Files bindings to fail")
	}
	if _, err := New(map[string][]string{"run": {"r"}, "reload": {"r"}}); err != nil {
		t.Fatalf("bindings in different contexts must be allowed: %v", err)
	}
}

func TestDefaultResponseExportBindings(t *testing.T) {
	bindings := Default()
	for _, test := range []struct {
		action Action
		key    rune
	}{
		{CopyResponseBody, 'y'},
		{CopyResponse, 'Y'},
		{SaveResponse, 's'},
		{SaveFullResponse, 'S'},
	} {
		if !bindings.Matches(test.action, tcell.NewEventKey(tcell.KeyRune, test.key, tcell.ModNone)) {
			t.Errorf("%s does not match %q", test.action, test.key)
		}
	}
}

func TestDefaultViewportNavigationBindings(t *testing.T) {
	bindings := Default()
	if !bindings.Matches(HalfPageDown, tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl)) {
		t.Fatal("half_page_down does not match Ctrl+d")
	}
	if !bindings.Matches(HalfPageUp, tcell.NewEventKey(tcell.KeyCtrlU, 0, tcell.ModCtrl)) {
		t.Fatal("half_page_up does not match Ctrl+u")
	}

	z := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)
	if got := bindings.MatchesSequence(CenterView, []*tcell.EventKey{z}); got != SequencePrefix {
		t.Fatalf("first z must be a sequence prefix, got %v", got)
	}
	if got := bindings.MatchesSequence(CenterView, []*tcell.EventKey{z, z}); got != SequenceFull {
		t.Fatalf("zz must complete center_view, got %v", got)
	}
	if bindings.Matches(CenterView, z) {
		t.Fatal("a single z must not match the zz binding")
	}
}

func TestCenterViewSupportsConfiguredPrintableSequence(t *testing.T) {
	bindings, err := New(map[string][]string{"center_view": {"жж"}})
	if err != nil {
		t.Fatal(err)
	}
	zh := tcell.NewEventKey(tcell.KeyRune, 'ж', tcell.ModNone)
	if got := bindings.MatchesSequence(CenterView, []*tcell.EventKey{zh, zh}); got != SequenceFull {
		t.Fatalf("configured sequence did not match: %v", got)
	}
	if got := bindings.Map()[string(CenterView)]; len(got) != 1 || got[0] != "жж" {
		t.Fatalf("configured sequence was not preserved: %v", got)
	}
}

func TestBindingsRejectSequencesForSingleKeyActions(t *testing.T) {
	if _, err := New(map[string][]string{"quit": {"qq"}}); err == nil {
		t.Fatal("expected a sequence assigned to quit to fail")
	}
}

func TestBindingsRejectSequencePrefixConflicts(t *testing.T) {
	if _, err := New(map[string][]string{"help": {"z"}}); err == nil {
		t.Fatal("expected z to conflict with the zz center_view binding")
	}
}
