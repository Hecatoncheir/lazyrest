package producer

import (
	"fmt"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"slices"
	"strings"
	"time"

	"github.com/rivo/tview"
)

const maxHistoryEntries = 50

type HistoryEntry struct {
	Suite     http.HttpSuite
	Response  runner.Response
	Err       error
	Rendered  string
	CreatedAt time.Time
}

func (widget *Producer) addHistory(suite http.HttpSuite, response runner.Response, err error, rendered string) {
	widget.history = append(widget.history, HistoryEntry{
		Suite:     suite,
		Response:  response,
		Err:       err,
		Rendered:  rendered,
		CreatedAt: time.Now(),
	})
	if len(widget.history) > maxHistoryEntries {
		widget.history = widget.history[len(widget.history)-maxHistoryEntries:]
	}
	widget.historyIndex = len(widget.history) - 1
}

func (widget *Producer) showHistory(delta int) {
	if len(widget.history) == 0 {
		return
	}
	widget.historyIndex += delta
	if widget.historyIndex < 0 {
		widget.historyIndex = 0
	}
	if widget.historyIndex >= len(widget.history) {
		widget.historyIndex = len(widget.history) - 1
	}
	entry := widget.history[widget.historyIndex]
	widget.setText(entry.Rendered)
	widget.Element.(*tview.TextView).SetTitle(fmt.Sprintf("Producer history %d/%d", widget.historyIndex+1, len(widget.history)))
}

func (widget *Producer) setText(text string) {
	widget.currentText = text
	widget.Element.(*tview.TextView).SetText(text)
}

func renderExecutionResult(suite http.HttpSuite, response runner.Response, err error) string {
	if err != nil {
		return "[red]Response error:[white]\n" + tview.Escape(err.Error())
	}

	var request strings.Builder
	request.WriteString("[yellow]Request:[white]\n")
	request.WriteString(tview.Escape(fmt.Sprintf("%s %s\n", suite.Method, suite.Uri)))

	headerKeys := make([]string, 0, len(suite.Header))
	for key := range suite.Header {
		headerKeys = append(headerKeys, key)
	}
	slices.Sort(headerKeys)
	for _, key := range headerKeys {
		value := suite.Header[key]
		if isSensitiveHeader(key) {
			value = "<redacted>"
		}
		request.WriteString(tview.Escape(fmt.Sprintf("%s: %s\n", key, value)))
	}
	if suite.Body != "" {
		request.WriteString("\n[yellow]Body:[white]\n")
		request.WriteString(tview.Escape(suite.Body))
	}

	responseColor := "white"
	switch {
	case strings.HasPrefix(response.Code, "2"):
		responseColor = "green"
	case strings.HasPrefix(response.Code, "3"):
		responseColor = "yellow"
	case strings.HasPrefix(response.Code, "4"), strings.HasPrefix(response.Code, "5"), response.Code == "FAILED":
		responseColor = "red"
	}

	separator := "\n" + strings.Repeat("─", 40) + "\n"
	responseText := fmt.Sprintf("[%s]Response:[white]\n%s\n\n%s",
		responseColor,
		tview.Escape(response.ToMiniString()),
		tview.Escape(response.Body),
	)
	return request.String() + separator + responseText
}

func isSensitiveHeader(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	switch normalized {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key":
		return true
	}
	return strings.Contains(normalized, "token") || strings.Contains(normalized, "secret")
}
