package footer

import (
	"testing"

	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/symbols"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/gdamore/tcell/v2"
)

func TestFooterKeepsSuiteAfterFileSelection(t *testing.T) {
	widget := New()
	widget.Build(Parameters{RootDirectoryPath: "/workspace", Theme: theme.NewDefault()})
	widget.SelectFile(finder.File{Name: "requests.http", Path: "/workspace/requests.http"})
	widget.UpdateSuite("List users")

	if got := widget.suiteElement.GetText(true); got != " Suite: List users " {
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

	if got := widget.suiteElement.GetText(true); got != " Suite: List users " {
		t.Fatalf("unexpected suite breadcrumb: %q", got)
	}
	if got := widget.statusElement.GetText(true); got != " Env: development" {
		t.Fatalf("unexpected environment status: %q", got)
	}
}

func TestSuiteSegmentUsesMatchingPowerlineArrow(t *testing.T) {
	footerTheme := theme.NewDefault().Footer
	segment, _, width := buildSuiteElement("List users", footerTheme)
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("initialize simulation screen: %v", err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(width, 1)
	segment.SetRect(0, 0, width, 1)
	segment.Draw(screen)
	screen.Show()

	cells, renderedWidth, renderedHeight := screen.GetContents()
	if renderedWidth != width || renderedHeight != 1 {
		t.Fatalf("unexpected rendered size: got %dx%d, want %dx1", renderedWidth, renderedHeight, width)
	}
	if got := cells[0].Runes[0]; got != symbols.ArrowLeft {
		t.Fatalf("unexpected leading arrow: got %q, want %q", got, symbols.ArrowLeft)
	}
	leftForeground, leftBackground, _ := cells[0].Style.Decompose()
	textForeground, textBackground, _ := cells[width-1].Style.Decompose()
	if leftForeground != footerTheme.SuiteBackground {
		t.Fatalf("suite arrow does not match suite background: arrow=%v suite=%v", leftForeground, footerTheme.SuiteBackground)
	}
	if leftBackground != footerTheme.Background {
		t.Fatalf("suite arrow does not match footer background: arrow=%v footer=%v", leftBackground, footerTheme.Background)
	}
	if textForeground != footerTheme.SuiteForeground || textBackground != footerTheme.SuiteBackground {
		t.Fatalf("suite text does not match segment colors: foreground=%v background=%v", textForeground, textBackground)
	}
}
