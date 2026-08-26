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

	// The row reads "GET Tags [beta]": the method, then the name it was given.
	if width := tview.TaggedStringWidth(main); width != len("GET Tags [beta]") {
		t.Errorf("the row shows %d characters of %d: %q", width, len("GET Tags [beta]"), main)
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

// unselectedRow renders a suite in a position that does not carry the
// selection, where the row keeps its markup.
func unselectedRow(t *testing.T, suite http.HttpSuite) string {
	t.Helper()
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnSuiteSelectCallbackType: func(http.HttpSuite) {}})
	widget.suites = []http.HttpSuite{{Name: "First", Method: "GET", Uri: "/first"}, suite}
	widget.render()

	main, _ := widget.Element.(*tview.List).GetItemText(1)
	return main
}

func TestRenderColoursTheMethod(t *testing.T) {
	// The default theme is gruvbox: a read takes its success colour, a create
	// its progress colour, and a delete its failure colour.
	cases := []struct {
		method string
		want   string
	}{
		{method: "GET", want: "[#b8bb26]GET[-]"},
		{method: "POST", want: "[#fabd2f]POST[-]"},
		{method: "PATCH", want: "[#83a598]PATCH[-]"},
		{method: "DELETE", want: "[#d65d0e]DELETE[-]"},
		{method: "HURL", want: "[#bdae93]HURL[-]"},
	}

	for _, testCase := range cases {
		t.Run(testCase.method, func(t *testing.T) {
			main := unselectedRow(t, http.HttpSuite{Name: "Do it", Method: testCase.method, Uri: "/x"})
			if !strings.HasPrefix(main, testCase.want) {
				t.Errorf("got %q, want it to start with %q", main, testCase.want)
			}
			if !strings.HasSuffix(main, " Do it") {
				t.Errorf("the name is missing: %q", main)
			}
		})
	}
}

func TestRenderDoesNotRepeatTheMethodOfAnUnnamedRequest(t *testing.T) {
	// The parser names an unnamed request "METHOD uri", which already leads
	// with the method.
	main := unselectedRow(t, http.HttpSuite{Name: "GET https://example.com/a", Method: "GET", Uri: "https://example.com/a"})

	if tview.TaggedStringWidth(main) != len("GET https://example.com/a") {
		t.Errorf("the method was repeated: %q", main)
	}
	if !strings.HasPrefix(main, "[#b8bb26]GET[-]") {
		t.Errorf("the method was not coloured: %q", main)
	}
}

func TestRenderDropsMarkupFromTheSelectedRow(t *testing.T) {
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnSuiteSelectCallbackType: func(http.HttpSuite) {}})
	widget.suites = []http.HttpSuite{
		{Name: "First", Method: "GET", Uri: "/a"},
		{Name: "Second", Method: "POST", Uri: "/b"},
	}
	widget.render()

	element := widget.Element.(*tview.List)
	selected, _ := element.GetItemText(0)
	if strings.Contains(selected, "[#") {
		t.Errorf("the selected row kept its colour: %q", selected)
	}
	if selected != "GET First" {
		t.Errorf("unexpected selected row: %q", selected)
	}

	other, _ := element.GetItemText(1)
	if !strings.HasPrefix(other, "[#fabd2f]POST[-]") {
		t.Errorf("an unselected row lost its colour: %q", other)
	}

	element.SetCurrentItem(1)
	first, _ := element.GetItemText(0)
	second, _ := element.GetItemText(1)
	if !strings.HasPrefix(first, "[#b8bb26]GET[-]") {
		t.Errorf("the row that lost the selection did not get its colour back: %q", first)
	}
	if strings.Contains(second, "[#") {
		t.Errorf("the row that gained the selection kept its colour: %q", second)
	}
}
