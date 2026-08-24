package footer

import "github.com/rivo/tview"

func (widget *Footer) render() {
	if widget.Element == nil {
		return
	}
	layout := widget.Element.(*tview.Flex)
	layout.Clear()
	breadcrumbs := tview.NewFlex().SetDirection(tview.FlexRowCSS)

	footerTheme := widget.Parameters.Theme.Footer
	indicatorTheme := widget.indicatorTheme()
	rootElement, rootSize := buildRootDirectoryElement(widget.Parameters.RootDirectoryPath, footerTheme)
	if textView, ok := rootElement.(*tview.TextView); ok {
		widget.rootDirectoryElement = textView
	}
	breadcrumbs.AddItem(rootElement, rootSize, 0, false)

	arrowTheme := footerTheme.RootDirectoryPath
	arrow, arrowSize := buildArrowRightElement(arrowTheme.ArrowBackground, arrowTheme.ArrowForeground)
	breadcrumbs.AddItem(arrow, arrowSize, 0, false)

	if widget.selectedFile != nil {
		selectedTheme := footerTheme.SelectedFileName

		rightArrow, rightArrowSize := buildArrowRightElement(selectedTheme.Background, selectedTheme.Foreground)
		breadcrumbs.AddItem(rightArrow, rightArrowSize+1, 0, false)

		fileElement, fileSize := buildSelectedFileElement(*widget.selectedFile, selectedTheme)
		breadcrumbs.AddItem(fileElement, fileSize, 0, false)

		fileArrow, fileArrowSize := buildArrowRightElement(selectedTheme.ArrowBackground, selectedTheme.ArrowForeground)
		breadcrumbs.AddItem(fileArrow, fileArrowSize, 0, false)
	}

	environmentText := tview.NewTextView()
	environmentText.SetTextAlign(tview.AlignLeft)
	environmentText.SetTextColor(footerTheme.Foreground)
	environmentText.SetBackgroundColor(footerTheme.Background)
	environment := ""
	if widget.Parameters.EnvironmentName != "" {
		environment = " Env: " + widget.Parameters.EnvironmentName
	}
	environmentText.SetText(environment)
	widget.environmentElement = environmentText

	statusText := tview.NewTextView()
	statusText.SetTextAlign(tview.AlignLeft)
	statusText.SetTextColor(indicatorTheme.Foreground)
	statusText.SetBackgroundColor(indicatorTheme.Background)
	status := ""
	if widget.status != "" {
		status = " " + widget.status + " "
	}
	statusText.SetText(status)
	widget.statusElement = statusText

	layout.AddItem(breadcrumbs, 0, 1, false)
	if widget.suiteName != "" {
		suiteSegment, suiteText, suiteWidth := buildSuiteElement(widget.suiteName, footerTheme, indicatorTheme)
		widget.suiteElement = suiteText
		layout.AddItem(suiteSegment, suiteWidth, 0, false)
	} else {
		widget.suiteElement = nil
	}
	if environment != "" {
		layout.AddItem(environmentText, tview.TaggedStringWidth(environment), 0, false)
	}
	if status != "" {
		layout.AddItem(statusText, tview.TaggedStringWidth(status), 0, false)
	}
}
