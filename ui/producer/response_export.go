package producer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Hecatoncheir/lazyrest/runner"
)

// ResponseExport is a secret-redacted snapshot of the response currently
// visible in Producer. Body follows the Pretty/Raw mode, while RawBody keeps
// the unformatted body for saving to disk.
type ResponseExport struct {
	Body                  string
	RawBody               string
	Full                  string
	SuggestedFileName     string
	SuggestedFullFileName string
	Truncated             bool
}

// CurrentResponse returns the response represented by the pane. It deliberately
// reads the sanitized in-memory history entry rather than the rendered TextView,
// which contains the request, labels, and tview colour markup.
func (widget *Producer) CurrentResponse() (ResponseExport, bool) {
	if widget == nil || widget.IsRunning() {
		return ResponseExport{}, false
	}
	widget.historyDataMutex.RLock()
	if !widget.resultAvailable || widget.historyIndex < 0 || widget.historyIndex >= len(widget.history) {
		widget.historyDataMutex.RUnlock()
		return ResponseExport{}, false
	}
	entry := widget.history[widget.historyIndex]
	widget.historyDataMutex.RUnlock()
	if entry.Err != nil || (entry.Response.Code == "" && entry.Response.StatusCode == 0) {
		return ResponseExport{}, false
	}
	body, _ := formatResponseBody(entry.Response, widget.bodyViewMode)
	return ResponseExport{
		Body:                  body,
		RawBody:               entry.Response.Body,
		Full:                  fullResponseText(entry.Response, body),
		SuggestedFileName:     suggestedResponseFileName(entry),
		SuggestedFullFileName: suggestedFullResponseFileName(entry),
		Truncated:             entry.Response.Truncated,
	}, true
}

func suggestedFullResponseFileName(entry HistoryEntry) string {
	bodyName := suggestedResponseFileName(entry)
	return strings.TrimSuffix(bodyName, filepath.Ext(bodyName)) + "-response.txt"
}

func fullResponseText(response runner.Response, body string) string {
	var output strings.Builder
	status := strings.TrimSpace(response.Protocol + " " + response.Code)
	if status == "" && response.StatusCode != 0 {
		status = fmt.Sprint(response.StatusCode)
	}
	output.WriteString(status)
	output.WriteByte('\n')
	if len(response.Header) > 0 {
		output.WriteString(renderHeaders(response.Header, nil))
	}
	output.WriteByte('\n')
	output.WriteString(body)
	return output.String()
}

func suggestedResponseFileName(entry HistoryEntry) string {
	stem := responseFileStem(entry.Suite.Name)
	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return fmt.Sprintf("%s-%s%s", stem, createdAt.Format("20060102-150405"), responseFileExtension(entry.Response))
}

func responseFileStem(name string) string {
	var stem strings.Builder
	dash := false
	for _, character := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if dash && stem.Len() > 0 {
				stem.WriteByte('-')
			}
			stem.WriteRune(character)
			dash = false
			continue
		}
		dash = true
	}
	if stem.Len() == 0 {
		return "response"
	}
	return stem.String()
}

func responseFileExtension(response runner.Response) string {
	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	switch {
	case strings.Contains(contentType, "json"):
		return ".json"
	case strings.Contains(contentType, "xml"):
		return ".xml"
	case strings.Contains(contentType, "html"):
		return ".html"
	case strings.HasPrefix(contentType, "text/"):
		return ".txt"
	}
	trimmed := strings.TrimSpace(response.Body)
	switch {
	case json.Valid([]byte(trimmed)):
		return ".json"
	case strings.HasPrefix(trimmed, "<"):
		return ".xml"
	default:
		return ".body"
	}
}
