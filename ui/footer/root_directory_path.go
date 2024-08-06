package footer

import (
	"fmt"
	"lazyrest/ui/symbols"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func buildRootDirectoryPath(parameters Parameters) tview.Primitive {
	rootDirectoryPath := parameters.RootDirectoryPath
	rootDirectoryPathText := fmt.Sprintf("%v", rootDirectoryPath)

	rootDirectoryPathTheme := parameters.Theme.Footer.RootDirectoryPath

	rootDirectoryStyle := tcell.Style{}.
		Background(rootDirectoryPathTheme.Background).
		Foreground(rootDirectoryPathTheme.Foreground)
	rootDirectoryPathView := tview.NewTextView().
		SetText(rootDirectoryPathText).
		SetTextStyle(rootDirectoryStyle)

	rootDirectoryPathBox := tview.NewBox().
		SetBackgroundColor(rootDirectoryPathTheme.Background)
	rootDirectoryPathView.Box = rootDirectoryPathBox

	arrowRightText := fmt.Sprintf("%c", symbols.ArrowRight)
	arrowRightStyle := tcell.Style{}.
		Background(rootDirectoryPathTheme.ArrowBackground).
		Foreground(rootDirectoryPathTheme.ArrowForeground)
	arrowRightView := tview.NewTextView().
		SetText(arrowRightText).
		SetTextStyle(arrowRightStyle)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRowCSS).
		AddItem(rootDirectoryPathView, len(rootDirectoryPathText), 1, false).
		AddItem(arrowRightView, 1, 0, false)

	return layout
}
