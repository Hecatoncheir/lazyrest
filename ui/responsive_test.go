package ui

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/keymap"
)

func TestOverlayFooterHintsExposeToggleActions(t *testing.T) {
	cases := []struct {
		name    string
		overlay Overlay
		key     string
	}{
		{name: "help", overlay: OverlayHelp, key: "?"},
		{name: "diagnostics", overlay: OverlayDiagnostics, key: "d"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			application := BuildApplication(t.TempDir(), Config{})
			application.openOverlay(testCase.overlay)
			hints := application.footerHints(120, nil)
			if !strings.Contains(hints, testCase.key+" "+application.config.Locale.Text("hint_toggle")) {
				t.Fatalf("footer hints %q do not expose %s toggle", hints, testCase.name)
			}
			if !strings.Contains(hints, application.config.Keybindings.Describe(keymap.Back)+" "+application.config.Locale.Text("hint_close")) {
				t.Fatalf("footer hints %q do not expose close action", hints)
			}
		})
	}
}
