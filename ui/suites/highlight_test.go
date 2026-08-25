package suites

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func renderedItem(t *testing.T, suite http.HttpSuite) (string, string) {
	t.Helper()
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnSuiteSelectCallbackType: func(http.HttpSuite) {}})
	widget.suites = []http.HttpSuite{suite}
	widget.render()

	element := widget.Element.(*tview.List)
	if element.GetItemCount() != 1 {
		t.Fatalf("expected one request, got %d", element.GetItemCount())
	}
	main, secondary := element.GetItemText(0)
	return main, secondary
}

func TestRenderHighlightsTheBodyPreview(t *testing.T) {
	_, secondary := renderedItem(t, http.HttpSuite{
		Name:     "Create user",
		Method:   "POST",
		Uri:      "/users",
		Body:     "{\n  \"name\": \"Ada\",\n  \"age\": 36\n}",
		BodyType: "json",
	})

	// The default theme is gruvbox: keys take its accent, strings its success
	// colour, numbers its progress colour.
	for _, want := range []string{`[#83a598]"name"`, `[#b8bb26]"Ada"`, `[#fabd2f]36`} {
		if !strings.Contains(secondary, want) {
			t.Errorf("missing %s in %q", want, secondary)
		}
	}
	if strings.Contains(secondary, "\n") {
		t.Errorf("the preview is not a single line: %q", secondary)
	}
}

func TestRenderKeepsBracketsOutOfStyleTags(t *testing.T) {
	// tview reads style tags in list items by default, so a body or a name
	// holding something like ["a"] would otherwise disappear from the row.
	body := `{"tags":["a"],"note":"[red] alert"}`
	main, secondary := renderedItem(t, http.HttpSuite{
		Name:     "Tags [beta]",
		Method:   "GET",
		Uri:      "/tags",
		Body:     body,
		BodyType: "json",
	})

	if !strings.Contains(main, "[beta]") {
		t.Errorf("the name lost its brackets: %q", main)
	}
	// The width tview measures ignores style tags, so it equals the number of
	// characters the row actually shows.
	if width := tview.TaggedStringWidth(secondary); width != len([]rune(body)) {
		t.Errorf("the row shows %d characters of %d: %q", width, len([]rune(body)), secondary)
	}
	// A bracket inside a string is the dangerous one: it has to be escaped.
	if !strings.Contains(secondary, "[red[]") {
		t.Errorf("a bracket inside a string was not escaped: %q", secondary)
	}
}

func TestRenderLeavesAPlainBodyUncoloured(t *testing.T) {
	_, secondary := renderedItem(t, http.HttpSuite{
		Name: "Send text",
		Body: "just some text",
	})

	if strings.Contains(secondary, "[#") {
		t.Errorf("a body without a known format was coloured: %q", secondary)
	}
	if secondary != "just some text" {
		t.Errorf("unexpected preview: %q", secondary)
	}
}
