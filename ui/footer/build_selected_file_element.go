package footer

import (
	"fmt"

	"lazyrest/finder"
	"lazyrest/ui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func buildSelectedFileElement(
	file finder.File,
	selectedFileNameTheme theme.RootDirectoryPathTheme,
) (
	element tview.Primitive,
	size int,
) {
	selectedFileName := file.Name
	selectedFileNameText := fmt.Sprintf("%v", selectedFileName)

	selectedFileStyle := tcell.Style{}.
		Background(selectedFileNameTheme.Background).
		Foreground(selectedFileNameTheme.Foreground)
	selectedFileNameView := tview.NewTextView().
		SetText(selectedFileNameText).
		SetTextStyle(selectedFileStyle)

	rootDirectoryNameBox := tview.NewBox().
		SetBackgroundColor(selectedFileNameTheme.Background)
	selectedFileNameView.Box = rootDirectoryNameBox

	return selectedFileNameView, len(selectedFileNameText)
}
