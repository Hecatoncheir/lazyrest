package theme

import (
	"lazyrest/color"

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
		Border:     color.Color("#a89984").ToTerminal(),
		Tree: TreeTheme{
			Title:           color.Color("#bdae93").ToTerminal(),
			TitleFocus:      color.Color("#83a598").ToTerminal(),
			Background:      color.Color("#504945").ToTerminal(),
			BackgroundFocus: color.Color("#3c3836").ToTerminal(),
			Border:          color.Color("#bdae93").ToTerminal(),
			BorderFocus:     color.Color("#fbf1c7").ToTerminal(),
			NodeDirectory: TreeNodeTheme{
				Foreground: color.Color("#fabd2f").ToTerminal(),
			},
			Node: TreeNodeTheme{
				Foreground: color.Color("#fbf1c7").ToTerminal(),
			},
		},
		Suites: SuitesTheme{
			Title:                color.Color("#bdae93").ToTerminal(),
			TitleFocus:           color.Color("#83a598").ToTerminal(),
			Background:           color.Color("#504945").ToTerminal(),
			BackgroundFocus:      color.Color("#3c3836").ToTerminal(),
			Border:               color.Color("#bdae93").ToTerminal(),
			BorderFocus:          color.Color("#fbf1c7").ToTerminal(),
			SuiteBackground:      color.Color("#504945").ToTerminal(),
			SuiteFocusBackground: color.Color("#fbf1c7").ToTerminal(),
			SuiteForeground:      color.Color("#bdae93").ToTerminal(),
			SuiteFocusForeground: color.Color("#282828").ToTerminal(),
		},
		Suite: SuiteTheme{
			Title:           color.Color("#bdae93").ToTerminal(),
			TitleFocus:      color.Color("#83a598").ToTerminal(),
			Foreground:      color.Color("#fbf1c7").ToTerminal(),
			Background:      color.Color("#504945").ToTerminal(),
			BackgroundFocus: color.Color("#3c3836").ToTerminal(),
			Border:          color.Color("#bdae93").ToTerminal(),
			BorderFocus:     color.Color("#fbf1c7").ToTerminal(),
		},
		Producer: ProducerTheme{
			Title:           color.Color("#bdae93").ToTerminal(),
			TitleFocus:      color.Color("#83a598").ToTerminal(),
			Foreground:      color.Color("#fbf1c7").ToTerminal(),
			Background:      color.Color("#504945").ToTerminal(),
			BackgroundFocus: color.Color("#3c3836").ToTerminal(),
			Border:          color.Color("#bdae93").ToTerminal(),
			BorderFocus:     color.Color("#fbf1c7").ToTerminal(),
		},
		Footer: FooterTheme{
			Background: color.Color("#504945").ToTerminal(),
			Foreground: color.Color("#bdae93").ToTerminal(),
			RootDirectoryPath: RootDirectoryPathTheme{
				Background:      color.Color("#bdae93").ToTerminal(),
				Foreground:      color.Color("#504945").ToTerminal(),
				ArrowBackground: color.Color("#504945").ToTerminal(),
				ArrowForeground: color.Color("#bdae93").ToTerminal(),
			},
		},
	}
	return theme
}
