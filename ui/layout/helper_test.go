package layout

import (
	"testing"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/Hecatoncheir/lazyrest/ui/workspace"
)

// buildWorkspace assembles the panes a layout is made of.
func buildWorkspace(t *testing.T, uiTheme theme.Theme) *workspace.Workspace {
	t.Helper()
	treeWidget := tree.New()
	treeWidget.Build(tree.Parameters{RootDirectoryPath: t.TempDir(), Theme: uiTheme})
	suitesWidget := suites.New()
	suitesWidget.Build(suites.Parameters{Theme: uiTheme, OnEscapeCallback: func() {}, OnSuiteSelectCallbackType: func(parserhttp.HttpSuite) {}})
	suiteWidget := suite.New()
	suiteWidget.Build(suite.Parameters{Theme: uiTheme, OnEscapeCallback: func() {}, OnRunCallback: func(parserhttp.HttpSuite) {}})
	producerWidget := producer.New()
	producerWidget.Build(producer.Parameters{Theme: uiTheme, OnEscapeCallback: func() {}})

	workspaceWidget := workspace.New()
	workspaceWidget.Build(workspace.Parameters{Theme: uiTheme}, treeWidget, suitesWidget, suiteWidget, producerWidget)
	return workspaceWidget
}
