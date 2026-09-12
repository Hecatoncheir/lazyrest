package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestResponsiveOverlayUsesMaximumSizeWhenSpaceIsAvailable(t *testing.T) {
	assertOverlaySize(t, 120, 40, 84, 24, 84, 24)
}

func TestResponsiveOverlayKeepsAMarginInSmallTerminals(t *testing.T) {
	assertOverlaySize(t, 20, 10, 84, 24, 18, 8)
}

func assertOverlaySize(t *testing.T, screenWidth, screenHeight, maxWidth, maxHeight, wantWidth, wantHeight int) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(screenWidth, screenHeight)

	content := tview.NewTextView()
	overlay := newResponsiveOverlay(content, maxWidth, maxHeight)
	overlay.Resize(screenWidth, screenHeight)
	overlay.Element.SetRect(0, 0, screenWidth, screenHeight)
	overlay.Element.Draw(screen)

	_, _, width, height := content.GetRect()
	if width != wantWidth || height != wantHeight {
		t.Fatalf("overlay size %dx%d, want %dx%d", width, height, wantWidth, wantHeight)
	}
}
