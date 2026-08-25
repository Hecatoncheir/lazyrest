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

// role names the kind of a token, which the writer turns into markup.
type role uint8

const (
	rolePlain role = iota
	roleKey
	roleString
	roleNumber
	roleLiteral
	roleKeyword
	roleVariable
	rolePunctuation
	roleComment
	roleCount
)

// Highlight returns render-ready markup: the text is escaped, so the caller
// must not escape it again.
func Highlight(text string, language Language, palette Palette) string {
	if language == LanguagePlain || len(text) > MaxHighlightBytes {
		return escape(text)
	}

	writer := newWriter(palette, len(text))
	switch language {
	case LanguageJSON:
		scanJSON(text, writer)
	case LanguageXML:
		scanXML(text, writer)
	case LanguageGraphQL:
		scanGraphQL(text, writer)
	default:
		return escape(text)
	}
	return writer.String()
}

type writer struct {
	out strings.Builder
	// tags holds the markup of every role, built once so that highlighting
	// never formats a colour per token.
	tags [roleCount]string
}

func newWriter(palette Palette, size int) *writer {
	w := &writer{}
	w.tags[roleKey] = colorTag(palette.Key)
	w.tags[roleString] = colorTag(palette.String)
	w.tags[roleNumber] = colorTag(palette.Number)
	w.tags[roleLiteral] = colorTag(palette.Literal)
	w.tags[roleKeyword] = colorTag(palette.Keyword)
	w.tags[roleVariable] = colorTag(palette.Variable)
	w.tags[rolePunctuation] = colorTag(palette.Punctuation)
	w.tags[roleComment] = colorTag(palette.Comment)
	w.out.Grow(size + size/2)
	return w
}

// token writes text in the colour of its role, restoring the pane foreground
// afterwards.
func (w *writer) token(text string, kind role) {
	if text == "" {
		return
	}
	tag := w.tags[kind]
	if tag == "" {
		w.plain(text)
		return
	}
	w.out.WriteString(tag)
	w.write(text)
	w.out.WriteString("[-]")
}

func (w *writer) plain(text string) {
	w.write(text)
}

func (w *writer) write(text string) {
	// tview.Escape runs a regular expression, which dominated highlighting
	// even though almost no token contains a bracket at all.
	if strings.IndexByte(text, '[') < 0 {
		w.out.WriteString(text)
		return
	}
	w.out.WriteString(tview.Escape(text))
}

func (w *writer) String() string {
	return w.out.String()
}

func escape(text string) string {
	if strings.IndexByte(text, '[') < 0 {
		return text
	}
	return tview.Escape(text)
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
