package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/ui/footer"
)

func TestCorruptedHistoryAppearsInApplicationDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}

	application := BuildApplication(t.TempDir(), Config{HistoryPath: path})
	state := application.Model.Snapshot()
	if len(state.HistoryErrors) != 1 || diagnosticsCount(state) != 1 {
		t.Fatalf("history error was not added to diagnostics: %+v", state.HistoryErrors)
	}
	rendered := application.Diagnostics.GetText(false)
	for _, expected := range []string{"History persistence", "load history", path} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("diagnostics %q do not contain %q", rendered, expected)
		}
	}
	if !strings.Contains(application.Diagnostics.GetTitle(), "(1)") {
		t.Fatalf("history error was not counted in the title: %q", application.Diagnostics.GetTitle())
	}
	if got := footerIndicatorState(state); got != footer.IndicatorFailure {
		t.Fatalf("history failure did not set the failure indicator: %v", got)
	}
}

func TestHistoryDiagnosticsAreDeduplicatedAndBounded(t *testing.T) {
	application := &Application{Model: NewModel(t.TempDir(), "")}
	application.recordHistoryError(os.ErrPermission)
	application.recordHistoryError(os.ErrPermission)
	for index := 0; index < maxHistoryErrors; index++ {
		application.recordHistoryError(filepath.ErrBadPattern)
		application.recordHistoryError(&os.PathError{Op: "write", Path: string(rune('a' + index)), Err: os.ErrInvalid})
	}
	errors := application.Model.Snapshot().HistoryErrors
	if len(errors) != maxHistoryErrors {
		t.Fatalf("history diagnostics were not deduplicated and bounded: %d", len(errors))
	}
}
