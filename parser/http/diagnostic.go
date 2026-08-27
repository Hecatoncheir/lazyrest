package http

import "fmt"

type DiagnosticSeverity uint8

const (
	DiagnosticWarning DiagnosticSeverity = iota
	DiagnosticError
)

type Diagnostic struct {
	Line     int
	Column   int
	Message  string
	Severity DiagnosticSeverity
}

func (diagnostic Diagnostic) IsBlocking() bool {
	return diagnostic.Severity >= DiagnosticError
}

func (diagnostic Diagnostic) String() string {
	if diagnostic.Line <= 0 {
		return diagnostic.Message
	}
	return fmt.Sprintf("line %d, column %d: %s", diagnostic.Line, diagnostic.Column, diagnostic.Message)
}

type ParseResult struct {
	Suites      []HttpSuite
	Diagnostics []Diagnostic
}
