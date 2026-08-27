package ui

import (
	"fmt"
	"strings"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	uiprogress "github.com/Hecatoncheir/lazyrest/ui/progress"
	"github.com/rivo/tview"
)

const (
	diagnosticsPage       = "diagnostics"
	helpPage              = "help"
	capturedPage          = "captured-responses"
	historyPage           = "history"
	commandPalettePage    = "command-palette"
	themePickerPage       = "theme-picker"
	environmentPickerPage = "environment-picker"
	saveResponsePage      = "save-response"
)

func (application *Application) buildOverlays() {
	translator := application.config.Locale
	application.Diagnostics = application.newOverlayView(translator.Text("diagnostics") + " — d/Esc " + translator.Text("close"))
	application.Help = application.newOverlayView(translator.Text("help") + " — ?/Esc " + translator.Text("close"))
	application.Help.SetText(helpText(application.config.Keybindings, translator))
	application.buildCapturedResponsesOverlay()
	application.buildHistoryOverlay()
	application.buildEnvironmentPicker()
	application.buildCommandPalette()
	application.buildSaveResponseInput()

	application.Pages.
		AddPage(diagnosticsPage, centered(application.Diagnostics, 84, 24), true, false).
		AddPage(helpPage, centered(application.Help, 72, 25), true, false).
		AddPage(capturedPage, centered(application.Captured, 84, 24), true, false).
		AddPage(historyPage, centered(application.History, 92, 24), true, false)
	application.Pages.AddPage(commandPalettePage, centered(application.CommandPalette, 58, 19), true, false)
	application.Pages.AddPage(themePickerPage, centered(application.ThemePicker, 58, 14), true, false)
	application.Pages.AddPage(environmentPickerPage, centered(application.EnvironmentPicker, 64, 16), true, false)
	application.Pages.AddPage(saveResponsePage, centered(application.SaveResponse, 92, 3), true, false)
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
	application.Pages.HidePage(capturedPage)
	application.Pages.HidePage(historyPage)
	application.Pages.HidePage(commandPalettePage)
	application.Pages.HidePage(themePickerPage)
	application.Pages.HidePage(environmentPickerPage)
	application.Pages.HidePage(saveResponsePage)

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
	case OverlayCaptured:
		application.refreshCapturedResponses()
		page = capturedPage
		focus = application.Captured
	case OverlayHistory:
		application.refreshHistory()
		page = historyPage
		focus = application.History
	case OverlayCommandPalette:
		page = commandPalettePage
		focus = application.CommandPalette
	case OverlayThemePicker:
		page = themePickerPage
		focus = application.ThemePicker
	case OverlayEnvironmentPicker:
		application.refreshEnvironmentPicker()
		page = environmentPickerPage
		focus = application.EnvironmentPicker
	case OverlaySaveResponse:
		page = saveResponsePage
		focus = application.SaveResponse
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
	wasSaveResponse := application.Model.CurrentOverlay() == OverlaySaveResponse
	application.Pages.HidePage(diagnosticsPage)
	application.Pages.HidePage(helpPage)
	application.Pages.HidePage(capturedPage)
	application.Pages.HidePage(historyPage)
	application.Pages.HidePage(commandPalettePage)
	application.Pages.HidePage(themePickerPage)
	application.Pages.HidePage(environmentPickerPage)
	application.Pages.HidePage(saveResponsePage)
	application.Model.update(func(state *State) {
		state.Overlay = OverlayNone
	})
	if wasSaveResponse {
		application.pendingExport = nil
		application.saveFullResponse = false
		application.saveOverwritePath = ""
	}
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
	application.Diagnostics.SetText(renderDiagnosticsWithLocale(state, application.config.Locale))
	application.Diagnostics.ScrollToBeginning()
	application.Diagnostics.SetTitle(fmt.Sprintf(
		application.config.Locale.Text("diagnostics")+" (%d) — d/Esc "+application.config.Locale.Text("close"),
		diagnosticsCount(state),
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
		application.Footer.UpdateStatus(application.config.Locale.Text("running") + " " + uiprogress.BodyLocalized(
			state.Request.Current,
			state.Request.Total,
			footerProgressWidth,
			footerProgressPulse,
			application.config.Locale,
		))
		return
	case state.Request.Phase == PhaseLoading:
		application.showFooterProgress(application.config.Locale.Text("running"))
		return
	case state.Parser.Phase == PhaseLoading:
		application.showFooterProgress(application.config.Locale.Text("parsing"))
		return
	case state.Environment.Phase == PhaseLoading:
		application.showFooterProgress(application.config.Locale.Text("loading_environment"))
		return
	case state.Startup.Phase == PhaseLoading || state.Files.Phase == PhaseLoading:
		application.showFooterProgress(application.config.Locale.Text("loading_short"))
		return
	case state.Environment.Phase == PhaseFailed || state.Startup.Phase == PhaseFailed || state.Files.Phase == PhaseFailed || state.Parser.Phase == PhaseFailed:
		status = application.config.Locale.Text("error_press")
	case len(state.HistoryErrors) > 0:
		status = application.config.Locale.PluralDiagnostics(diagnosticsCount(state)) + " — " + application.config.Locale.Text("press_d")
	case state.Request.Outcome == OutcomeSuccess:
		status = application.config.Locale.Text("success")
	case state.Request.Outcome == OutcomeFailure:
		status = application.config.Locale.Text("failed")
	case len(state.Diagnostics) > 0:
		status = application.config.Locale.PluralDiagnostics(len(state.Diagnostics)) + " — " + application.config.Locale.Text("press_d")
	}
	application.stopFooterProgress()
	application.Footer.UpdateStatus(status)
}

func footerIndicatorState(state State) footer.IndicatorState {
	switch {
	case state.Request.Phase == PhaseLoading || state.Parser.Phase == PhaseLoading || state.Environment.Phase == PhaseLoading ||
		state.Startup.Phase == PhaseLoading || state.Files.Phase == PhaseLoading:
		return footer.IndicatorDefault
	case state.Environment.Phase == PhaseFailed || state.Startup.Phase == PhaseFailed || state.Files.Phase == PhaseFailed ||
		state.Parser.Phase == PhaseFailed || state.Request.Outcome == OutcomeFailure || len(state.HistoryErrors) > 0:
		return footer.IndicatorFailure
	case state.Request.Outcome == OutcomeSuccess:
		return footer.IndicatorSuccess
	default:
		return footer.IndicatorDefault
	}
}

func renderDiagnosticsWithLocale(state State, translator *locale.Translator) string {
	var sections []string
	switch state.Environment.Phase {
	case PhaseLoading:
		sections = append(sections, translator.Text("environment")+"\n"+translator.Text("loading_environment"))
	case PhaseFailed:
		sections = append(sections, translator.Text("environment_error")+"\n"+state.Environment.Error)
	}
	switch state.Startup.Phase {
	case PhaseLoading:
		sections = append(sections, translator.Text("startup")+"\n"+translator.Text("loading_environment"))
	case PhaseFailed:
		sections = append(sections, translator.Text("startup_error")+"\n"+state.Startup.Error)
	}
	switch {
	case state.Files.Phase == PhaseLoading:
		sections = append(sections, translator.Text("file_discovery")+"\n"+translator.Format("scanning", state.RootDirectoryPath))
	case state.Files.Phase == PhaseFailed:
		sections = append(sections, translator.Text("file_discovery_error")+"\n"+state.Files.Error)
	case len(state.Directory.Warnings) > 0:
		sections = append(sections, translator.Text("file_discovery_warnings")+"\n- "+strings.Join(state.Directory.Warnings, "\n- "))
	}
	if len(state.HistoryErrors) > 0 {
		sections = append(sections, translator.Text("history_persistence")+"\n- "+strings.Join(state.HistoryErrors, "\n- "))
	}

	fileName := ""
	if state.SelectedFile != nil {
		fileName = state.SelectedFile.Path
	}
	switch state.Parser.Phase {
	case PhaseLoading:
		sections = append(sections, translator.Text("parser")+"\n"+translator.Format("parser_parsing", fileName))
	case PhaseFailed:
		sections = append(sections, translator.Text("parser_error")+"\n"+state.Parser.Error)
	case PhaseReady:
		if len(state.Diagnostics) == 0 {
			sections = append(sections, translator.Text("parser")+"\n"+translator.Format("no_diagnostics_for", fileName))
		} else {
			lines := make([]string, 0, len(state.Diagnostics)+1)
			lines = append(lines, translator.Format("parser_diagnostics_for", fileName))
			for _, diagnostic := range state.Diagnostics {
				severity := translator.Text("diagnostic_warning")
				if diagnostic.IsBlocking() {
					severity = translator.Text("diagnostic_error")
				}
				lines = append(lines, fmt.Sprintf("- %s: %s", severity, diagnostic.String()))
			}
			sections = append(sections, strings.Join(lines, "\n"))
		}
	default:
		if state.SelectedFile == nil {
			sections = append(sections, translator.Text("parser")+"\n"+translator.Text("select_file_diagnostics"))
		}
	}

	if len(sections) == 0 {
		return translator.Text("no_diagnostics")
	}
	return strings.Join(sections, "\n\n")
}

func helpText(bindings *keymap.Bindings, translator *locale.Translator) string {
	line := func(action keymap.Action, description string) string {
		return fmt.Sprintf("  %-20s %s", bindings.Describe(action), description)
	}
	return strings.Join([]string{
		translator.Text("global"),
		line(keymap.Help, translator.Text("help_toggle")),
		line(keymap.Diagnostics, translator.Text("diagnostics_toggle")),
		line(keymap.Quit, translator.Text("quit_help")),
		line(keymap.FocusLeft, translator.Text("focus_left")),
		line(keymap.FocusDown, translator.Text("focus_down")),
		line(keymap.FocusUp, translator.Text("focus_up")),
		line(keymap.FocusRight, translator.Text("focus_right")),
		line(keymap.HalfPageDown, translator.Text("half_page_down")),
		line(keymap.HalfPageUp, translator.Text("half_page_up")),
		line(keymap.PageDown, translator.Text("page_down")),
		line(keymap.PageUp, translator.Text("page_up")),
		line(keymap.GoToTop, translator.Text("go_to_top")),
		line(keymap.GoToBottom, translator.Text("go_to_bottom")),
		line(keymap.AlignTop, translator.Text("align_top")),
		line(keymap.CenterView, translator.Text("center_view")),
		line(keymap.AlignBottom, translator.Text("align_bottom")),
		line(keymap.CommandPalette, translator.Text("command_palette")),
		line(keymap.ReloadConfig, translator.Text("reload_config")),
		"",
		translator.Text("files_help"),
		line(keymap.Open, translator.Text("open_file")),
		line(keymap.Search, translator.Text("search_files")),
		line(keymap.SearchNext, translator.Text("next_match")),
		line(keymap.SearchPrevious, translator.Text("previous_match")),
		line(keymap.Reload, translator.Text("reload_files")),
		"",
		translator.Text("suites_help"),
		line(keymap.MoveDown, translator.Text("move_down")),
		line(keymap.MoveUp, translator.Text("move_up")),
		line(keymap.Open, translator.Text("open_request")),
		line(keymap.Run, translator.Text("execute_request")),
		line(keymap.Back, translator.Text("go_back")),
		"",
		translator.Text("producer_help"),
		line(keymap.Search, translator.Text("search_response")),
		line(keymap.SearchNext, translator.Text("next_match")),
		line(keymap.SearchPrevious, translator.Text("previous_match")),
		line(keymap.ToggleBody, translator.Text("toggle_body")),
		line(keymap.RerunRequest, translator.Text("rerun_request")),
		line(keymap.CopyResponseBody, translator.Text("copy_response_body")),
		line(keymap.CopyResponse, translator.Text("copy_response")),
		line(keymap.CopyAsCurl, translator.Text("copy_as_curl")),
		line(keymap.SaveResponse, translator.Text("save_response")),
		line(keymap.SaveFullResponse, translator.Text("save_full_response")),
		line(keymap.HistoryPrevious, translator.Text("previous_history")),
		line(keymap.HistoryNext, translator.Text("next_history")),
		line(keymap.Back, translator.Text("cancel_back")),
		"",
		translator.Format("search_finish", bindings.Describe(keymap.SearchFinish)),
	}, "\n")
}
