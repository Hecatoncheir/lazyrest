package workspace

import (
	"testing"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/Hecatoncheir/lazyrest/ui/tree"

	"github.com/rivo/tview"
)

func build(t *testing.T, uiTheme theme.Theme) *Workspace {
	t.Helper()
	treeWidget := tree.New()
	treeWidget.Build(tree.Parameters{RootDirectoryPath: t.TempDir(), Theme: uiTheme})
	suitesWidget := suites.New()
	suitesWidget.Build(suites.Parameters{Theme: uiTheme, OnEscapeCallback: func() {}, OnSuiteSelectCallbackType: func(parserhttp.HttpSuite) {}})
	suiteWidget := suite.New()
	suiteWidget.Build(suite.Parameters{Theme: uiTheme, OnEscapeCallback: func() {}, OnRunCallback: func(parserhttp.HttpSuite) {}})
	producerWidget := producer.New()
	producerWidget.Build(producer.Parameters{Theme: uiTheme, OnEscapeCallback: func() {}})

	widget := New()
	widget.Build(Parameters{Theme: uiTheme}, treeWidget, suitesWidget, suiteWidget, producerWidget)
	return widget
}

func TestBuildAssemblesThePanes(t *testing.T) {
	uiTheme := theme.NewDefault()
	widget := build(t, uiTheme)

	if widget.Element == nil {
		t.Fatal("the workspace was not built")
	}
	box, ok := widget.Element.(*tview.Flex)
	if !ok {
		t.Fatalf("unexpected element: %T", widget.Element)
	}
	if got := box.GetBackgroundColor(); got != uiTheme.Background {
		t.Fatalf("got background %v, want %v", got, uiTheme.Background)
	}
	if widget.suitesArea == nil {
		t.Fatal("the column holding Suites and Suite is missing")
	}
}

func TestApplySettingsRepaintsTheInnerColumnToo(t *testing.T) {
	widget := build(t, theme.NewDefault())
	other, err := theme.FromConfig(theme.Config{Preset: "dracula"})
	if err != nil {
		t.Fatal(err)
	}

	widget.ApplySettings(other)

	if got := widget.Element.(*tview.Flex).GetBackgroundColor(); got != other.Background {
		t.Errorf("the workspace kept the old background: %v", got)
	}
	// The column between the tree and the producer is easy to forget, and a
	// missed repaint leaves a stripe of the previous theme on screen.
	if got := widget.suitesArea.GetBackgroundColor(); got != other.Background {
		t.Errorf("the inner column kept the old background: %v", got)
	}
}

func TestResizeAdaptsVisiblePanesToTerminalWidthAndFocus(t *testing.T) {
	widget := build(t, theme.NewDefault())
	box := widget.Element.(*tview.Flex)

	widget.Resize(wideLayoutWidth, widget.treeElement)
	if got := box.GetItemCount(); got != 3 {
		t.Fatalf("wide layout has %d panes, want 3", got)
	}

	widget.Resize(wideLayoutWidth-1, widget.treeElement)
	if got := box.GetItemCount(); got != 2 || box.GetItem(0) != widget.treeElement || box.GetItem(1) != widget.suitesArea {
		t.Fatalf("medium file layout did not keep files and requests")
	}

	widget.Resize(wideLayoutWidth-1, widget.producerElement)
	if got := box.GetItemCount(); got != 2 || box.GetItem(0) != widget.suitesArea || box.GetItem(1) != widget.producerElement {
		t.Fatalf("medium response layout did not keep requests and response")
	}

	widget.Resize(compactLayoutWidth-1, widget.producerElement)
	if got := box.GetItemCount(); got != 1 || box.GetItem(0) != widget.producerElement {
		t.Fatalf("compact layout did not isolate the active response pane")
	}
}
