package theme

import (
	"github.com/Hecatoncheir/lazyrest/color"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"

	"github.com/rivo/tview"
)

func NewDefault() Theme {
	tview.Borders.Horizontal = tview.BoxDrawingsLightDoubleDashHorizontal
	tview.Borders.Vertical = tview.BoxDrawingsLightQuadrupleDashVertical
	tview.Borders.TopLeft = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRight = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeft = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRight = tview.BoxDrawingsLightUpAndLeft

	tview.Borders.HorizontalFocus = tview.BoxDrawingsHeavyDoubleDashHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsHeavyQuadrupleDashVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsHeavyDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsHeavyDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsHeavyUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsHeavyUpAndLeft

	theme := Theme{
		Background: color.Color("#1d2021").ToTerminal(),
		Border:     color.Color("#bdae93").ToTerminal(),
		Tree: TreeTheme{
			Title:           color.Color("#d5c4a1").ToTerminal(),
			TitleFocus:      color.Color("#8ec0c8").ToTerminal(),
			Background:      color.Color("#504945").ToTerminal(),
			BackgroundFocus: color.Color("#3c3836").ToTerminal(),
			Border:          color.Color("#bdae93").ToTerminal(),
			BorderFocus:     color.Color("#fbf1c7").ToTerminal(),
			NodeDirectory: TreeNodeTheme{
				Foreground: color.Color("#8ec0c8").ToTerminal(),
			},
			Node: TreeNodeTheme{
				Foreground: color.Color("#fbf1c7").ToTerminal(),
			},
		},
		Suites: SuitesTheme{
			Title:                color.Color("#d5c4a1").ToTerminal(),
			TitleFocus:           color.Color("#8ec0c8").ToTerminal(),
			Background:           color.Color("#504945").ToTerminal(),
			BackgroundFocus:      color.Color("#3c3836").ToTerminal(),
			Border:               color.Color("#bdae93").ToTerminal(),
			BorderFocus:          color.Color("#fbf1c7").ToTerminal(),
			SuiteBackground:      color.Color("#504945").ToTerminal(),
			SuiteFocusBackground: color.Color("#fbf1c7").ToTerminal(),
			SuiteForeground:      color.Color("#d5c4a1").ToTerminal(),
			SuiteFocusForeground: color.Color("#282828").ToTerminal(),
		},
		Suite: SuiteTheme{
			Title:           color.Color("#d5c4a1").ToTerminal(),
			TitleFocus:      color.Color("#8ec0c8").ToTerminal(),
			Foreground:      color.Color("#fbf1c7").ToTerminal(),
			Background:      color.Color("#504945").ToTerminal(),
			BackgroundFocus: color.Color("#3c3836").ToTerminal(),
			Border:          color.Color("#bdae93").ToTerminal(),
			BorderFocus:     color.Color("#fbf1c7").ToTerminal(),
		},
		Producer: ProducerTheme{
			Title:           color.Color("#d5c4a1").ToTerminal(),
			TitleFocus:      color.Color("#8ec0c8").ToTerminal(),
			Foreground:      color.Color("#fbf1c7").ToTerminal(),
			Background:      color.Color("#504945").ToTerminal(),
			BackgroundFocus: color.Color("#3c3836").ToTerminal(),
			Border:          color.Color("#bdae93").ToTerminal(),
			BorderFocus:     color.Color("#fbf1c7").ToTerminal(),
		},
		Footer: FooterTheme{
			Background:      color.Color("#504945").ToTerminal(),
			Foreground:      color.Color("#d5c4a1").ToTerminal(),
			SuiteBackground: color.Color("#fabd2f").ToTerminal(),
			SuiteForeground: color.Color("#3c3836").ToTerminal(),
			SuiteSuccess: FooterIndicatorTheme{
				Background: color.Color("#b8bb26").ToTerminal(),
				Foreground: color.Color("#3c3836").ToTerminal(),
			},
			SuiteFailure: FooterIndicatorTheme{
				Background: color.Color("#fe8019").ToTerminal(),
				Foreground: color.Color("#1d2021").ToTerminal(),
			},
			RootDirectoryPath: RootDirectoryPathTheme{
				Background:      color.Color("#bdae93").ToTerminal(),
				Foreground:      color.Color("#3c3836").ToTerminal(),
				ArrowBackground: color.Color("#504945").ToTerminal(),
				ArrowForeground: color.Color("#bdae93").ToTerminal(),
			},
			SelectedFileName: SelectedFileNameTheme{
				RootDirectoryArrowBackground: color.Color("#fbf1c7").ToTerminal(),
				RootDirectoryArrowForeground: color.Color("#bdae93").ToTerminal(),
				Background:                   color.Color("#fbf1c7").ToTerminal(),
				Foreground:                   color.Color("#282828").ToTerminal(),
				ArrowBackground:              color.Color("#504945").ToTerminal(),
				ArrowForeground:              color.Color("#fbf1c7").ToTerminal(),
			},
		},
	}
	theme.Syntax = syntax.Palette{
		Key:         color.Color("#8ec0c8").ToTerminal(),
		String:      color.Color("#b8bb26").ToTerminal(),
		Number:      color.Color("#fabd2f").ToTerminal(),
		Literal:     color.Color("#fe8019").ToTerminal(),
		Keyword:     color.Color("#fe8019").ToTerminal(),
		Variable:    color.Color("#fabd2f").ToTerminal(),
		Punctuation: color.Color("#d5c4a1").ToTerminal(),
		Comment:     color.Color("#d5c4a1").ToTerminal(),
	}
	theme.Methods = syntax.MethodPalette{
		Read:   color.Color("#b8bb26").ToTerminal(),
		Create: color.Color("#fabd2f").ToTerminal(),
		Update: color.Color("#8ec0c8").ToTerminal(),
		Delete: color.Color("#fe8019").ToTerminal(),
		Other:  color.Color("#d5c4a1").ToTerminal(),
	}
	return theme
}
