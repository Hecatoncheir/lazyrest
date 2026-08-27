package http

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"strings"
)

var ErrRequestNotRunnable = errors.New("request has blocking parser diagnostics")

type HttpSuite struct {
	Name     string
	Method   string
	Uri      string
	Header   nethttp.Header
	Body     string
	BodyType string
	IsHurl   bool
	// SourceFilePath scopes captured responses to the request file they came
	// from. It belongs to the current session and is not persisted in history.
	SourceFilePath string `json:"-"`
	HurlFilePath   string
	// HurlEntry is the position of an entry in its Hurl file, counted from
	// one. Hurl runs a file in order, so an entry is reached by running
	// everything up to it.
	HurlEntry int
	// Variables are passed to Hurl, which does its own substitution.
	Variables map[string]string
	// GraphQLVariables and GraphQLOperation complete a GraphQL request whose
	// query is held in Body.
	GraphQLVariables string
	GraphQLOperation string
	SecretValues     []string
	// Diagnostics belong to this request rather than the document as a whole.
	// Error-severity diagnostics prevent execution; warnings remain informational.
	Diagnostics []Diagnostic `json:"-"`
}

func NewHttpSuite() HttpSuite {
	return HttpSuite{
		Header: nethttp.Header{},
	}
}

func (suite HttpSuite) ValidateForExecution() error {
	messages := make([]string, 0, len(suite.Diagnostics))
	for _, diagnostic := range suite.Diagnostics {
		if diagnostic.IsBlocking() {
			messages = append(messages, diagnostic.String())
		}
	}
	if len(messages) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrRequestNotRunnable, strings.Join(messages, "; "))
}
