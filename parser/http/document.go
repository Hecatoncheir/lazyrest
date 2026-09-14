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
// documentParser accumulates what reading a document produces: the requests,
// the diagnostics, the variables declared along the way, and the name a comment
// left for the request that follows it.
type documentParser struct {
	options     ParseOptions
	variables   map[string]string
	suites      []HttpSuite
	diagnostics []Diagnostic
	pendingName string
}

func parseDocument(source string, options ParseOptions) ([]HttpSuite, []Diagnostic) {
	variables := maps.Clone(options.Variables)
	if variables == nil {
		variables = make(map[string]string)
	}
	parser := &documentParser{
		options:     options,
		variables:   variables,
		suites:      []HttpSuite{},
		diagnostics: []Diagnostic{},
	}
	for _, current := range splitDocument(source) {
		parser.readBlock(current)
	}
	return parser.suites, parser.diagnostics
}

func (parser *documentParser) readBlock(current block) {
	switch current.kind {
	case blockVariable:
		parser.readVariable(current)
	case blockComment:
		parser.readComment(current)
	case blockRequest:
		parser.readRequest(current)
	}
}

func (parser *documentParser) readVariable(current block) {
	name, value, ok := parseVariableDeclaration(current.text)
	if !ok {
		parser.diagnostics = append(parser.diagnostics,
			diagnosticAt(current.line, "invalid variable declaration"))
		return
	}
	parser.variables[name] = value
}

// readComment keeps a name for the request that follows it, which is how a
// request is named at all.
func (parser *documentParser) readComment(current block) {
	if name := getNameFromComment(current.text); name != "" {
		parser.pendingName = name
	}
}

func (parser *documentParser) readRequest(current block) {
	suite := newSuiteFromText(current.text)
	if !suite.isRecognizedRequest() {
		parser.diagnostics = append(parser.diagnostics,
			diagnosticAt(current.line, "unrecognized content"))
		return
	}
	if parser.pendingName != "" {
		suite.Name = parser.pendingName
		parser.pendingName = ""
	}

	// The body is loaded before substitution so that the file can use variables
	// too.
	if err := loadExternalBody(&suite, parser.options.baseDirectory); err != nil {
		parser.report(&suite, current.line, err.Error())
	}
	applyGraphQL(&suite)
	parser.substitute(&suite, current.line)

	if suite.Name == "" {
		suite.Name = strings.TrimSpace(suite.Method + " " + suite.Uri)
	}
	parser.suites = append(parser.suites, suite)
}

// substitute fills in the variables the request uses and reports the ones it
// cannot.
func (parser *documentParser) substitute(suite *HttpSuite, line int) {
	resolution := resolveSuiteVariables(suite, parser.variables)
	suite.SecretValues = resolveSecretVariables(parser.options.SecretVariables, parser.variables)

	for _, name := range resolution.Missing {
		parser.report(suite, line, "undefined variable: "+name)
	}
	for _, cycle := range resolution.Cycles {
		parser.report(suite, line, "cyclic variable reference: "+cycle)
	}
}

// report records a blocking diagnostic twice: once for the document, and once
// on the request it belongs to, so a pane showing a single request still sees
// the reason it cannot run.
func (parser *documentParser) report(suite *HttpSuite, line int, message string) {
	diagnostic := blockingDiagnosticAt(line, message)
	parser.diagnostics = append(parser.diagnostics, diagnostic)
	suite.Diagnostics = append(suite.Diagnostics, diagnostic)
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

func blockingDiagnosticAt(line int, message string) Diagnostic {
	return Diagnostic{Line: line, Column: 1, Message: message, Severity: DiagnosticError}
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
