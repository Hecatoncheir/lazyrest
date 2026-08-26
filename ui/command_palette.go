package ui

import (
	appconfig "github.com/Hecatoncheir/lazyrest/config"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func (application *Application) buildCommandPalette() {
	translator := application.config.Locale
	palette := tview.NewList().ShowSecondaryText(false)
	palette.SetBorder(true).
		SetTitle(translator.Text("command_palette")).
		SetTitleAlign(tview.AlignCenter)
	palette.AddItem(translator.Text("reload_config"), "", 0, func() {
		application.closeOverlay()
		application.reloadConfiguration()
	})
	palette.AddItem(translator.Text("choose_theme"), "", 0, func() {
		application.openOverlay(OverlayThemePicker)
	})
	palette.AddItem(translator.Text("reload_files_command"), "", 0, func() {
		application.closeOverlay()
		onReloadFiles(application)()
	})
	palette.AddItem(translator.Text("copy_response_body"), "", 0, func() {
		application.closeOverlay()
		application.copyResponse(false)
	})
	palette.AddItem(translator.Text("copy_response"), "", 0, func() {
		application.closeOverlay()
		application.copyResponse(true)
	})
	palette.AddItem(translator.Text("save_response"), "", 0, func() {
		application.openSaveResponse(false)
	})
	palette.AddItem(translator.Text("save_full_response"), "", 0, func() {
		application.openSaveResponse(true)
	})
	palette.AddItem(translator.Text("captured_responses"), "", 0, func() {
		application.openOverlay(OverlayCaptured)
	})
	palette.AddItem(translator.Text("diagnostics"), "", 0, func() {
		application.openOverlay(OverlayDiagnostics)
	})
	palette.AddItem(translator.Text("help"), "", 0, func() {
		application.openOverlay(OverlayHelp)
	})
	palette.AddItem(translator.Text("quit"), "", 0, func() {
		stopApplication(application)
	})
	application.applyCommandPaletteTheme(palette)
	application.CommandPalette = palette
	application.buildThemePicker()
}

func (application *Application) buildThemePicker() {
	translator := application.config.Locale
	picker := tview.NewList().ShowSecondaryText(false)
	picker.SetBorder(true).
		SetTitle(translator.Text("theme_picker")).
		SetTitleAlign(tview.AlignCenter)
	for _, preset := range theme.PresetNames() {
		name := preset
		picker.AddItem(name, "", 0, func() {
			application.closeOverlay()
			application.selectThemePreset(name)
		})
	}
	application.applyCommandPaletteTheme(picker)
	application.ThemePicker = picker
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
	for _, list := range []*tview.List{application.CommandPalette, application.ThemePicker} {
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
	application.Footer.ApplySettings(settings.Theme, settings.Locale)
	application.buildOverlays()
	if focused != nil {
		application.Element.SetFocus(focused)
	}
	application.Footer.UpdateIndicatorState(footer.IndicatorDefault)
	application.Footer.UpdateStatus(settings.Locale.Text("config_reloaded"))
}
