package tree

import (
	"fmt"
	"lazyrest/ui/theme"

	"github.com/rivo/tview"
)

func buildNoFilesFound(rootDirectotyPath string, theme theme.TreeTheme) *tview.TextView {
	text := fmt.Sprintf("No files found in directory: %v", rootDirectotyPath)
	widget := tview.NewTextView().
		SetText(text)
	return widget
}
