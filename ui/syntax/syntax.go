// Package syntax turns request and response bodies into tview markup, colouring
// the tokens of the formats lazyrest can display.
package syntax

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MaxHighlightBytes caps the text that is coloured. Scanning happens on the
// draw goroutine, so a large body is rendered plain rather than stalling it.
const MaxHighlightBytes = 256 << 10

// Language selects the scanner used for a body.
type Language uint8

const (
	LanguagePlain Language = iota
	LanguageJSON
	LanguageXML
	LanguageGraphQL
)

// Palette assigns a colour to every token role. A zero colour leaves the token
// in the foreground of the pane, so a zero Palette renders plain text.
type Palette struct {
	Key         tcell.Color
	String      tcell.Color
	Number      tcell.Color
	Literal     tcell.Color
	Keyword     tcell.Color
	Variable    tcell.Color
	Punctuation tcell.Color
	Comment     tcell.Color
}

// Highlight returns render-ready markup: the text is escaped, so the caller
// must not escape it again.
func Highlight(text string, language Language, palette Palette) string {
	if language == LanguagePlain || len(text) > MaxHighlightBytes {
		return tview.Escape(text)
	}

	writer := &writer{palette: palette}
	switch language {
	case LanguageJSON:
		scanJSON(text, writer)
	case LanguageXML:
		scanXML(text, writer)
	case LanguageGraphQL:
		scanGraphQL(text, writer)
	default:
		return tview.Escape(text)
	}
	return writer.String()
}

type writer struct {
	out     strings.Builder
	palette Palette
}

// token writes text in the colour of its role, restoring the pane foreground
// afterwards.
func (w *writer) token(text string, color tcell.Color) {
	if text == "" {
		return
	}
	tag := colorTag(color)
	if tag == "" {
		w.plain(text)
		return
	}
	w.out.WriteString(tag)
	w.out.WriteString(tview.Escape(text))
	w.out.WriteString("[-]")
}

func (w *writer) plain(text string) {
	w.out.WriteString(tview.Escape(text))
}

func (w *writer) String() string {
	return w.out.String()
}

func colorTag(color tcell.Color) string {
	if !color.Valid() {
		return ""
	}
	hex := color.Hex()
	if hex < 0 {
		return ""
	}
	return fmt.Sprintf("[#%06x]", hex)
}
