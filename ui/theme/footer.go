package theme

import "github.com/gdamore/tcell/v2"

type FooterTheme struct {
	Background        tcell.Color
	Foreground        tcell.Color
	RootDirectoryPath RootDirectoryPathTheme
	SelectedFileName  SelectedFileNameTheme
}

type RootDirectoryPathTheme struct {
	Background      tcell.Color
	Foreground      tcell.Color
	ArrowBackground tcell.Color
	ArrowForeground tcell.Color
}

type SelectedFileNameTheme struct {
	Background                   tcell.Color
	Foreground                   tcell.Color
	RootDirectoryArrowBackground tcell.Color
	RootDirectoryArrowForeground tcell.Color
	ArrowBackground              tcell.Color
	ArrowForeground              tcell.Color
}
