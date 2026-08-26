package layout

import (
	"testing"

	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

func TestBuildPaintsTheLayoutWithTheTheme(t *testing.T) {
	uiTheme := theme.NewDefault()
	footerWidget := footer.New()
	footerWidget.Build(footer.Parameters{Theme: uiTheme})

	widget := New()
	element := widget.Build(Parameters{Theme: uiTheme}, buildWorkspace(t, uiTheme), footerWidget)

	if element == nil || widget.Element == nil {
		t.Fatal("the layout was not built")
	}
	if got := widget.Element.GetBackgroundColor(); got != uiTheme.Background {
		t.Fatalf("got background %v, want %v", got, uiTheme.Background)
	}
}

func TestApplySettingsRepaintsTheLayout(t *testing.T) {
	uiTheme := theme.NewDefault()
	footerWidget := footer.New()
	footerWidget.Build(footer.Parameters{Theme: uiTheme})
	widget := New()
	widget.Build(Parameters{Theme: uiTheme}, buildWorkspace(t, uiTheme), footerWidget)

	other, err := theme.FromConfig(theme.Config{Preset: "nord"})
	if err != nil {
		t.Fatal(err)
	}
	widget.ApplySettings(other)

	if got := widget.Element.GetBackgroundColor(); got != other.Background {
		t.Fatalf("the layout kept the old background: %v", got)
	}
}
