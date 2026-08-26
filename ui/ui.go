package ui

import (
	"context"

	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/Hecatoncheir/lazyrest/ui/layout"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/Hecatoncheir/lazyrest/ui/workspace"
	"github.com/rivo/tview"
)

func Run(rootDirectoryPath string, config Config) error {
	applicationWidget := BuildApplication(rootDirectoryPath, config)
	applicationWidget.Start()
	defer applicationWidget.stopFooterProgress()
	defer applicationWidget.Producer.WaitForHistory()
	return applicationWidget.Element.Run()
}

func BuildApplication(rootDirectoryPath string, config Config) *Application {
	uiTheme := config.Theme
	if uiTheme == (theme.Theme{}) {
		uiTheme = theme.NewDefault()
	}
	if config.Keybindings == nil {
		config.Keybindings = keymap.Default()
	}
	if config.Locale == nil {
		config.Locale = locale.English()
	}
	environmentName := config.Environment.Name
	if environmentName == "" {
		environmentName = config.EnvironmentName
	}

	// Application
	applicationWidget := NewApplication()
	applicationElement := applicationWidget.Build()
	applicationWidget.Model = NewModel(rootDirectoryPath, environmentName)
	applicationWidget.config = config
	applicationWidget.theme = uiTheme
	applicationWidget.loadEnvironment = environment.Load

	// HttpFilesTree
	httpFilesExtensions := []string{".http", ".hurl"}
	httpFilesTreeWidget := tree.New()
	httpFilesTreeParameters := tree.Parameters{
		RootDirectoryPath:    rootDirectoryPath,
		Theme:                uiTheme,
		FilesExtension:       httpFilesExtensions,
		Ignore:               config.Ignore,
		OnSelectFileCallback: onSelectFileCallback(applicationWidget),
		OnReloadCallback:     onReloadFiles(applicationWidget),
		Keybindings:          config.Keybindings,
		Locale:               config.Locale,
	}
	httpFilesTreeElement := httpFilesTreeWidget.Build(httpFilesTreeParameters)
	applicationWidget.HttpFilesTree = httpFilesTreeWidget
	applicationWidget.scanFiles = func(ctx context.Context) tree.ScanResult {
		return httpFilesTreeWidget.Scan(ctx)
	}

	// Suite
	suiteWidget := suite.New()
	suiteParameters := suite.Parameters{
		Theme:            uiTheme,
		OnEscapeCallback: onSuiteEscape(applicationWidget),
		OnRunCallback:    onSuiteRun(applicationWidget),
		Keybindings:      config.Keybindings,
		Locale:           config.Locale,
	}
	suiteWidget.Build(suiteParameters)
	applicationWidget.Suite = suiteWidget

	// Suites
	suitesWidget := suites.New()
	suitesParameters := suites.Parameters{
		Theme:                     uiTheme,
		OnEscapeCallback:          onSuitesEscape(applicationWidget),
		OnSuiteSelectCallbackType: onSuiteSelect(applicationWidget),
		ParseOptions:              config.ParseOptions,
		Keybindings:               config.Keybindings,
		Locale:                    config.Locale,
	}
	suitesWidget.Build(suitesParameters)
	applicationWidget.Suites = suitesWidget

	// Producer
	producerWidget := producer.New()
	producerParameters := producer.Parameters{
		Theme:                 uiTheme,
		OnEscapeCallback:      onProducerEscape(applicationWidget),
		OnProgressCallback:    onRunProgress(applicationWidget),
		OnRunFinishedCallback: onRunFinished(applicationWidget),
		App:                   applicationElement,
		RunnerConfig:          config.Runner,
		Keybindings:           config.Keybindings,
		Locale:                config.Locale,
		HistoryPath:           config.HistoryPath,
	}
	producerWidget.Build(producerParameters)
	applicationWidget.Producer = producerWidget

	// Workspace
	workspaceWidget := workspace.New()
	workspaceParameters := workspace.Parameters{
		RootDirectoryPath: rootDirectoryPath,
		Theme:             uiTheme,
	}
	workspaceWidget.Build(
		workspaceParameters,
		httpFilesTreeWidget,
		suitesWidget,
		suiteWidget,
		producerWidget,
	)
	applicationWidget.Workspace = workspaceWidget

	// Footer
	footerWidget := footer.New()
	footerParameters := footer.Parameters{
		RootDirectoryPath: rootDirectoryPath,
		Theme:             uiTheme,
		EnvironmentName:   environmentName,
		Locale:            config.Locale,
	}
	footerWidget.Build(footerParameters)
	applicationWidget.Footer = footerWidget

	// Layout
	layoutWidget := layout.New()
	layoutParameters := layout.Parameters{
		RootDirectoryPath: rootDirectoryPath,
		Theme:             uiTheme,
	}
	layoutElement := layoutWidget.Build(
		layoutParameters,
		workspaceWidget,
		footerWidget,
	)
	applicationWidget.Layout = layoutWidget

	pages := tview.NewPages().
		AddPage("main", layoutElement, true, true)
	pages.SetBackgroundColor(uiTheme.Background)
	applicationWidget.Pages = pages
	applicationWidget.buildOverlays()

	applicationElement.
		SetRoot(pages, true).
		SetFocus(httpFilesTreeElement)

	applicationElement.SetInputCapture(onInputCallback(applicationWidget))
	return applicationWidget
}
