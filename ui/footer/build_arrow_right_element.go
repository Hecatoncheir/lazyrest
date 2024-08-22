package footer

import (
	"fmt"

	"lazyrest/ui/symbols"

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
	arrowRightText := fmt.Sprintf("%c", symbols.ArrowRight)
	arrowRightStyle := tcell.Style{}.
		Background(background).
		Foreground(foreground)
	arrowRightView := tview.NewTextView().
		SetText(arrowRightText).
		SetTextStyle(arrowRightStyle)
	marginRight := 2
	return arrowRightView, marginRight
}
