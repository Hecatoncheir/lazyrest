package theme

import (
	"github.com/Hecatoncheir/lazyrest/ui/syntax"
	"github.com/gdamore/tcell/v2"
)

type Theme struct {
	Background tcell.Color
	Border     tcell.Color
	Tree       TreeTheme
	Suites     SuitesTheme
	Suite      SuiteTheme
	Producer   ProducerTheme
	Footer     FooterTheme
	// Syntax colours highlighted bodies. It is shared by every panel that
	// shows one, and is derived from the semantic colours of the theme.
	Syntax syntax.Palette
	// Methods colours an HTTP method by what it does.
	Methods syntax.MethodPalette
}
