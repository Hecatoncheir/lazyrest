package http

import (
	"maps"
	"strings"
)

// requestMethods are the methods that mark a line as the start of a new request
// in the middle of a document. A request introduced by a separator, a comment,
// or the start of the file is not checked against this list, so an unusual
// method still works there.
var requestMethods = map[string]struct{}{
	"GET": {}, "HEAD": {}, "POST": {}, "PUT": {}, "PATCH": {}, "DELETE": {},
	"CONNECT": {}, "OPTIONS": {}, "TRACE": {}, "GRAPHQL": {},
}

type blockKind uint8

const (
	blockComment blockKind = iota
	blockVariable
	blockRequest
)

// block is one top level piece of a document: a comment, a variable
// declaration, or the whole text of a request.
type block struct {
	kind blockKind
	text string
	line int
}

// splitDocument divides a file into the blocks it is made of. A request runs
// until the next separator, the comment that names the following request, or a
// blank line followed by something that plainly starts a new block, so that a
// body keeps whatever it holds.
func splitDocument(source string) []block {
	lines := strings.SplitAfter(source, "\n")
	blocks := []block{}

	for index := 0; index < len(lines); {
		trimmed := strings.TrimSpace(lines[index])
		switch {
		case trimmed == "":
			index++
		case isCommentLine(trimmed):
			blocks = append(blocks, block{kind: blockComment, text: trimmed, line: index + 1})
			index++
		case strings.HasPrefix(trimmed, "@"):
			blocks = append(blocks, block{kind: blockVariable, text: trimmed, line: index + 1})
			index++
		default:
			start := index
			for index++; index < len(lines) && !startsNewBlock(lines, index); index++ {
			}
			blocks = append(blocks, block{
				kind: blockRequest,
				text: strings.Join(lines[start:index], ""),
				line: start + 1,
			})
		}
	}
	return blocks
}

func startsNewBlock(lines []string, index int) bool {
	trimmed := strings.TrimSpace(lines[index])
	if isSeparatorLine(trimmed) {
		return true
	}
	if isCommentLine(trimmed) && getNameFromComment(trimmed) != "" {
		return true
	}
	// Everything else ends a request only after a blank line. Inside a body a
	// line that looks like a request is far more likely to be content.
	if strings.TrimSpace(lines[index-1]) != "" {
		return false
	}
	return looksLikeRequestLine(trimmed) || isVariableDeclaration(trimmed)
}

func isSeparatorLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "###")
}

func isCommentLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//")
}

func isVariableDeclaration(trimmed string) bool {
	if !strings.HasPrefix(trimmed, "@") {
		return false
	}
	_, _, ok := parseVariableDeclaration(trimmed)
	return ok
}

func looksLikeRequestLine(trimmed string) bool {
	fields := strings.Fields(trimmed)
	if len(fields) < 2 {
		return false
	}
	if _, known := requestMethods[fields[0]]; !known {
		return false
	}
	return looksLikeTarget(fields[1])
}

func looksLikeTarget(target string) bool {
	return strings.Contains(target, "://") ||
		strings.HasPrefix(target, "/") ||
		strings.HasPrefix(target, "{{")
}

// parseDocument reads every request of a file, together with the diagnostics
// that describe what could not be read.
func parseDocument(source string, options ParseOptions) ([]HttpSuite, []Diagnostic) {
	suites := []HttpSuite{}
	diagnostics := []Diagnostic{}
	pendingName := ""
	variables := maps.Clone(options.Variables)
	if variables == nil {
		variables = make(map[string]string)
	}

	for _, current := range splitDocument(source) {
		switch current.kind {
		case blockVariable:
			name, value, ok := parseVariableDeclaration(current.text)
			if !ok {
				diagnostics = append(diagnostics, diagnosticAt(current.line, "invalid variable declaration"))
				continue
			}
			variables[name] = value

		case blockComment:
			if name := getNameFromComment(current.text); name != "" {
				pendingName = name
			}

		case blockRequest:
			suite := newSuiteFromText(current.text)
			if !suite.isRecognizedRequest() {
				diagnostics = append(diagnostics, diagnosticAt(current.line, "unrecognized content"))
				continue
			}
			if pendingName != "" {
				suite.Name = pendingName
				pendingName = ""
			}
			// The body is loaded before substitution so that the file can use
			// variables too.
			if err := loadExternalBody(&suite, options.baseDirectory); err != nil {
				diagnostics = append(diagnostics, diagnosticAt(current.line, err.Error()))
			}
			applyGraphQL(&suite)
			resolution := resolveSuiteVariables(&suite, variables)
			suite.SecretValues = resolveSecretVariables(options.SecretVariables, variables)
			for _, name := range resolution.Missing {
				diagnostics = append(diagnostics, diagnosticAt(current.line, "undefined variable: "+name))
			}
			for _, cycle := range resolution.Cycles {
				diagnostics = append(diagnostics, diagnosticAt(current.line, "cyclic variable reference: "+cycle))
			}
			if suite.Name == "" {
				suite.Name = strings.TrimSpace(suite.Method + " " + suite.Uri)
			}
			suites = append(suites, suite)
		}
	}

	return suites, diagnostics
}

func newSuiteFromText(text string) HttpSuite {
	parsed := parseRequestText(text)

	suite := NewHttpSuite()
	suite.Method = parsed.Method
	suite.Uri = parsed.Uri
	suite.Header = parsed.Header
	suite.Body = parsed.Body
	suite.BodyType = detectBodyType(suite)
	return suite
}

// isRecognizedRequest reports whether a block really is a request, so that prose
// in a file is reported rather than executed.
func (suite HttpSuite) isRecognizedRequest() bool {
	if suite.Method == "" || suite.Uri == "" {
		return false
	}
	_, known := requestMethods[suite.Method]
	return known || looksLikeTarget(suite.Uri)
}

func diagnosticAt(line int, message string) Diagnostic {
	return Diagnostic{Line: line, Column: 1, Message: message}
}

func getNameFromComment(comment string) string {
	comment = strings.TrimSpace(comment)
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimSpace(strings.TrimLeft(comment, "#"))
	if strings.HasPrefix(comment, "@name ") {
		return strings.TrimSpace(strings.TrimPrefix(comment, "@name "))
	}
	for _, prefix := range []string{"@suite(", "@test("} {
		if !strings.HasPrefix(comment, prefix) || !strings.HasSuffix(comment, ")") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(comment, prefix), ")")
		return strings.Trim(strings.TrimSpace(name), "\"'")
	}
	return ""
}
