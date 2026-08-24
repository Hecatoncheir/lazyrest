package producer

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	"github.com/Hecatoncheir/lazyrest/runner"
)

type BodyViewMode uint8

const (
	BodyViewPretty BodyViewMode = iota
	BodyViewRaw
)

func (mode BodyViewMode) String() string {
	if mode == BodyViewRaw {
		return "Raw"
	}
	return "Pretty"
}

func (widget *Producer) toggleBodyView() {
	if widget.bodyViewMode == BodyViewPretty {
		widget.bodyViewMode = BodyViewRaw
	} else {
		widget.bodyViewMode = BodyViewPretty
	}
	if !widget.IsRunning() && len(widget.history) > 0 {
		entry := widget.history[widget.historyIndex]
		widget.setText(renderExecutionResultWithMode(entry.Suite, entry.Response, entry.Err, widget.bodyViewMode))
	}
	widget.updateTitle()
}

func formatResponseBody(response runner.Response, mode BodyViewMode) string {
	if mode == BodyViewRaw || strings.TrimSpace(response.Body) == "" {
		return response.Body
	}

	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	trimmed := strings.TrimSpace(response.Body)
	if strings.Contains(contentType, "json") || json.Valid([]byte(trimmed)) {
		var output bytes.Buffer
		if err := json.Indent(&output, []byte(trimmed), "", "  "); err == nil {
			return output.String()
		}
	}
	if strings.Contains(contentType, "xml") || strings.HasPrefix(trimmed, "<") {
		if formatted, ok := formatXML(trimmed); ok {
			return formatted
		}
	}
	return response.Body
}

func formatXML(body string) (string, bool) {
	decoder := xml.NewDecoder(strings.NewReader(body))
	var output bytes.Buffer
	encoder := xml.NewEncoder(&output)
	encoder.Indent("", "  ")
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", false
		}
		if err := encoder.EncodeToken(token); err != nil {
			return "", false
		}
	}
	if err := encoder.Flush(); err != nil {
		return "", false
	}
	return output.String(), true
}
