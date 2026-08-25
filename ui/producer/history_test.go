package producer

import (
	"context"
	"errors"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	nethttp "net/http"
	"strings"
	"testing"
)

func TestStartRunCancelsPreviousRun(t *testing.T) {
	widget := New()
	firstContext, firstID := widget.StartRun()
	secondContext, secondID := widget.StartRun()

	select {
	case <-firstContext.Done():
	default:
		t.Fatal("previous run was not cancelled")
	}
	if widget.IsCurrentRun(firstID) || !widget.IsCurrentRun(secondID) {
		t.Fatal("current run identifier was not updated")
	}

	widget.CancelActive()
	if !errors.Is(secondContext.Err(), context.Canceled) {
		t.Fatalf("active run was not cancelled: %v", secondContext.Err())
	}
}

func TestRenderExecutionResult_IncludesResponseMetadataAndRedactsSecrets(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Method:       "GET",
		Uri:          "https://example.com?token=secret-value",
		SecretValues: []string{"secret-value"},
	}
	response := runner.Response{
		Code:     "200 OK",
		Protocol: "HTTP/2.0",
		Header: nethttp.Header{
			"Content-Type": []string{"application/json"},
			"Set-Cookie":   []string{"session=secret-value"},
		},
		Body: `{"token":"secret-value","ok":true}`,
	}

	text := renderExecutionResult(suite, response, nil)
	if !strings.Contains(text, "HTTP/2.0") || !strings.Contains(text, "Content-Type: application/json") {
		t.Fatalf("response metadata is missing: %q", text)
	}
	if strings.Contains(text, "secret-value") {
		t.Fatalf("secret value was rendered: %q", text)
	}
	if !strings.Contains(text, "Set-Cookie: <redacted>") {
		t.Fatalf("sensitive response header was not redacted: %q", text)
	}
}

func TestFormatResponseBody_PrettyAndRawJSON(t *testing.T) {
	response := runner.Response{
		Header: nethttp.Header{"Content-Type": []string{"application/json"}},
		Body:   `{"ok":true,"items":[1,2]}`,
	}

	pretty := formatResponseBody(response, BodyViewPretty)
	if !strings.Contains(pretty, "\n  \"items\"") {
		t.Fatalf("JSON was not formatted: %q", pretty)
	}
	if raw := formatResponseBody(response, BodyViewRaw); raw != response.Body {
		t.Fatalf("raw body changed: %q", raw)
	}
}

func TestRenderExecutionResult_RedactsSensitiveHeaders(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Method: "GET",
		Uri:    "https://example.com",
		Header: nethttp.Header{
			"Authorization": []string{"Bearer secret"},
			"Accept":        []string{"application/json"},
		},
	}
	text := renderExecutionResult(suite, runner.Response{Code: "200 OK"}, nil)

	if strings.Contains(text, "Bearer secret") {
		t.Fatal("sensitive header was rendered")
	}
	if !strings.Contains(text, "Authorization: <redacted>") {
		t.Fatalf("redaction marker is missing: %q", text)
	}
	if !strings.Contains(text, "Accept: application/json") {
		t.Fatalf("regular header is missing: %q", text)
	}
}

func TestRenderExecutionResult_ShowsRepeatedRequestHeaders(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Method: "GET",
		Uri:    "https://example.com",
		Header: nethttp.Header{"Accept": []string{"application/json", "text/html"}},
	}

	text := renderExecutionResult(suite, runner.Response{Code: "200 OK"}, nil)
	if !strings.Contains(text, "Accept: application/json, text/html") {
		t.Fatalf("repeated request header is not shown: %q", text)
	}
}

func TestRenderExecutionResult_ShowsGraphQLQueryVariablesAndErrors(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Method:           "POST",
		Uri:              "https://example.com/graphql",
		Body:             "query GetUser($id: ID!) {\n  user(id: $id) { name }\n}",
		BodyType:         parserhttp.BodyTypeGraphQL,
		GraphQLVariables: `{"id":"42"}`,
		Header:           nethttp.Header{},
	}
	response := runner.Response{
		Code:          "200 OK",
		StatusCode:    200,
		Body:          `{"data":null,"errors":[{"message":"Unknown field"}]}`,
		GraphQLErrors: []string{"Unknown field"},
	}

	text := renderExecutionResult(suite, response, nil)
	if !strings.Contains(text, "Query:") {
		t.Errorf("the query is not labelled: %q", text)
	}
	if !strings.Contains(text, "Variables:") || !strings.Contains(text, "\"id\": \"42\"") {
		t.Errorf("variables are not shown separately: %q", text)
	}
	if !strings.Contains(text, "GraphQL errors:") || !strings.Contains(text, "- Unknown field") {
		t.Errorf("GraphQL errors are not shown: %q", text)
	}
	if !strings.Contains(text, "[red]") {
		t.Errorf("a failed GraphQL response is not marked as failed: %q", text)
	}
}
