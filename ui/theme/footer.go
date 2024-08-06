package theme

import "github.com/gdamore/tcell/v2"

type FooterTheme struct {
	Background        tcell.Color
	Foreground        tcell.Color
	RootDirectoryPath struct {
		Background      tcell.Color
		Foreground      tcell.Color
		ArrowBackground tcell.Color
		ArrowForeground tcell.Color
	}
}

type RootDirectoryPathTheme struct {
	Background      tcell.Color
	Foreground      tcell.Color
	ArrowBackground tcell.Color
	ArrowForeground tcell.Color
}
