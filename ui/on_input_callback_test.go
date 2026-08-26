package ui

import (
	"testing"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestCtrlNavigation(t *testing.T) {
	tests := []struct {
		name string
		from func(*Application) tview.Primitive
		key  tcell.Key
		to   func(*Application) tview.Primitive
	}{
		{name: "Files to Suites", from: filesElement, key: tcell.KeyCtrlL, to: suitesElement},
		{name: "Suites to Files", from: suitesElement, key: tcell.KeyCtrlH, to: filesElement},
		{name: "Suites to Suite", from: suitesElement, key: tcell.KeyCtrlJ, to: suiteElement},
		{name: "Suites to Producer", from: suitesElement, key: tcell.KeyCtrlL, to: producerElement},
		{name: "Suite to Files", from: suiteElement, key: tcell.KeyCtrlH, to: filesElement},
		{name: "Suite to Suites", from: suiteElement, key: tcell.KeyCtrlK, to: suitesElement},
		{name: "Suite to Producer", from: suiteElement, key: tcell.KeyCtrlL, to: producerElement},
		{name: "Producer to Suite", from: producerElement, key: tcell.KeyCtrlH, to: suiteElement},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			application := newNavigationTestApplication()
			application.Element.SetFocus(test.from(application))

			event := tcell.NewEventKey(test.key, 0, tcell.ModCtrl)
			if returned := onInputCallback(application)(event); returned != nil {
				t.Fatal("handled navigation event was not consumed")
			}

			if focused := application.Element.GetFocus(); focused != test.to(application) {
				t.Fatalf("unexpected focused element: got %T, want %T", focused, test.to(application))
			}
		})
	}
}

func TestCtrlHNavigationAcceptsBackspaceEvent(t *testing.T) {
	application := newNavigationTestApplication()
	application.Element.SetFocus(application.Producer.Element)

	event := tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone)
	if returned := onInputCallback(application)(event); returned != nil {
		t.Fatal("handled navigation event was not consumed")
	}

	if focused := application.Element.GetFocus(); focused != application.Suite.Element {
		t.Fatalf("unexpected focused element: got %T, want %T", focused, application.Suite.Element)
	}
}

func TestConfiguredNavigationSupportsMultipleKeys(t *testing.T) {
	bindings, err := keymap.New(map[string][]string{"focus_right": {"right", "ctrl+l"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyCtrlL, 0, tcell.ModCtrl),
	} {
		application := newNavigationTestApplication()
		application.config.Keybindings = bindings
		application.Element.SetFocus(application.Suites.Element)
		if returned := onInputCallback(application)(event); returned != nil {
			t.Fatal("configured navigation event was not consumed")
		}
		if focused := application.Element.GetFocus(); focused != application.Producer.Element {
			t.Fatalf("unexpected focused element: got %T", focused)
		}
	}
}

func TestUnsupportedCtrlDirectionKeepsFocus(t *testing.T) {
	application := newNavigationTestApplication()
	application.Element.SetFocus(application.Suites.Element)

	event := tcell.NewEventKey(tcell.KeyCtrlK, 0, tcell.ModCtrl)
	if returned := onInputCallback(application)(event); returned != nil {
		t.Fatal("navigation event was not consumed")
	}

	if focused := application.Element.GetFocus(); focused != application.Suites.Element {
		t.Fatalf("unexpected focused element: got %T, want %T", focused, application.Suites.Element)
	}
}

func TestQuitBindingClosesAnOverlayBeforeQuitting(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	application.openOverlay(OverlayHelp)

	event := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
	if returned := onInputCallback(application)(event); returned != nil {
		t.Fatal("handled q event was not consumed")
	}
	if overlay := application.Model.CurrentOverlay(); overlay != OverlayNone {
		t.Fatalf("q did not close the overlay: %v", overlay)
	}
}

func TestCtrlCRemainsAnUnconditionalQuitWithAnOverlay(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	application.openOverlay(OverlayHelp)

	event := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl)
	if returned := onInputCallback(application)(event); returned != nil {
		t.Fatal("handled Ctrl+C event was not consumed")
	}
	if overlay := application.Model.CurrentOverlay(); overlay != OverlayHelp {
		t.Fatalf("Ctrl+C was handled as an overlay close: %v", overlay)
	}
}

func newNavigationTestApplication() *Application {
	return &Application{
		Element:       tview.NewApplication(),
		HttpFilesTree: &tree.Tree{Element: tview.NewBox()},
		Suites:        &suites.Suites{Element: tview.NewBox()},
		Suite:         &suite.Suite{Element: tview.NewBox()},
		Producer:      &producer.Producer{Element: tview.NewBox()},
	}
}

func filesElement(application *Application) tview.Primitive {
	return application.HttpFilesTree.Element
}

func suitesElement(application *Application) tview.Primitive {
	return application.Suites.Element
}

func suiteElement(application *Application) tview.Primitive {
	return application.Suite.Element
}

func producerElement(application *Application) tview.Primitive {
	return application.Producer.Element
}
