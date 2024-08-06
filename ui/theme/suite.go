package theme

import "github.com/gdamore/tcell/v2"

type SuiteTheme struct {
	Title           tcell.Color
	TitleFocus      tcell.Color
	Background      tcell.Color
	BackgroundFocus tcell.Color
	Foreground      tcell.Color
	Border          tcell.Color
	BorderFocus     tcell.Color
}
