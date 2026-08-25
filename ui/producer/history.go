package producer

import (
	"fmt"
	nethttp "net/http"
	"slices"
	"strings"
	"time"

	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"

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
	widget.history = append(widget.history, sanitizedHistoryEntry(suite, response, err, time.Now()))
	if len(widget.history) > maxHistoryEntries {
		widget.history = widget.history[len(widget.history)-maxHistoryEntries:]
	}
	widget.historyIndex = len(widget.history) - 1
	widget.historyVisible = false
	widget.updateTitle()
	widget.persistHistory()
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
	widget.setText(widget.renderEntry(entry))
	widget.updateTitle()
}

// renderEntry draws an entry with the settings the widget currently holds.
func (widget *Producer) renderEntry(entry HistoryEntry) string {
	return widget.renderResult(entry.Suite, entry.Response, entry.Err)
}

func (widget *Producer) renderResult(suite http.HttpSuite, response runner.Response, err error) string {
	return renderExecutionResultWithLocale(suite, response, err, widget.bodyViewMode, widget.locale, widget.syntax)
}

func (widget *Producer) setText(text string) {
	widget.currentText = text
	widget.Element.(*tview.TextView).SetText(text)
}

func renderExecutionResult(suite http.HttpSuite, response runner.Response, err error) string {
	return renderExecutionResultWithMode(suite, response, err, BodyViewPretty)
}

func renderExecutionResultWithMode(suite http.HttpSuite, response runner.Response, err error, mode BodyViewMode) string {
	return renderExecutionResultWithLocale(suite, response, err, mode, locale.English(), syntax.Palette{})
}

func renderExecutionResultWithLocale(suite http.HttpSuite, response runner.Response, err error, mode BodyViewMode, translator *locale.Translator, palette syntax.Palette) string {
	if err != nil {
		return "[red]" + translator.Text("response_error") + ":[-]\n" + tview.Escape(redactSecrets(err.Error(), suite.SecretValues))
	}

	var request strings.Builder
	request.WriteString("[yellow]" + translator.Text("request") + ":[-]\n")
	request.WriteString(tview.Escape(fmt.Sprintf("%s %s\n", suite.Method, redactSecrets(suite.Uri, suite.SecretValues))))

	request.WriteString(tview.Escape(renderHeaders(suite.Header, suite.SecretValues)))
	bodyLabel := "body"
	if suite.BodyType == http.BodyTypeGraphQL {
		bodyLabel = "query"
	}
	if suite.Body != "" {
		request.WriteString("\n[yellow]" + translator.Text(bodyLabel) + ":[-]\n")
		request.WriteString(syntax.Highlight(redactSecrets(suite.Body, suite.SecretValues), requestLanguage(suite, mode), palette))
	}
	if suite.GraphQLVariables != "" {
		request.WriteString("\n\n[yellow]" + translator.Text("variables") + ":[-]\n")
		request.WriteString(syntax.Highlight(redactSecrets(prettyJSON(suite.GraphQLVariables), suite.SecretValues), jsonLanguage(mode), palette))
	}

	responseColor := "white"
	switch {
	case len(response.GraphQLErrors) > 0:
		// GraphQL answers with 200 even when the operation failed.
		responseColor = "red"
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
		responseDetails.WriteString(translator.Text("protocol") + ": " + response.Protocol + "\n")
	}
	if len(response.GraphQLErrors) > 0 {
		responseDetails.WriteString("\n" + translator.Text("graphql_errors") + ":\n")
		for _, message := range response.GraphQLErrors {
			responseDetails.WriteString("- " + redactSecrets(message, suite.SecretValues) + "\n")
		}
	}
	if len(response.Header) > 0 {
		responseDetails.WriteString("\n" + translator.Text("headers") + ":\n")
		responseDetails.WriteString(renderHeaders(response.Header, suite.SecretValues))
	}
	formatted, language := formatResponseBody(response, mode)
	body := redactSecrets(formatted, suite.SecretValues)
	responseText := fmt.Sprintf("[%s]%s:[-]\n%s\n%s",
		responseColor,
		translator.Text("response"),
		tview.Escape(responseDetails.String()),
		syntax.Highlight(body, language, palette),
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
