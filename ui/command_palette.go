package ui

import (
	appconfig "github.com/Hecatoncheir/lazyrest/config"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
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
	palette.AddItem(translator.Text("reload_files_command"), "", 0, func() {
		application.closeOverlay()
		onReloadFiles(application)()
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
	application.HttpFilesTree.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Suites.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Suite.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Producer.ApplySettings(settings.Theme, settings.Locale, settings.Keybindings)
	application.Footer.ApplySettings(settings.Theme, settings.Locale)
	application.buildOverlays()
	application.Footer.UpdateIndicatorState(footer.IndicatorDefault)
	application.Footer.UpdateStatus(settings.Locale.Text("config_reloaded"))
}
