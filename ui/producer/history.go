package producer

import (
	"fmt"
	nethttp "net/http"
	"slices"
	"strings"
	"time"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"

	"github.com/rivo/tview"
)

const maxHistoryEntries = 50

type HistoryEntry struct {
	Suite     http.HttpSuite
	Response  runner.Response
	Err       error
	CreatedAt time.Time
}

func (widget *Producer) addHistory(suite http.HttpSuite, response runner.Response, err error) {
	widget.history = append(widget.history, HistoryEntry{
		Suite:     suite,
		Response:  response,
		Err:       err,
		CreatedAt: time.Now(),
	})
	if len(widget.history) > maxHistoryEntries {
		widget.history = widget.history[len(widget.history)-maxHistoryEntries:]
	}
	widget.historyIndex = len(widget.history) - 1
	widget.historyVisible = false
	widget.updateTitle()
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
	widget.historyVisible = true
	widget.setText(renderExecutionResultWithMode(entry.Suite, entry.Response, entry.Err, widget.bodyViewMode))
	widget.updateTitle()
}

func (widget *Producer) setText(text string) {
	widget.currentText = text
	widget.Element.(*tview.TextView).SetText(text)
}

func renderExecutionResult(suite http.HttpSuite, response runner.Response, err error) string {
	return renderExecutionResultWithMode(suite, response, err, BodyViewPretty)
}

func renderExecutionResultWithMode(suite http.HttpSuite, response runner.Response, err error, mode BodyViewMode) string {
	if err != nil {
		return "[red]Response error:[white]\n" + tview.Escape(redactSecrets(err.Error(), suite.SecretValues))
	}

	var request strings.Builder
	request.WriteString("[yellow]Request:[white]\n")
	request.WriteString(tview.Escape(fmt.Sprintf("%s %s\n", suite.Method, redactSecrets(suite.Uri, suite.SecretValues))))

	headerKeys := make([]string, 0, len(suite.Header))
	for key := range suite.Header {
		headerKeys = append(headerKeys, key)
	}
	slices.Sort(headerKeys)
	for _, key := range headerKeys {
		displayKey := redactSecrets(key, suite.SecretValues)
		value := suite.Header[key]
		if isSensitiveHeader(key) {
			value = "<redacted>"
		} else {
			value = redactSecrets(value, suite.SecretValues)
		}
		request.WriteString(tview.Escape(fmt.Sprintf("%s: %s\n", displayKey, value)))
	}
	if suite.Body != "" {
		request.WriteString("\n[yellow]Body:[white]\n")
		request.WriteString(tview.Escape(redactSecrets(suite.Body, suite.SecretValues)))
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
	var responseDetails strings.Builder
	responseDetails.WriteString(response.ToMiniString())
	if response.Protocol != "" {
		responseDetails.WriteString("Protocol: " + response.Protocol + "\n")
	}
	if len(response.Header) > 0 {
		responseDetails.WriteString("\nHeaders:\n")
		responseDetails.WriteString(renderHeaders(response.Header, suite.SecretValues))
	}
	body := redactSecrets(formatResponseBody(response, mode), suite.SecretValues)
	responseText := fmt.Sprintf("[%s]Response:[white]\n%s\n%s",
		responseColor,
		tview.Escape(responseDetails.String()),
		tview.Escape(body),
	)
	return request.String() + separator + responseText
}

func renderHeaders(headers nethttp.Header, secretValues []string) string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	var output strings.Builder
	for _, key := range keys {
		displayKey := redactSecrets(key, secretValues)
		value := strings.Join(headers.Values(key), ", ")
		if isSensitiveHeader(key) {
			value = "<redacted>"
		} else {
			value = redactSecrets(value, secretValues)
		}
		output.WriteString(fmt.Sprintf("%s: %s\n", displayKey, value))
	}
	return output.String()
}

func redactSecrets(value string, secrets []string) string {
	return http.RedactSecrets(value, secrets)
}

func isSensitiveHeader(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	switch normalized {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key":
		return true
	}
	return strings.Contains(normalized, "token") || strings.Contains(normalized, "secret")
}
