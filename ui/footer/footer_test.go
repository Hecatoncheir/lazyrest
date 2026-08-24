package footer

import (
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"testing"
)

func TestFooterKeepsSuiteAfterFileSelection(t *testing.T) {
	widget := New()
	widget.Build(Parameters{RootDirectoryPath: "/workspace", Theme: theme.NewDefault()})
	widget.SelectFile(finder.File{Name: "requests.http", Path: "/workspace/requests.http"})
	widget.UpdateSuite("List users")

	if got := widget.suiteElement.GetText(true); got != " Suite: List users" {
		t.Fatalf("unexpected suite breadcrumb: %q", got)
	}
}

func TestFooterShowsSelectedEnvironment(t *testing.T) {
	widget := New()
	widget.Build(Parameters{
		RootDirectoryPath: "/workspace",
		EnvironmentName:   "development",
		Theme:             theme.NewDefault(),
	})
	widget.UpdateSuite("List users")

	if got := widget.suiteElement.GetText(true); got != " Env: development Suite: List users" {
		t.Fatalf("unexpected environment breadcrumb: %q", got)
	}
}
