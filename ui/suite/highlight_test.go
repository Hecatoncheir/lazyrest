package suite

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

func TestRenderHighlightsTheBody(t *testing.T) {
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnRunCallback: func(http.HttpSuite) {}})

	text := widget.render(http.HttpSuite{
		Method:   "POST",
		Uri:      "https://example.com/users",
		Body:     "{\n  \"name\": \"Ada\",\n  \"age\": 36\n}",
		BodyType: "json",
	})

	if !strings.Contains(text, "POST https://example.com/users") {
		t.Errorf("the request line is missing: %q", text)
	}
	if !strings.Contains(text, "(json):\n") {
		t.Errorf("the body does not start on its own line: %q", text)
	}
	for _, want := range []string{`[#83a598]"name"`, `[#b8bb26]"Ada"`, `[#fabd2f]36`} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %s in %q", want, text)
		}
	}
}

func TestRenderEscapesMarkupInTheRequest(t *testing.T) {
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnRunCallback: func(http.HttpSuite) {}})

	text := widget.render(http.HttpSuite{
		Method: "GET",
		Uri:    "https://example.com/[beta]/users",
		Body:   `{"tags":["a"]}`,
	})

	if strings.Contains(text, "[beta]/") {
		t.Errorf("the URI was left as a colour tag: %q", text)
	}
	if !strings.Contains(text, "[beta[]") {
		t.Errorf("the URI was not escaped: %q", text)
	}
}

func TestRenderWithoutBody(t *testing.T) {
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnRunCallback: func(http.HttpSuite) {}})

	text := widget.render(http.HttpSuite{Method: "GET", Uri: "https://example.com/users"})
	if strings.Contains(text, widget.locale.Text("body")) {
		t.Errorf("an empty body was labelled: %q", text)
	}
}
