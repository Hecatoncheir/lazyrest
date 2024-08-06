package theme

import "github.com/gdamore/tcell/v2"

type Theme struct {
	Background tcell.Color
	Border     tcell.Color
	Tree       TreeTheme
	Suites     SuitesTheme
	Suite      SuiteTheme
	Producer   ProducerTheme
	Footer     FooterTheme
}
