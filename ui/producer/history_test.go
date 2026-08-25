package producer

import (
	"context"
	"errors"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	nethttp "net/http"

	"github.com/Hecatoncheir/lazyrest/ui/syntax"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
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

	pretty, language := formatResponseBody(response, BodyViewPretty)
	if !strings.Contains(pretty, "\n  \"items\"") {
		t.Fatalf("JSON was not formatted: %q", pretty)
	}
	if language != syntax.LanguageJSON {
		t.Fatalf("unexpected language: %v", language)
	}
	raw, rawLanguage := formatResponseBody(response, BodyViewRaw)
	if raw != response.Body {
		t.Fatalf("raw body changed: %q", raw)
	}
	if rawLanguage != syntax.LanguagePlain {
		t.Fatalf("raw body was marked for highlighting: %v", rawLanguage)
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

func TestRenderResult_UsesTheThemePalette(t *testing.T) {
	widget := New()
	widget.Build(Parameters{Theme: theme.NewDefault(), App: tview.NewApplication()})

	suite := parserhttp.HttpSuite{
		Method: "GET",
		Uri:    "https://example.com/users",
		Header: nethttp.Header{},
	}
	response := runner.Response{
		Code:       "200 OK",
		StatusCode: 200,
		Header:     nethttp.Header{"Content-Type": []string{"application/json"}},
		Body:       `{"name":"Ada","age":36}`,
	}

	text := widget.renderResult(suite, response, nil)
	// The default theme is gruvbox: keys take its accent, strings its success
	// colour, numbers its progress colour.
	for _, want := range []string{`[#83a598]"name"`, `[#b8bb26]"Ada"`, `[#fabd2f]36`} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %s in %q", want, text)
		}
	}

	widget.bodyViewMode = BodyViewRaw
	if raw := widget.renderResult(suite, response, nil); strings.Contains(raw, "[#83a598]\"name\"") {
		t.Errorf("the raw view was highlighted: %q", raw)
	}
}
