package ui

import (
	"fmt"
	"strings"

	"github.com/Hecatoncheir/lazyrest/ui/footer"
	uiprogress "github.com/Hecatoncheir/lazyrest/ui/progress"
	"github.com/rivo/tview"
)

const (
	diagnosticsPage = "diagnostics"
	helpPage        = "help"
)

func (application *Application) buildOverlays() {
	application.Diagnostics = application.newOverlayView("Diagnostics — d/Esc close")
	application.Help = application.newOverlayView("Help — ?/Esc close")
	application.Help.SetText(helpText)

	application.Pages.
		AddPage(diagnosticsPage, centered(application.Diagnostics, 84, 24), true, false).
		AddPage(helpPage, centered(application.Help, 72, 25), true, false)
	application.refreshDiagnostics()
}

func (application *Application) newOverlayView(title string) *tview.TextView {
	view := tview.NewTextView()
	view.SetBorder(true)
	view.SetTitle(title)
	view.SetTitleAlign(tview.AlignCenter)
	view.SetScrollable(true)
	view.SetWrap(true)
	view.SetTextColor(application.theme.Suite.Foreground)
	view.SetBackgroundColor(application.theme.Suite.BackgroundFocus)
	view.SetBorderColor(application.theme.Suite.BorderFocus)
	view.SetTitleColor(application.theme.Suite.TitleFocus)
	return view
}

func centered(primitive tview.Primitive, width, height int) tview.Primitive {
	columns := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(primitive, width, 0, true).
		AddItem(nil, 0, 1, false)
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(columns, height, 0, true).
		AddItem(nil, 0, 1, false)
}

func (application *Application) openOverlay(overlay Overlay) {
	if application.Model.CurrentOverlay() == OverlayNone {
		application.previousFocus = application.Element.GetFocus()
	}
	application.Pages.HidePage(diagnosticsPage)
	application.Pages.HidePage(helpPage)

	var page string
	var focus tview.Primitive
	switch overlay {
	case OverlayDiagnostics:
		application.refreshDiagnostics()
		page = diagnosticsPage
		focus = application.Diagnostics
	case OverlayHelp:
		page = helpPage
		focus = application.Help
	default:
		application.closeOverlay()
		return
	}

	application.Model.update(func(state *State) {
		state.Overlay = overlay
	})
	application.Pages.ShowPage(page)
	application.Element.SetFocus(focus)
}

func (application *Application) closeOverlay() {
	application.Pages.HidePage(diagnosticsPage)
	application.Pages.HidePage(helpPage)
	application.Model.update(func(state *State) {
		state.Overlay = OverlayNone
	})
	if application.previousFocus != nil {
		application.Element.SetFocus(application.previousFocus)
	}
}

func (application *Application) refreshDiagnostics() {
	application.refreshStatus()
	if application.Diagnostics == nil || application.Model == nil {
		return
	}
	state := application.Model.Snapshot()
	application.Diagnostics.SetText(renderDiagnostics(state))
	application.Diagnostics.ScrollToBeginning()
	application.Diagnostics.SetTitle(fmt.Sprintf(
		"Diagnostics (%d) — d/Esc close",
		len(state.Diagnostics),
	))
}

func (application *Application) refreshStatus() {
	if application.Footer == nil || application.Model == nil {
		return
	}
	state := application.Model.Snapshot()
	application.Footer.UpdateIndicatorState(footerIndicatorState(state))
	status := ""
	switch {
	case state.Request.Phase == PhaseLoading && state.Request.HasProgress:
		application.stopFooterProgress()
		application.Footer.UpdateStatus("Running " + uiprogress.Body(
			state.Request.Current,
			state.Request.Total,
			footerProgressWidth,
			footerProgressPulse,
		))
		return
	case state.Request.Phase == PhaseLoading:
		application.showFooterProgress("Running")
		return
	case state.Parser.Phase == PhaseLoading:
		application.showFooterProgress("Parsing")
		return
	case state.Startup.Phase == PhaseLoading || state.Files.Phase == PhaseLoading:
		application.showFooterProgress("Loading")
		return
	case state.Startup.Phase == PhaseFailed || state.Files.Phase == PhaseFailed || state.Parser.Phase == PhaseFailed:
		status = "Error — press d"
	case state.Request.Outcome == OutcomeSuccess:
		status = "Success"
	case state.Request.Outcome == OutcomeFailure:
		status = "Failed"
	case len(state.Diagnostics) > 0:
		status = fmt.Sprintf("%d diagnostics — press d", len(state.Diagnostics))
	}
	application.stopFooterProgress()
	application.Footer.UpdateStatus(status)
}

func footerIndicatorState(state State) footer.IndicatorState {
	switch {
	case state.Request.Phase == PhaseLoading || state.Parser.Phase == PhaseLoading ||
		state.Startup.Phase == PhaseLoading || state.Files.Phase == PhaseLoading:
		return footer.IndicatorDefault
	case state.Startup.Phase == PhaseFailed || state.Files.Phase == PhaseFailed ||
		state.Parser.Phase == PhaseFailed || state.Request.Outcome == OutcomeFailure:
		return footer.IndicatorFailure
	case state.Request.Outcome == OutcomeSuccess:
		return footer.IndicatorSuccess
	default:
		return footer.IndicatorDefault
	}
}

func renderDiagnostics(state State) string {
	var sections []string
	if state.Startup.Phase == PhaseLoading {
		sections = append(sections, "Startup\nLoading environment...")
	} else if state.Startup.Phase == PhaseFailed {
		sections = append(sections, "Startup error\n"+state.Startup.Error)
	}
	if state.Files.Phase == PhaseLoading {
		sections = append(sections, "File discovery\nScanning "+state.RootDirectoryPath+"...")
	} else if state.Files.Phase == PhaseFailed {
		sections = append(sections, "File discovery error\n"+state.Files.Error)
	} else if len(state.Directory.Warnings) > 0 {
		sections = append(sections, "File discovery warnings\n- "+strings.Join(state.Directory.Warnings, "\n- "))
	}

	fileName := ""
	if state.SelectedFile != nil {
		fileName = state.SelectedFile.Path
	}
	switch state.Parser.Phase {
	case PhaseLoading:
		sections = append(sections, "Parser\nParsing "+fileName+"...")
	case PhaseFailed:
		sections = append(sections, "Parser error\n"+state.Parser.Error)
	case PhaseReady:
		if len(state.Diagnostics) == 0 {
			sections = append(sections, "Parser\nNo diagnostics for "+fileName+".")
		} else {
			lines := make([]string, 0, len(state.Diagnostics)+1)
			lines = append(lines, "Parser diagnostics for "+fileName)
			for _, diagnostic := range state.Diagnostics {
				lines = append(lines, "- "+diagnostic.String())
			}
			sections = append(sections, strings.Join(lines, "\n"))
		}
	default:
		if state.SelectedFile == nil {
			sections = append(sections, "Parser\nSelect an .http file to see parser diagnostics.")
		}
	}

	if len(sections) == 0 {
		return "No diagnostics."
	}
	return strings.Join(sections, "\n\n")
}

const helpText = `Global
  ?              Open or close this help
  d              Open or close diagnostics
  q / Ctrl+C     Quit lazyrest
  Ctrl+h/j/k/l   Move between panels

Files
  Enter          Open a directory or parse a request file
  Ctrl+l         Parse the selected file and open Suites
  /              Search files
  n / N          Next / previous match
  r              Reload files in the background

Suites and Suite
  j / k          Move between requests
  Enter          Open or execute the selected request
  Esc            Go back

Producer
  /              Search the response
  p              Toggle Pretty / Raw body
  [ / ]          Previous / next history entry
  Esc            Cancel the active request and go back

Search input captures printable keys. Press Enter or Esc to finish searching.`
