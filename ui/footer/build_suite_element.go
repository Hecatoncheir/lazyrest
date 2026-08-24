package footer

import (
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

const maxSuiteSegmentWidth = 42

func buildSuiteElement(suiteName string, footerTheme theme.FooterTheme) (*tview.Flex, *tview.TextView, int) {
	leftArrow, leftArrowSize := buildArrowLeftElement(footerTheme.Background, footerTheme.SuiteBackground)

	text := " Suite: " + suiteName + " "
	textView := tview.NewTextView().
		SetText(text).
		SetTextColor(footerTheme.SuiteForeground)
	textView.SetBackgroundColor(footerTheme.SuiteBackground)

	segmentWidth := tview.TaggedStringWidth(text) + leftArrowSize
	if segmentWidth > maxSuiteSegmentWidth {
		segmentWidth = maxSuiteSegmentWidth
	}
	segment := tview.NewFlex().
		SetDirection(tview.FlexRowCSS).
		AddItem(leftArrow, leftArrowSize, 0, false).
		AddItem(textView, 0, 1, false)
	return segment, textView, segmentWidth
}
