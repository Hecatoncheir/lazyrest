package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/ui/producer"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (application *Application) buildHistoryOverlay() {
	history := tview.NewList().ShowSecondaryText(true)
	history.SetBorder(true).SetTitleAlign(tview.AlignCenter)
	history.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case application.config.Keybindings.Matches(keymap.ClearHistory, event):
			application.clearHistory()
			return nil
		case application.config.Keybindings.Matches(keymap.Open, event):
			return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		default:
			return event
		}
	})
	application.applyCommandPaletteTheme(history)
	application.History = history
	application.refreshHistory()
}

func (application *Application) refreshHistory() {
	if application.History == nil || application.Producer == nil {
		return
	}
	summaries := application.Producer.HistorySummaries()
	history := application.History
	history.Clear()
	translator := application.config.Locale
	if len(summaries) == 0 {
		history.AddItem(translator.Text("no_history_entries"), "", 0, nil)
	}
	for _, summary := range summaries {
		entryIndex := summary.Index
		history.AddItem(
			historyEntryTitle(summary, translator.Text("unnamed_request")),
			historyEntryDetails(summary, application.Model.Snapshot().RootDirectoryPath, application),
			0,
			func() { application.selectHistory(entryIndex) },
		)
	}
	history.SetTitle(fmt.Sprintf(
		"%s (%d) — %s %s · Enter %s · q/Esc %s",
		translator.Text("history_window"),
		len(summaries),
		application.config.Keybindings.Describe(keymap.ClearHistory),
		translator.Text("clear_history"),
		translator.Text("open_history_entry"),
		translator.Text("close"),
	))
}

func historyEntryTitle(summary producer.HistorySummary, unnamed string) string {
	name := singleLine(summary.Name)
	if name == "" {
		name = unnamed
	}
	return strings.TrimSpace(singleLine(summary.Method) + " " + name)
}

func historyEntryDetails(summary producer.HistorySummary, rootDirectory string, application *Application) string {
	status := singleLine(summary.Status)
	if summary.Failed {
		status = application.config.Locale.Text("failed")
	} else if status == "" {
		status = application.config.Locale.Text("unknown_status")
	}
	parts := []string{summary.CreatedAt.Local().Format("2006-01-02 15:04:05"), status}
	if summary.Duration > 0 {
		parts = append(parts, historyDuration(summary.Duration))
	}
	if summary.SourceFilePath != "" {
		parts = append(parts, displayCapturedSource(rootDirectory, summary.SourceFilePath, application.config.Locale))
	}
	return strings.Join(parts, " · ")
}

func historyDuration(duration time.Duration) string {
	return duration.String()
}

func (application *Application) selectHistory(index int) {
	if !application.Producer.SelectHistory(index) {
		return
	}
	application.closeOverlay()
	application.Element.SetFocus(application.Producer.Element)
}

func (application *Application) clearHistory() {
	count := application.Producer.ClearHistory()
	application.refreshHistory()
	application.Footer.UpdateStatus(application.config.Locale.Format("history_cleared", count))
}
