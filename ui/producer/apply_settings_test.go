package producer

import (
	"testing"

	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestApplyProducerThemeUpdatesFilledTextBackground(t *testing.T) {
	uiTheme := theme.ProducerTheme{
		Background:      tcell.NewHexColor(0x282a36),
		BackgroundFocus: tcell.NewHexColor(0x3b3d4d),
	}
	for _, test := range []struct {
		name       string
		focused    bool
		background tcell.Color
	}{
		{name: "unfocused", background: uiTheme.Background},
		{name: "focused", focused: true, background: uiTheme.BackgroundFocus},
	} {
		t.Run(test.name, func(t *testing.T) {
			view := tview.NewTextView().SetText("response")
			view.SetBackgroundColor(tcell.NewHexColor(0x504945))
			view.SetRect(0, 0, 12, 3)
			applyProducerTheme(view, uiTheme, test.focused)

			if got := drawnTextBackground(t, view); got != test.background {
				t.Fatalf("filled text background is %v, want %v", got, test.background)
			}
		})
	}
}

func drawnTextBackground(t *testing.T, view *tview.TextView) tcell.Color {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(12, 3)
	view.Draw(screen)
	_, style, _ := screen.Get(0, 0)
	_, background, _ := style.Decompose()
	return background
}
