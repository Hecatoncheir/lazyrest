package producer

import (
	"strconv"
	"strings"
	"time"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
)

// HistorySummary contains the safe metadata needed by the project history
// window. Index identifies the entry in Producer's chronological history.
type HistorySummary struct {
	Index          int
	Name           string
	Method         string
	SourceFilePath string
	Status         string
	Duration       time.Duration
	CreatedAt      time.Time
	Failed         bool
}

// HistorySummaries returns the newest entries first without exposing request
// or response bodies and headers.
func (widget *Producer) HistorySummaries() []HistorySummary {
	if widget == nil {
		return nil
	}
	widget.historyDataMutex.RLock()
	defer widget.historyDataMutex.RUnlock()
	summaries := make([]HistorySummary, 0, len(widget.history))
	for index := len(widget.history) - 1; index >= 0; index-- {
		entry := widget.history[index]
		status := strings.TrimSpace(entry.Response.Code)
		if status == "" && entry.Response.StatusCode != 0 {
			status = strconv.Itoa(entry.Response.StatusCode)
		}
		summaries = append(summaries, HistorySummary{
			Index:          index,
			Name:           strings.TrimSpace(entry.Suite.Name),
			Method:         strings.TrimSpace(entry.Suite.Method),
			SourceFilePath: entry.Suite.SourceFilePath,
			Status:         status,
			Duration:       entry.Response.Time,
			CreatedAt:      entry.CreatedAt,
			Failed:         entry.Err != nil,
		})
	}
	return summaries
}

// SelectHistory displays one history entry in Producer.
func (widget *Producer) SelectHistory(index int) bool {
	if widget == nil || widget.Element == nil {
		return false
	}
	widget.historyDataMutex.Lock()
	if index < 0 || index >= len(widget.history) {
		widget.historyDataMutex.Unlock()
		return false
	}
	widget.historyIndex = index
	entry := widget.history[index]
	widget.historyVisible = true
	widget.resultAvailable = entry.Err == nil && (entry.Response.Code != "" || entry.Response.StatusCode != 0)
	widget.historyDataMutex.Unlock()
	widget.setText(widget.renderEntry(entry))
	widget.updateTitle()
	return true
}

// ClearHistory removes the in-memory and persisted history for this project.
// A request already running may add a new first entry when it completes.
func (widget *Producer) ClearHistory() int {
	if widget == nil {
		return 0
	}
	widget.historyDataMutex.Lock()
	count := len(widget.history)
	widget.history = nil
	widget.historyIndex = -1
	widget.historyVisible = false
	widget.resultAvailable = false
	widget.requestAvailable = false
	widget.suite = parserhttp.HttpSuite{}
	widget.searchMode = false
	widget.searchQuery = ""
	widget.searchMatches = nil
	widget.searchIndex = -1
	widget.historyDataMutex.Unlock()
	if widget.Element != nil && !widget.IsRunning() {
		widget.setText("")
		widget.updateTitle()
	}
	widget.persistHistory()
	return count
}
