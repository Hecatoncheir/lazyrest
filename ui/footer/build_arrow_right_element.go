package footer

import (
	"fmt"

	"github.com/Hecatoncheir/lazyrest/ui/symbols"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func buildArrowRightElement(
	background,
	foreground tcell.Color,
) (
	element tview.Primitive,
	size int,
) {
	return buildArrowElement(symbols.ArrowRight, background, foreground)
}

func buildArrowLeftElement(
	background,
	foreground tcell.Color,
) (
	element tview.Primitive,
	size int,
) {
	return buildArrowElement(symbols.ArrowLeft, background, foreground)
}

func buildArrowElement(
	symbol rune,
	background,
	foreground tcell.Color,
) (
	element tview.Primitive,
	size int,
) {
	arrowText := fmt.Sprintf("%c", symbol)
	arrowStyle := tcell.Style{}.
		Background(background).
		Foreground(foreground)
	arrowView := tview.NewTextView().
		SetText(arrowText).
		SetTextStyle(arrowStyle)
	return arrowView, 1
}
