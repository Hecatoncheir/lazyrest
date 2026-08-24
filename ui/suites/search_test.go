package suites

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func TestRenderFiltersSuites(t *testing.T) {
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnSuiteSelectCallbackType: func(http.HttpSuite) {}})
	widget.suites = []http.HttpSuite{
		{Name: "List users", Method: "GET", Uri: "/users"},
		{Name: "Create project", Method: "POST", Uri: "/projects"},
	}
	widget.searchQuery = "users"
	widget.render()

	if count := widget.Element.(*tview.List).GetItemCount(); count != 1 {
		t.Fatalf("expected one filtered request, got %d", count)
	}
}

func TestRenderRedactsEnvironmentSecrets(t *testing.T) {
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), OnEscapeCallback: func() {}, OnSuiteSelectCallbackType: func(http.HttpSuite) {}})
	widget.suites = []http.HttpSuite{{
		Name:         "GET https://example.com/private-token",
		Body:         `{"token":"private-token"}`,
		SecretValues: []string{"private-token"},
	}}
	widget.render()

	mainText, secondaryText := widget.Element.(*tview.List).GetItemText(0)
	if strings.Contains(mainText+secondaryText, "private-token") {
		t.Fatalf("secret was rendered in Suites: %q %q", mainText, secondaryText)
	}
}
