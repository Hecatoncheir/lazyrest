package theme

import "github.com/gdamore/tcell/v2"

type ProducerTheme struct {
	Title           tcell.Color
	TitleFocus      tcell.Color
	Background      tcell.Color
	BackgroundFocus tcell.Color
	Foreground      tcell.Color
	Border          tcell.Color
	BorderFocus     tcell.Color
	Syntax          SyntaxTheme
}

// SyntaxTheme colours the tokens of a highlighted body. The colours come from
// the semantic palette of the theme, so every preset gets a matching one.
type SyntaxTheme struct {
	Key         tcell.Color
	String      tcell.Color
	Number      tcell.Color
	Literal     tcell.Color
	Keyword     tcell.Color
	Variable    tcell.Color
	Punctuation tcell.Color
	Comment     tcell.Color
}
