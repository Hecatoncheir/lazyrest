package theme

import "github.com/gdamore/tcell/v2"

type TreeTheme struct {
	Title           tcell.Color
	TitleFocus      tcell.Color
	Background      tcell.Color
	BackgroundFocus tcell.Color
	Border          tcell.Color
	BorderFocus     tcell.Color
	Node            TreeNodeTheme
	NodeDirectory   TreeNodeTheme
}

type TreeNodeTheme struct {
	Foreground tcell.Color
}
