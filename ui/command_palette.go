package ui

import (
	"fmt"
	"reflect"
	"strings"

	appconfig "github.com/Hecatoncheir/lazyrest/config"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

type commandEntry struct {
	label    string
	selected func()
}

func (application *Application) buildCommandPalette() {
	translator := application.config.Locale
	palette := tview.NewList().ShowSecondaryText(false)
	palette.SetBorder(true).
		SetTitleAlign(tview.AlignCenter)
	application.commandItems = []commandEntry{
		{label: translator.Text("reload_config"), selected: func() {
			application.closeOverlay()
			application.reloadConfiguration()
		}},
		{label: translator.Text("choose_theme"), selected: func() {
			application.openOverlay(OverlayThemePicker)
		}},
		{label: translator.Text("choose_environment"), selected: func() {
			application.openOverlay(OverlayEnvironmentPicker)
		}},
		{label: translator.Text("reload_files_command"), selected: func() {
			application.closeOverlay()
			onReloadFiles(application)()
		}},
		{label: translator.Text("rerun_request"), selected: func() {
			application.closeOverlay()
			application.rerunCurrentRequest()
		}},
		{label: translator.Text("copy_response_body"), selected: func() {
			application.closeOverlay()
			application.copyResponse(false)
		}},
		{label: translator.Text("copy_response"), selected: func() {
			application.closeOverlay()
			application.copyResponse(true)
		}},
		{label: translator.Text("copy_as_curl"), selected: func() {
			application.closeOverlay()
			application.copyAsCurl()
		}},
		{label: translator.Text("save_response"), selected: func() {
			application.openSaveResponse(false)
		}},
		{label: translator.Text("save_full_response"), selected: func() {
			application.openSaveResponse(true)
		}},
		{label: translator.Text("captured_responses"), selected: func() {
			application.openOverlay(OverlayCaptured)
		}},
		{label: translator.Text("history_window"), selected: func() {
			application.openOverlay(OverlayHistory)
		}},
		{label: translator.Text("diagnostics"), selected: func() {
			application.openOverlay(OverlayDiagnostics)
		}},
		{label: translator.Text("help"), selected: func() {
			application.openOverlay(OverlayHelp)
		}},
		{label: translator.Text("quit"), selected: func() {
			stopApplication(application)
		}},
	}
	application.applyCommandPaletteTheme(palette)
	application.CommandPalette = palette
	application.resetCommandPaletteSearch()
	application.buildThemePicker()
}

func (application *Application) renderCommandPalette() {
	if application.CommandPalette == nil {
		return
	}
	palette := application.CommandPalette
	palette.Clear()
	query := strings.ToLower(strings.TrimSpace(application.commandQuery))
	application.commandMatches = 0
	for _, item := range application.commandItems {
		if query != "" && !strings.Contains(strings.ToLower(item.label), query) {
			continue
		}
		application.commandMatches++
		palette.AddItem(item.label, "", 0, item.selected)
	}
	if application.commandMatches == 0 {
		palette.AddItem(application.config.Locale.Format("no_command_results", application.commandQuery), "", 0, nil)
	}
	title := fmt.Sprintf("%s (%d/%d)", application.config.Locale.Text("command_palette"), application.commandMatches, len(application.commandItems))
	if application.commandSearchMode || application.commandQuery != "" {
		title = fmt.Sprintf("%s /%s (%d/%d)", application.config.Locale.Text("command_palette"), tview.Escape(application.commandQuery), application.commandMatches, len(application.commandItems))
	}
	palette.SetTitle(title)
}

func (application *Application) resetCommandPaletteSearch() {
	application.commandQuery = ""
	application.commandSearchMode = false
	application.renderCommandPalette()
}

func (application *Application) buildThemePicker() {
	picker := tview.NewList().ShowSecondaryText(false)
	picker.SetBorder(true).
		SetTitleAlign(tview.AlignCenter)
	application.ThemePicker = picker
	application.refreshThemePicker()
	application.applyCommandPaletteTheme(picker)
}

func (application *Application) refreshThemePicker() {
	if application.ThemePicker == nil {
		return
	}
	picker := application.ThemePicker
	picker.Clear()
	picker.SetTitle(application.config.Locale.Text("theme_picker"))
	selected := 0
	for _, preset := range theme.PresetNames() {
		name := preset
		label := name
		configured, err := theme.FromConfig(theme.Config{Preset: name})
		if err == nil && reflect.DeepEqual(configured, application.theme) {
			label = "✓ " + label
			selected = picker.GetItemCount()
		}
		picker.AddItem(label, "", 0, func() {
			application.closeOverlay()
			application.selectThemePreset(name)
		})
	}
	picker.SetCurrentItem(selected)
}

func (application *Application) selectThemePreset(name string) {
	selected, err := theme.FromConfig(theme.Config{Preset: name})
	if err != nil {
		application.Footer.UpdateIndicatorState(footer.IndicatorFailure)
		application.Footer.UpdateStatus(application.config.Locale.Format("config_error", err))
		return
	}
	application.config.Theme = selected
	application.theme = selected
	focused := application.Element.GetFocus()
	application.Pages.SetBackgroundColor(selected.Background)
	application.Layout.ApplySettings(selected)
	application.Workspace.ApplySettings(selected)
	application.HttpFilesTree.ApplySettings(selected, application.config.Locale, application.config.Keybindings)
	application.Suites.ApplySettings(selected, application.config.Locale, application.config.Keybindings)
	application.Suite.ApplySettings(selected, application.config.Locale, application.config.Keybindings)
	application.Producer.ApplySettings(selected, application.config.Locale, application.config.Keybindings)
	application.Footer.ApplySettings(selected, application.config.Locale)
	application.applyOverlayTheme()
	application.refreshThemePicker()
	if focused != nil {
		application.Element.SetFocus(focused)
	}
	application.Footer.UpdateIndicatorState(footer.IndicatorDefault)
	application.Footer.UpdateStatus(application.config.Locale.Format("theme_changed", name))
}

func (application *Application) applyOverlayTheme() {
	for _, view := range []*tview.TextView{application.Diagnostics, application.Help, application.Captured} {
		if view == nil {
			continue
		}
		uiTheme := application.theme.Suite
		view.SetTextColor(uiTheme.Foreground).
			SetBackgroundColor(uiTheme.BackgroundFocus).
			SetBorderColor(uiTheme.BorderFocus).
			SetTitleColor(uiTheme.TitleFocus)
	}
	for _, list := range []*tview.List{application.CommandPalette, application.ThemePicker, application.EnvironmentPicker, application.History} {
		if list != nil {
			application.applyCommandPaletteTheme(list)
		}
	}
	application.applySaveResponseTheme()
}

func (application *Application) applyCommandPaletteTheme(palette *tview.List) {
	uiTheme := application.theme.Suites
	palette.SetTitleColor(uiTheme.TitleFocus).
		SetBackgroundColor(uiTheme.BackgroundFocus).
		SetBorderColor(uiTheme.BorderFocus)
	palette.SetMainTextColor(uiTheme.SuiteForeground).
		SetSecondaryTextColor(uiTheme.SuiteForeground).
		SetSelectedTextColor(uiTheme.SuiteFocusForeground).
		SetSelectedBackgroundColor(uiTheme.SuiteFocusBackground)
}

func (application *Application) reloadConfiguration() {
	application.stopFooterProgress()
	path := application.config.ConfigPath
	if path == "" {
		var err error
		path, err = appconfig.DefaultPath()
		if err != nil {
			application.Footer.UpdateIndicatorState(footer.IndicatorFailure)
			application.Footer.UpdateStatus(application.config.Locale.Format("config_error", err))
			return
		}
	}
	paths := application.config.ConfigPaths
	if len(paths) == 0 {
		paths = []string{path}
	}
	settings, err := appconfig.LoadFiles(paths)
	if err != nil {
		application.Footer.UpdateIndicatorState(footer.IndicatorFailure)
		application.Footer.UpdateStatus(application.config.Locale.Format("config_error", err))
		return
	}
	application.config.Keybindings = settings.Keybindings
	application.config.Locale = settings.Locale
	application.config.Theme = settings.Theme
	application.config.HistoryBodies = !settings.HistoryMetadata
	application.config.ConfigPath = path
	application.theme = settings.Theme
	focused := application.Element.GetFocus()
	application.Pages.SetBackgroundColor(settings.Theme.Background)
	application.Layout.ApplySettings(settings.Theme)
	application.Workspace.ApplySettings(settings.Theme)
	application.HttpFilesTree.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Suites.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Suite.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Producer.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Producer.SetHistoryMode(historyMode(!settings.HistoryMetadata))
	application.Footer.ApplySettings(settings.Theme, settings.Locale)
	application.buildOverlays()
	if focused != nil {
		application.Element.SetFocus(focused)
	}
	application.Footer.UpdateIndicatorState(footer.IndicatorDefault)
	application.Footer.UpdateStatus(settings.Locale.Text("config_reloaded"))
}
