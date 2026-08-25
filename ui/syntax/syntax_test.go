package syntax

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func testPalette() Palette {
	return Palette{
		Key:         tcell.NewRGBColor(0x11, 0x11, 0x11),
		String:      tcell.NewRGBColor(0x22, 0x22, 0x22),
		Number:      tcell.NewRGBColor(0x33, 0x33, 0x33),
		Literal:     tcell.NewRGBColor(0x44, 0x44, 0x44),
		Keyword:     tcell.NewRGBColor(0x55, 0x55, 0x55),
		Variable:    tcell.NewRGBColor(0x66, 0x66, 0x66),
		Punctuation: tcell.NewRGBColor(0x77, 0x77, 0x77),
		Comment:     tcell.NewRGBColor(0x88, 0x88, 0x88),
	}
}

func TestHighlightJSON(t *testing.T) {
	text := "{\n  \"name\": \"Ada\",\n  \"age\": 36,\n  \"active\": true,\n  \"tag\": null\n}"
	got := Highlight(text, LanguageJSON, testPalette())

	for _, want := range []string{
		`[#111111]"name"[-]`,
		`[#222222]"Ada"[-]`,
		`[#333333]36[-]`,
		`[#444444]true[-]`,
		`[#444444]null[-]`,
		`[#777777]{[-]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %q", want, got)
		}
	}
}

func TestHighlightJSONKeepsEveryCharacter(t *testing.T) {
	text := "{\n  \"greeting\": \"привет — ok\",\n  \"escaped\": \"a\\\"b\"\n}"
	got := Highlight(text, LanguageJSON, testPalette())

	if stripped := stripTags(got); stripped != text {
		t.Fatalf("the text changed:\n got %q\nwant %q", stripped, text)
	}
}

func TestHighlightEscapesMarkupInContent(t *testing.T) {
	text := `{"note": "[red]not a tag[-]"}`
	got := Highlight(text, LanguageJSON, testPalette())

	if strings.Contains(got, `"[red]`) {
		t.Fatalf("content was left as a colour tag: %q", got)
	}
	if !strings.Contains(got, "[red[]") {
		t.Fatalf("content was not escaped: %q", got)
	}
}

func TestHighlightWithoutPaletteIsPlain(t *testing.T) {
	text := `{"a": 1}`
	if got := Highlight(text, LanguageJSON, Palette{}); got != tview.Escape(text) {
		t.Fatalf("a zero palette added markup: %q", got)
	}
}

func TestHighlightSkipsLargeBodies(t *testing.T) {
	text := "{\"a\": \"" + strings.Repeat("x", MaxHighlightBytes) + "\"}"
	if got := Highlight(text, LanguageJSON, testPalette()); got != tview.Escape(text) {
		t.Fatal("a body over the limit was highlighted")
	}
}

func TestHighlightPlainLanguage(t *testing.T) {
	text := "just words"
	if got := Highlight(text, LanguagePlain, testPalette()); got != tview.Escape(text) {
		t.Fatalf("plain text was highlighted: %q", got)
	}
}

func TestHighlightXML(t *testing.T) {
	text := `<!-- note --><message id="7"><text>Hello</text></message>`
	got := Highlight(text, LanguageXML, testPalette())

	for _, want := range []string{
		`[#888888]<!-- note -->[-]`,
		`[#111111]message[-]`,
		`[#666666]id[-]`,
		`[#222222]"7"[-]`,
		`[#777777]<[-]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %q", want, got)
		}
	}
	if stripped := stripTags(got); stripped != text {
		t.Fatalf("the text changed:\n got %q\nwant %q", stripped, text)
	}
}

func TestHighlightGraphQL(t *testing.T) {
	text := "# a comment\nquery GetUser($id: ID!) {\n  user(id: $id) { name active }\n}"
	got := Highlight(text, LanguageGraphQL, testPalette())

	for _, want := range []string{
		`[#888888]# a comment[-]`,
		`[#555555]query[-]`,
		`[#666666]$id[-]`,
		`[#777777]{[-]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %q", want, got)
		}
	}
	if strings.Contains(got, `[#555555]GetUser[-]`) {
		t.Errorf("an operation name was coloured as a keyword: %q", got)
	}
	if stripped := stripTags(got); stripped != text {
		t.Fatalf("the text changed:\n got %q\nwant %q", stripped, text)
	}
}

// stripTags removes the markup that Highlight added, undoing the escaping as
// well, so that the original text can be compared against.
func stripTags(text string) string {
	var out strings.Builder
	for index := 0; index < len(text); index++ {
		if text[index] != '[' {
			out.WriteByte(text[index])
			continue
		}
		end := strings.IndexByte(text[index:], ']')
		if end < 0 {
			out.WriteByte(text[index])
			continue
		}
		tag := text[index : index+end+1]
		if tag == "[]" {
			// tview escaping: "[foo[]" stands for the literal "[foo]"
			out.WriteString("]")
		}
		index += end
	}
	return out.String()
}
