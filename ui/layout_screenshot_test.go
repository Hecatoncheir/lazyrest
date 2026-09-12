package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestResponsiveLayoutScreenshotsAtSupportedWidths(t *testing.T) {
	application := BuildApplication(t.TempDir(), Config{})
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()

	for _, testCase := range []struct {
		width          int
		workspacePanes int
		onboarding     bool
	}{
		{width: 60, workspacePanes: 1, onboarding: false},
		{width: 80, workspacePanes: 2, onboarding: true},
		{width: 120, workspacePanes: 3, onboarding: true},
		{width: 160, workspacePanes: 3, onboarding: true},
	} {
		t.Run(fmt.Sprintf("width-%d", testCase.width), func(t *testing.T) {
			screen.SetSize(testCase.width, 24)
			application.Layout.Element.SetRect(0, 0, testCase.width, 24)
			application.updateResponsiveUI(screen)
			application.Layout.Element.Draw(screen)
			screen.Sync()

			if got := application.Workspace.Element.(*tview.Flex).GetItemCount(); got != testCase.workspacePanes {
				t.Fatalf("width %d rendered %d workspace panes, want %d", testCase.width, got, testCase.workspacePanes)
			}
			snapshot := simulationText(screen)
			if !strings.Contains(snapshot, "Files") {
				t.Fatalf("width %d screenshot does not contain Files title:\n%s", testCase.width, snapshot)
			}
			hasCommands := strings.Contains(snapshot, ": commands")
			if hasCommands != testCase.onboarding {
				t.Fatalf("width %d screenshot onboarding commands=%v, want %v:\n%s", testCase.width, hasCommands, testCase.onboarding, snapshot)
			}
		})
	}
}
