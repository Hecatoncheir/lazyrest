package footer

import (
	"fmt"

	"lazyrest/ui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func buildRootDirectoryElement(
	rootDirectoryPath string,
	footerTheme theme.FooterTheme,
) (
	element tview.Primitive,
	size int,
) {
	marginLeft := 1
	rootDirectoryPathText := ""
	for range marginLeft {
		rootDirectoryPathText += " "
	}

	rootDirectoryPathText += fmt.Sprintf("%v", rootDirectoryPath)

	rootDirectoryPathTheme := footerTheme.RootDirectoryPath
	rootDirectoryStyle := tcell.Style{}.
		Background(rootDirectoryPathTheme.Background).
		Foreground(rootDirectoryPathTheme.Foreground)
	rootDirectoryPathView := tview.NewTextView().
		SetText(rootDirectoryPathText).
		SetTextStyle(rootDirectoryStyle)

	rootDirectoryPathBox := tview.NewBox().
		SetBackgroundColor(rootDirectoryPathTheme.Background)
	rootDirectoryPathView.Box = rootDirectoryPathBox

	marginRight := 1
	return rootDirectoryPathView, len(rootDirectoryPathText) + marginRight
}
