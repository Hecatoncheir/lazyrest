package ui

import "github.com/rivo/tview"

const overlayMargin = 1

type responsiveOverlay struct {
	Element   *tview.Flex
	columns   *tview.Flex
	primitive tview.Primitive
	maxWidth  int
	maxHeight int
}

func newResponsiveOverlay(primitive tview.Primitive, width, height int) *responsiveOverlay {
	columns := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(primitive, width, 0, true).
		AddItem(nil, 0, 1, false)
	rows := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(columns, height, 0, true).
		AddItem(nil, 0, 1, false)
	return &responsiveOverlay{
		Element: rows, primitive: primitive, columns: columns,
		maxWidth: width, maxHeight: height,
	}
}

func (overlay *responsiveOverlay) Resize(screenWidth, screenHeight int) {
	width := min(overlay.maxWidth, max(1, screenWidth-overlayMargin*2))
	height := min(overlay.maxHeight, max(1, screenHeight-overlayMargin*2))
	overlay.columns.ResizeItem(overlay.primitive, width, 0)
	overlay.Element.ResizeItem(overlay.columns, height, 0)
}
