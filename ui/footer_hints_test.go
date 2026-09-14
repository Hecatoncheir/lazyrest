package ui

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/environment"
)

func hintsApplication(t *testing.T) *Application {
	t.Helper()
	return BuildApplication(t.TempDir(), Config{Environment: environment.Config{Name: "test"}})
}

// With no pane focused the default branch already offers help, and the wide
// layout used to append it a second time: "? help · ? help · : commands".
func TestFooterHintsDoNotRepeatTheHelpHint(t *testing.T) {
	application := hintsApplication(t)

	for _, width := range []int{60, 80, 100, 120, 160} {
		hints := application.footerHints(width, nil)
		if got := strings.Count(hints, "help"); got > 1 {
			t.Errorf("width %d: help appears %d times in %q", width, got, hints)
		}
	}
}

func TestFooterHintsStillOfferHelpAndCommandsWhenWide(t *testing.T) {
	application := hintsApplication(t)
	hints := application.footerHints(160, nil)

	if !strings.Contains(hints, "help") {
		t.Errorf("hints = %q, want help offered", hints)
	}
	if !strings.Contains(hints, "commands") {
		t.Errorf("hints = %q, want the command palette offered", hints)
	}
}

// A focused pane keeps its own hints plus one help hint, not two.
func TestFooterHintsForAFocusedPaneOfferHelpOnce(t *testing.T) {
	application := hintsApplication(t)
	hints := application.footerHints(160, application.Suite.Element)

	if got := strings.Count(hints, "help"); got != 1 {
		t.Errorf("help appears %d times in %q, want once", got, hints)
	}
}

func TestFooterHintsForTheStreamPane(t *testing.T) {
	application := hintsApplication(t)
	hints := application.footerHints(160, application.Stream.Element)

	for _, expected := range []string{"follow", "send", "clear"} {
		if !strings.Contains(hints, expected) {
			t.Errorf("hints = %q, want %q offered for the stream pane", hints, expected)
		}
	}
	if got := strings.Count(hints, "help"); got != 1 {
		t.Errorf("help appears %d times in %q, want once", got, hints)
	}
}
