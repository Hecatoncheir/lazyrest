package producer

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func TestHistoryWindowSummariesSelectAndClearEntries(t *testing.T) {
	historyPath := filepath.Join(t.TempDir(), "history.json")
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), HistoryPath: historyPath})
	firstTime := time.Date(2026, time.August, 25, 10, 0, 0, 0, time.UTC)
	widget.history = []HistoryEntry{
		{
			Suite:     parserhttp.HttpSuite{Name: "List users", Method: http.MethodGet, Uri: "https://example.test/users", SourceFilePath: "users.http"},
			Response:  runner.Response{Code: "200 OK", StatusCode: 200, Time: 20 * time.Millisecond, Body: "secret body"},
			CreatedAt: firstTime,
		},
		{
			Suite:     parserhttp.HttpSuite{Name: "Create user", Method: http.MethodPost, SourceFilePath: "users.http"},
			Response:  runner.Response{Time: 30 * time.Millisecond},
			Err:       errors.New("private failure detail"),
			CreatedAt: firstTime.Add(time.Minute),
		},
	}
	widget.historyIndex = 1
	widget.persistHistory()

	summaries := widget.HistorySummaries()
	if len(summaries) != 2 || summaries[0].Index != 1 || summaries[0].Name != "Create user" || !summaries[0].Failed {
		t.Fatalf("unexpected newest-first summaries: %+v", summaries)
	}
	if summaries[1].Status != "200 OK" || summaries[1].Duration != 20*time.Millisecond {
		t.Fatalf("unexpected successful summary: %+v", summaries[1])
	}
	if rendered := summaries[0].Name + summaries[0].Status; strings.Contains(rendered, "private failure") || strings.Contains(rendered, "secret body") {
		t.Fatalf("history summary exposed entry details: %q", rendered)
	}

	title := widget.Element.(*tview.TextView).GetTitle()
	if !widget.SelectHistory(0) || !strings.Contains(widget.currentText, "https://example.test/users") {
		t.Fatalf("history entry was not selected: title=%q text=%q", title, widget.currentText)
	}
	title = widget.Element.(*tview.TextView).GetTitle()
	if !strings.Contains(title, "1/2") {
		t.Fatalf("selected history position is missing from title: %q", title)
	}
	if removed := widget.ClearHistory(); removed != 2 || len(widget.HistorySummaries()) != 0 || widget.currentText != "" {
		t.Fatalf("history clear removed %d entries, summaries=%v, text=%q", removed, widget.HistorySummaries(), widget.currentText)
	}
	widget.WaitForHistory()
	restored := &Producer{historyPath: historyPath}
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 0 {
		t.Fatalf("cleared history was persisted again: %+v", restored.history)
	}
}
