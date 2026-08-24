package http

import "fmt"

type Diagnostic struct {
	Line    int
	Column  int
	Message string
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
