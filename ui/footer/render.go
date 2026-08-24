package footer

import "github.com/rivo/tview"

func (widget *Footer) render() {
	if widget.Element == nil {
		return
	}
	layout := widget.Element.(*tview.Flex)
	layout.Clear()

	footerTheme := widget.Parameters.Theme.Footer
	rootElement, rootSize := buildRootDirectoryElement(widget.Parameters.RootDirectoryPath, footerTheme)
	if textView, ok := rootElement.(*tview.TextView); ok {
		widget.rootDirectoryElement = textView
	}
	layout.AddItem(rootElement, rootSize, 0, false)

	arrowTheme := footerTheme.RootDirectoryPath
	arrow, arrowSize := buildArrowRightElement(arrowTheme.ArrowBackground, arrowTheme.ArrowForeground)
	layout.AddItem(arrow, arrowSize, 0, false)

	if widget.selectedFile != nil {
		selectedTheme := footerTheme.SelectedFileName
		fileElement, fileSize := buildSelectedFileElement(*widget.selectedFile, selectedTheme)
		layout.AddItem(fileElement, fileSize, 0, false)

		fileArrow, fileArrowSize := buildArrowRightElement(selectedTheme.ArrowBackground, selectedTheme.ArrowForeground)
		layout.AddItem(fileArrow, fileArrowSize, 0, false)
	}

	suiteText := tview.NewTextView()
	suiteText.SetTextAlign(tview.AlignLeft)
	suiteText.SetTextColor(footerTheme.SuiteForeground)
	suiteText.SetBackgroundColor(footerTheme.Background)
	if widget.suiteName != "" {
		suiteText.SetText(" Suite: " + widget.suiteName)
	}
	widget.suiteElement = suiteText
	layout.AddItem(suiteText, 0, 1, false)
}
