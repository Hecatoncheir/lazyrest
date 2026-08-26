package http

import (
	nethttp "net/http"
	"regexp"
	"strings"
)

// externalBodyPattern matches a body that names a file to send, written as
// `< ./payload.json`, optionally with the JetBrains encoding marker.
var externalBodyPattern = regexp.MustCompile(`^<@?(?:[A-Za-z0-9_-]+)?[ \t]+(\S.*)$`)

// headerLinePattern matches the "Name:" prefix of a header line. The name is an
// RFC 7230 token, optionally assembled from {{variable}} placeholders, which
// keeps body lines such as `{"key": "value"}` from being read as headers.
var headerLinePattern = regexp.MustCompile("^(?:\\{\\{[^{}]*\\}\\}|[!#$%&'*+\\-.^_`|~0-9A-Za-z]+)+[ \t]*:")

// requestText is the result of reading the raw text of a single request.
type requestText struct {
	Method string
	Uri    string
	Header nethttp.Header
	Body   string
}

// parseRequestText reads a request the way an HTTP message is structured: a
// request line, header lines, then the body after the first blank line.
func parseRequestText(text string) requestText {
	parsed := requestText{Header: nethttp.Header{}}
	lines := strings.SplitAfter(text, "\n")

	index := 0
	for ; index < len(lines); index++ {
		fields := strings.Fields(lines[index])
		if len(fields) == 0 {
			continue
		}
		if len(fields) >= 2 {
			parsed.Method = fields[0]
			parsed.Uri = fields[1]
		} else {
			parsed.Uri = fields[0]
		}
		index++
		break
	}

	previousName := ""
	for ; index < len(lines); index++ {
		line := lines[index]
		if strings.TrimSpace(line) == "" {
			index++
			break
		}
		if previousName != "" && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			appendFoldedHeaderLine(parsed.Header, previousName, strings.TrimSpace(line))
			continue
		}
		name, value, ok := splitHeaderLine(line)
		if !ok {
			// A body that is not separated by a blank line starts here.
			break
		}
		parsed.Header.Add(name, value)
		previousName = name
	}

	parsed.Body = strings.TrimSpace(strings.Join(lines[index:], ""))
	return parsed
}

func splitHeaderLine(line string) (string, string, bool) {
	line = strings.TrimRight(line, "\r\n")
	if !headerLinePattern.MatchString(line) {
		return "", "", false
	}
	name, value, _ := strings.Cut(line, ":")
	return strings.TrimSpace(name), strings.TrimSpace(value), true
}

// appendFoldedHeaderLine continues the last value of name with an obsolete
// folded continuation line.
func appendFoldedHeaderLine(header nethttp.Header, name, continuation string) {
	values := header.Values(name)
	if len(values) == 0 {
		return
	}
	values[len(values)-1] = strings.TrimSpace(values[len(values)-1] + " " + continuation)
}

// externalBodyPath reports the file a body refers to, if the body is nothing
// but such a reference.
func externalBodyPath(body string) (string, bool) {
	body = strings.TrimSpace(body)
	if strings.Contains(body, "\n") {
		return "", false
	}
	match := externalBodyPattern.FindStringSubmatch(body)
	if match == nil {
		return "", false
	}
	return strings.TrimSpace(match[1]), true
}
