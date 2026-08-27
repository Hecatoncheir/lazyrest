package http

import (
	"encoding/json"
	nethttp "net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// IsSensitiveHeader reports whether a header normally carries authentication
// or session material and should never be rendered or persisted verbatim.
func IsSensitiveHeader(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	switch normalized {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key":
		return true
	}
	return sensitiveIdentifier(normalized) || strings.HasSuffix(normalized, "-auth")
}

func sensitiveResponseReference(kind, path string) bool {
	if kind == "headers" {
		return IsSensitiveHeader(path)
	}
	path = strings.TrimSpace(path)
	if path == "" || path == "$" || path == "*" {
		// A whole response body can contain credentials under keys that are no
		// longer visible after substitution, so handle it conservatively.
		return true
	}
	return sensitiveIdentifier(path)
}

func sensitiveIdentifier(value string) bool {
	normalized := strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			return unicode.ToLower(character)
		}
		return -1
	}, value)
	for _, marker := range []string{
		"token", "secret", "password", "passwd", "apikey", "session",
		"cookie", "credential", "privatekey", "clientkey", "bearer",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

// SensitiveResponseValues extracts credentials and session material from
// sensitive response headers and JSON members. The raw response remains
// available for chaining in memory; these values are used only for redaction.
func SensitiveResponseValues(body string, header nethttp.Header) []string {
	values := map[string]struct{}{}
	for name, headerValues := range header {
		if !IsSensitiveHeader(name) {
			continue
		}
		for _, value := range headerValues {
			addSensitiveValue(values, value)
		}
	}

	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err == nil {
		collectSensitiveJSONValues(values, document, false)
	}

	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func collectSensitiveJSONValues(output map[string]struct{}, value any, sensitive bool) {
	switch typed := value.(type) {
	case map[string]any:
		for name, child := range typed {
			collectSensitiveJSONValues(output, child, sensitive || sensitiveIdentifier(name))
		}
	case []any:
		for _, child := range typed {
			collectSensitiveJSONValues(output, child, sensitive)
		}
	case string:
		if sensitive {
			addSensitiveValue(output, typed)
		}
	case json.Number:
		if sensitive {
			addSensitiveValue(output, typed.String())
		}
	case bool:
		if sensitive {
			addSensitiveValue(output, strconv.FormatBool(typed))
		}
	}
}

func addSensitiveValue(output map[string]struct{}, value string) {
	if value = strings.TrimSpace(value); value != "" {
		output[value] = struct{}{}
	}
}
