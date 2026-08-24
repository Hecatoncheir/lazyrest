package footer

import (
	"testing"

	appcolor "github.com/Hecatoncheir/lazyrest/color"
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/ui/symbols"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
	if got := widget.environmentElement.GetText(true); got != " Env: development" {
		t.Fatalf("unexpected environment status: %q", got)
	}
}

func TestSuiteSegmentUsesMatchingPowerlineArrow(t *testing.T) {
	footerTheme := theme.NewDefault().Footer
	indicatorTheme := theme.FooterIndicatorTheme{
		Background: footerTheme.SuiteBackground,
		Foreground: footerTheme.SuiteForeground,
	}
	segment, _, width := buildSuiteElement("List users", footerTheme, indicatorTheme)
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
	if leftForeground != indicatorTheme.Background {
		t.Fatalf("suite arrow does not match suite background: arrow=%v suite=%v", leftForeground, indicatorTheme.Background)
	}
	if leftBackground != footerTheme.Background {
		t.Fatalf("suite arrow does not match footer background: arrow=%v footer=%v", leftBackground, footerTheme.Background)
	}
	if textForeground != indicatorTheme.Foreground || textBackground != indicatorTheme.Background {
		t.Fatalf("suite text does not match segment colors: foreground=%v background=%v", textForeground, textBackground)
	}
}

func TestFooterAppliesIndicatorStateToSuiteAndProgress(t *testing.T) {
	footerTheme := theme.NewDefault().Footer
	tests := []struct {
		name       string
		state      IndicatorState
		background tcell.Color
		foreground tcell.Color
	}{
		{
			name:       "default",
			state:      IndicatorDefault,
			background: footerTheme.SuiteBackground,
			foreground: footerTheme.SuiteForeground,
		},
		{
			name:       "success",
			state:      IndicatorSuccess,
			background: footerTheme.SuiteSuccess.Background,
			foreground: footerTheme.SuiteSuccess.Foreground,
		},
		{
			name:       "failure",
			state:      IndicatorFailure,
			background: footerTheme.SuiteFailure.Background,
			foreground: footerTheme.SuiteFailure.Foreground,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			widget := New()
			widget.Build(Parameters{RootDirectoryPath: "/workspace", Theme: theme.NewDefault()})
			widget.UpdateSuite("List users")
			widget.UpdateStatus("Running [====>-------]")
			widget.UpdateIndicatorState(test.state)

			assertTextViewColors(t, widget.suiteElement, test.foreground, test.background)
			assertTextViewColors(t, widget.statusElement, test.foreground, test.background)
		})
	}
}

func TestDefaultFooterIndicatorPalette(t *testing.T) {
	footerTheme := theme.NewDefault().Footer
	tests := []struct {
		name string
		got  tcell.Color
		want tcell.Color
	}{
		{name: "default background", got: footerTheme.SuiteBackground, want: appcolor.Color("#fabd2f").ToTerminal()},
		{name: "default foreground", got: footerTheme.SuiteForeground, want: appcolor.Color("#3c3836").ToTerminal()},
		{name: "success background", got: footerTheme.SuiteSuccess.Background, want: appcolor.Color("#b8bb26").ToTerminal()},
		{name: "success foreground", got: footerTheme.SuiteSuccess.Foreground, want: appcolor.Color("#3c3836").ToTerminal()},
		{name: "failure background", got: footerTheme.SuiteFailure.Background, want: appcolor.Color("#d65d0e").ToTerminal()},
		{name: "failure foreground", got: footerTheme.SuiteFailure.Foreground, want: appcolor.Color("#fbf1c7").ToTerminal()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("unexpected color: got %v, want %v", test.got, test.want)
			}
		})
	}
}

func assertTextViewColors(t *testing.T, view *tview.TextView, foreground, background tcell.Color) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("initialize simulation screen: %v", err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(40, 1)
	view.SetRect(0, 0, 40, 1)
	view.Draw(screen)
	screen.Show()

	cells, _, _ := screen.GetContents()
	gotForeground, gotBackground, _ := cells[0].Style.Decompose()
	if gotForeground != foreground || gotBackground != background {
		t.Fatalf("unexpected colors: foreground=%v background=%v, want foreground=%v background=%v", gotForeground, gotBackground, foreground, background)
	}
}
