package producer

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"
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
	entry, available := widget.currentHistoryEntry()
	if !widget.IsRunning() && available {
		widget.historyDataMutex.RLock()
		resultAvailable := widget.resultAvailable
		widget.historyDataMutex.RUnlock()
		if resultAvailable {
			widget.setText(widget.renderEntry(entry))
		}
	}
	widget.updateTitle()
}

// formatResponseBody returns the body as it should be displayed together with
// the scanner that fits it. Raw is left untouched and unhighlighted, which is
// what makes it the way to see exactly what came over the wire.
func formatResponseBody(response runner.Response, mode BodyViewMode) (string, syntax.Language) {
	if mode == BodyViewRaw || strings.TrimSpace(response.Body) == "" {
		return response.Body, syntax.LanguagePlain
	}

	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	trimmed := strings.TrimSpace(response.Body)
	if strings.Contains(contentType, "json") || json.Valid([]byte(trimmed)) {
		var output bytes.Buffer
		if err := json.Indent(&output, []byte(trimmed), "", "  "); err == nil {
			return output.String(), syntax.LanguageJSON
		}
	}
	if strings.Contains(contentType, "xml") || strings.HasPrefix(trimmed, "<") {
		if formatted, ok := formatXML(trimmed); ok {
			return formatted, syntax.LanguageXML
		}
	}
	return response.Body, syntax.LanguagePlain
}

// requestLanguage picks the scanner for the body a request sends.
func requestLanguage(suite parserhttp.HttpSuite, mode BodyViewMode) syntax.Language {
	if mode == BodyViewRaw {
		return syntax.LanguagePlain
	}
	return syntax.LanguageForBodyType(suite.BodyType)
}

func jsonLanguage(mode BodyViewMode) syntax.Language {
	if mode == BodyViewRaw {
		return syntax.LanguagePlain
	}
	return syntax.LanguageJSON
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

// prettyJSON formats a JSON document for display, leaving it untouched when it
// cannot be parsed.
func prettyJSON(text string) string {
	var output bytes.Buffer
	if err := json.Indent(&output, []byte(text), "", "  "); err != nil {
		return text
	}
	return output.String()
}
