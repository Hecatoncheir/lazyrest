package hurl

import "strings"

// entryMethods open an entry in a Hurl file. Hurl writes the method in upper
// case, which keeps a body line from being read as the start of an entry.
var entryMethods = map[string]struct{}{
	"GET": {}, "HEAD": {}, "POST": {}, "PUT": {}, "PATCH": {}, "DELETE": {},
	"CONNECT": {}, "OPTIONS": {}, "TRACE": {}, "LINK": {}, "UNLINK": {},
	"PURGE": {}, "LOCK": {}, "UNLOCK": {}, "PROPFIND": {}, "VIEW": {},
}

// entry is one exchange of a Hurl file: its request line, the text of the whole
// exchange, and the position Hurl counts it at.
type entry struct {
	Number int
	Name   string
	Method string
	Uri    string
	Text   string
}

// splitEntries divides a Hurl file into the exchanges it runs in order. An
// entry starts at an unindented request line and ends where the next one
// begins, so the assertions and captures that follow a request stay with it.
func splitEntries(source string) []entry {
	lines := strings.SplitAfter(source, "\n")
	entries := []entry{}
	pendingName := ""

	for index := 0; index < len(lines); {
		method, uri, isRequest := requestLine(lines[index])
		if !isRequest {
			if name := nameFromComment(lines[index]); name != "" {
				pendingName = name
			}
			index++
			continue
		}

		start := index
		for index++; index < len(lines); index++ {
			if _, _, next := requestLine(lines[index]); next {
				break
			}
		}
		entries = append(entries, entry{
			Number: len(entries) + 1,
			Name:   pendingName,
			Method: method,
			Uri:    uri,
			Text:   strings.TrimSpace(strings.Join(lines[start:index], "")),
		})
		pendingName = ""
	}
	return entries
}

// requestLine reports the method and target of an entry's opening line. The
// line must not be indented, because an indented one belongs to a body.
func requestLine(line string) (string, string, bool) {
	if line == "" || line[0] == ' ' || line[0] == '\t' {
		return "", "", false
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", false
	}
	if _, known := entryMethods[fields[0]]; !known {
		return "", "", false
	}
	return fields[0], fields[1], true
}

func nameFromComment(line string) string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return ""
	}
	trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
	if !strings.HasPrefix(trimmed, "@name ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, "@name "))
}
