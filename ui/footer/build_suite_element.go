package footer

import (
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

const maxSuiteSegmentWidth = 42

func buildSuiteElement(
	suiteName string,
	footerTheme theme.FooterTheme,
	indicatorTheme theme.FooterIndicatorTheme,
) (*tview.Flex, *tview.TextView, int) {
	leftArrow, leftArrowSize := buildArrowLeftElement(footerTheme.Background, indicatorTheme.Background)

	text := " Suite: " + suiteName + " "
	textView := tview.NewTextView().
		SetText(text).
		SetTextColor(indicatorTheme.Foreground)
	textView.SetBackgroundColor(indicatorTheme.Background)

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
