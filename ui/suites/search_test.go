package suites

import (
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
