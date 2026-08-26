package producer

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
)

func TestBuildCurlCommandQuotesHeadersURLAndBody(t *testing.T) {
	command, err := BuildCurlCommand(parserhttp.HttpSuite{
		Method: http.MethodPost,
		Uri:    "https://example.test/users?name=O'Reilly",
		Header: http.Header{
			"X-Token": {"secret"},
			"Accept":  {"application/json", "text/plain"},
		},
		Body:     `{"name":"O'Reilly"}`,
		BodyType: parserhttp.BodyTypeJSON,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedInOrder := []string{
		"curl \\\n  --globoff",
		"--request 'POST'",
		`--url 'https://example.test/users?name=O'"'"'Reilly'`,
		"--header 'Accept: application/json'",
		"--header 'Accept: text/plain'",
		"--header 'Content-Type: application/json; charset=utf-8'",
		"--header 'X-Token: secret'",
		`--data-raw '{"name":"O'"'"'Reilly"}'`,
	}
	position := 0
	for _, expected := range expectedInOrder {
		index := strings.Index(command[position:], expected)
		if index < 0 {
			t.Fatalf("cURL command does not contain %q after position %d:\n%s", expected, position, command)
		}
		position += index + len(expected)
	}
}

func TestBuildCurlCommandUsesEncodedGraphQLPayload(t *testing.T) {
	command, err := BuildCurlCommand(parserhttp.HttpSuite{
		Method:           http.MethodPost,
		Uri:              "https://example.test/graphql",
		Header:           http.Header{},
		Body:             "query User { user { id } }",
		BodyType:         parserhttp.BodyTypeGraphQL,
		GraphQLVariables: `{"limit":2}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"--header 'Content-Type: application/json; charset=utf-8'",
		`--data-raw '{"query":"query User { user { id } }","variables":{"limit":2}}'`,
	} {
		if !strings.Contains(command, expected) {
			t.Errorf("GraphQL cURL command does not contain %q:\n%s", expected, command)
		}
	}
}

func TestBuildCurlCommandRejectsHurlWorkflow(t *testing.T) {
	_, err := BuildCurlCommand(parserhttp.HttpSuite{IsHurl: true})
	if !errors.Is(err, ErrHurlCurlUnsupported) {
		t.Fatalf("unexpected Hurl cURL error: %v", err)
	}
}

func TestCurrentRequestFollowsRuntimeHistorySelection(t *testing.T) {
	widget := &Producer{requestAvailable: true, suite: parserhttp.HttpSuite{Method: http.MethodGet, Uri: "https://example.test/latest", Header: http.Header{}}}
	older := parserhttp.HttpSuite{Method: http.MethodPost, Uri: "https://example.test/older", Header: http.Header{"Authorization": {"Bearer secret"}}}
	widget.history = []HistoryEntry{{request: &older}}
	widget.historyIndex = 0
	widget.historyVisible = true

	selected, ok := widget.CurrentRequest()
	if !ok || selected.Uri != older.Uri || selected.Header.Get("Authorization") != "Bearer secret" {
		t.Fatalf("selected runtime request was not returned: %+v, %v", selected, ok)
	}
	selected.Header.Set("Authorization", "changed")
	if older.Header.Get("Authorization") != "Bearer secret" {
		t.Fatal("CurrentRequest returned shared request headers")
	}

	widget.history[0].request = nil
	if _, ok := widget.CurrentRequest(); ok {
		t.Fatal("restored redacted history must not become a runnable request")
	}
}
