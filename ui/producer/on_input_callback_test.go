package producer

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestResponseSearchMovesCyclicallyThroughMatches(t *testing.T) {
	widget := buildSearchTestProducer()
	widget.setText("before\nNeedle one\nbetween\nneedle two\nafter\nNEEDLE three")
	handler := onInputCallback(widget)

	handler(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	for _, r := range "needle" {
		handler(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	assertSearchPosition(t, widget, 0, 1, "1/3")

	handler(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	assertSearchPosition(t, widget, 1, 3, "2/3")
	handler(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	assertSearchPosition(t, widget, 2, 5, "3/3")
	handler(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	assertSearchPosition(t, widget, 0, 1, "1/3")
	handler(tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone))
	assertSearchPosition(t, widget, 2, 5, "3/3")
}

func TestResponseSearchRebuildsMatchesWhenTextChanges(t *testing.T) {
	widget := buildSearchTestProducer()
	widget.searchQuery = "ok"
	widget.setText("ok\nno\nOK")
	assertSearchPosition(t, widget, 0, 0, "1/2")

	widget.moveToSearchMatch(1)
	widget.setText("missing")
	if widget.searchIndex != -1 || len(widget.searchMatches) != 0 {
		t.Fatalf("stale matches remain: index=%d matches=%v", widget.searchIndex, widget.searchMatches)
	}
	if title := widget.Element.(*tview.TextView).GetTitle(); !strings.Contains(title, "0/0") {
		t.Fatalf("zero-match count is missing from title %q", title)
	}
}

func TestProducerRequestActionsUseConfiguredBindings(t *testing.T) {
	reruns := 0
	curlCopies := 0
	widget := New()
	widget.Build(Parameters{
		Keybindings:            keymap.Default(),
		Locale:                 locale.English(),
		OnRerunRequestCallback: func() { reruns++ },
		OnCopyAsCurlCallback:   func() { curlCopies++ },
	})
	handler := onInputCallback(widget)
	handler(tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone))
	handler(tcell.NewEventKey(tcell.KeyRune, 'C', tcell.ModNone))
	if reruns != 1 || curlCopies != 1 {
		t.Fatalf("unexpected request actions: reruns=%d curl copies=%d", reruns, curlCopies)
	}
}

func buildSearchTestProducer() *Producer {
	widget := New()
	widget.Build(Parameters{Keybindings: keymap.Default(), Locale: locale.English()})
	return widget
}

func assertSearchPosition(t *testing.T, widget *Producer, index, row int, count string) {
	t.Helper()
	if widget.searchIndex != index {
		t.Fatalf("search index %d, want %d", widget.searchIndex, index)
	}
	actualRow, _ := widget.Element.(*tview.TextView).GetScrollOffset()
	if actualRow != row {
		t.Fatalf("search row %d, want %d", actualRow, row)
	}
	if title := widget.Element.(*tview.TextView).GetTitle(); !strings.Contains(title, count) {
		t.Fatalf("title %q does not contain %q", title, count)
	}
}
