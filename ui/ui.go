package ui

import (
	"lazyrest/ui/footer"
	"lazyrest/ui/layout"
	"lazyrest/ui/producer"
	"lazyrest/ui/suite"
	"lazyrest/ui/suites"
	"lazyrest/ui/theme"
	"lazyrest/ui/tree"
	"lazyrest/ui/workspace"
)

func Run(rootDirectoryPath string) error {
	uiTheme := theme.NewDefault()

	// Application
	applicationWidget := NewApplication()
	applicationElement := applicationWidget.Build()

	// HttpFilesTree
	httpFilesExtension := ".http"
	httpFilesTreeWidget := tree.New()
	httpFilesTreeParameters := tree.Parameters{
		RootDirectoryPath:    rootDirectoryPath,
		Theme:                uiTheme,
		FilesExtension:       httpFilesExtension,
		OnSelectFileCallback: onSelectFileCallback(&applicationWidget),
	}
	httpFilesTreeElement := httpFilesTreeWidget.Build(httpFilesTreeParameters)
	applicationWidget.HttpFilesTree = httpFilesTreeWidget

	// Suite
	suiteWidget := suite.New()
	suiteParameters := suite.Parameters{
		Theme:            uiTheme,
		OnEscapeCallback: onSuiteEscape(&applicationWidget),
		OnRunCallback:    onSuiteRun(&applicationWidget),
	}
	suiteWidget.Build(suiteParameters)
	applicationWidget.Suite = suiteWidget

	// Suites
	suitesWidget := suites.New()
	suitesParameters := suites.Parameters{
		Theme:                     uiTheme,
		OnEscapeCallback:          onSuitesEscape(&applicationWidget),
		OnSuiteSelectCallbackType: onSuiteSelect(&applicationWidget),
	}
	suitesWidget.Build(suitesParameters)
	applicationWidget.Suites = suitesWidget

	// Producer
	producerWidget := producer.New()
	producerParameters := producer.Parameters{
		Theme:            uiTheme,
		OnEscapeCallback: onProducerEscape(&applicationWidget),
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

	applicationElement.
		SetRoot(layoutElement, true).
		SetFocus(httpFilesTreeElement)

	applicationElement.SetInputCapture(onInputCallback(&applicationWidget))

	err := applicationElement.Run()
	if err != nil {
		return err
	}

	return nil
}
