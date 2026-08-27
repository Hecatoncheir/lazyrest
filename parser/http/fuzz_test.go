package http

import (
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func FuzzHTTPDocumentSyntax(f *testing.F) {
	for _, source := range []string{
		"GET https://example.test\n",
		"# @name login\nPOST https://example.test/login\nContent-Type: application/json\n\n{\"name\":\"demo\"}\n",
		"@host = example.test\n### first\nGET https://{{host}}/one\n\n### second\nPATCH /two\nX-Trace: one\n two\n\nbody\n",
		"GRAPHQL https://example.test/graphql\n\nquery Viewer { viewer { id } }\n\n{\"limit\": 10}\n",
		"not a request\r\n\r\nDELETE https://example.test/items/1\r\n",
	} {
		f.Add(source)
	}

	f.Fuzz(func(t *testing.T, source string) {
		first := splitDocument(source)
		second := splitDocument(source)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("document splitting is not deterministic")
		}
		previousLine := 0
		for _, current := range first {
			if current.line < 1 || current.line < previousLine {
				t.Fatalf("invalid block line sequence: previous=%d current=%d", previousLine, current.line)
			}
			previousLine = current.line
			if current.kind != blockRequest {
				continue
			}
			parsed := parseRequestText(current.text)
			if parsed.Header == nil {
				t.Fatal("request parser returned a nil header map")
			}
			for name := range parsed.Header {
				if strings.ContainsAny(name, "\r\n") {
					t.Fatalf("parsed header name contains a line break: %q", name)
				}
			}
		}
	})
}

func FuzzResolveVariables(f *testing.F) {
	for _, seed := range [][3]string{
		{"example.test", "v1", "https://{{a}}/{{b}}"},
		{"{{b}}", "{{a}}", "{{a}}"},
		{"prefix-{{missing}}", "value", "{{a}}/{{missing}}"},
		{"$value", "line one\nline two", "{{b}}"},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}

	f.Fuzz(func(t *testing.T, firstValue, secondValue, text string) {
		resolve := func() (HttpSuite, VariableResolution) {
			suite := HttpSuite{
				Uri:              text,
				Body:             text,
				GraphQLVariables: text,
				Header:           http.Header{"X-Fuzz": {text}},
			}
			result := resolveSuiteVariables(&suite, map[string]string{
				"a": firstValue,
				"b": secondValue,
			})
			return suite, result
		}

		firstSuite, firstResult := resolve()
		secondSuite, secondResult := resolve()
		if !reflect.DeepEqual(firstSuite, secondSuite) || !reflect.DeepEqual(firstResult, secondResult) {
			t.Fatal("variable resolution is not deterministic")
		}
		if !slices.IsSorted(firstResult.Missing) || !slices.IsSorted(firstResult.Cycles) {
			t.Fatalf("variable diagnostics are not sorted: %+v", firstResult)
		}
	})
}

func FuzzResolveResponseReferences(f *testing.F) {
	for _, seed := range [][2]string{
		{`{"token":"abc123"}`, "$.token"},
		{`{"data":{"items":[{"id":42}]}}`, "$.data.items[0].id"},
		{`[true, false]`, "$[1]"},
		{"not-json", "$.token"},
		{`{"nested":{"value":"ok"}}`, "$.missing"},
	} {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, body, path string) {
		store := &ResponseStore{}
		store.Record(HttpSuite{Name: "login", SourceFilePath: "fuzz.http"}, ResponseValue{
			Body:   body,
			Header: http.Header{"X-Session-Token": {"session-value"}},
			Status: "200 OK",
		})
		reference := "{{login.response.body." + path + "}}"
		resolve := func() (HttpSuite, []string) {
			suite := HttpSuite{
				Uri:            "https://example.test/" + reference,
				Body:           reference,
				Header:         http.Header{"Authorization": {"Bearer " + reference}},
				SourceFilePath: "fuzz.http",
			}
			unresolved := ResolveResponseReferences(&suite, store)
			return suite, unresolved
		}

		firstSuite, firstUnresolved := resolve()
		secondSuite, secondUnresolved := resolve()
		if !reflect.DeepEqual(firstSuite, secondSuite) || !reflect.DeepEqual(firstUnresolved, secondUnresolved) {
			t.Fatal("response reference resolution is not deterministic")
		}
		if !slices.IsSorted(firstSuite.SecretValues) {
			t.Fatalf("secret values are not sorted: %#v", firstSuite.SecretValues)
		}
	})
}

func FuzzSecretRedaction(f *testing.F) {
	for _, seed := range [][3]string{
		{"Bearer abc123", "abc123", ""},
		{"token=short-and-long", "short", "short-and-long"},
		{"пароль: секрет", "секрет", "пароль"},
		{"nothing sensitive", "", "missing"},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}

	f.Fuzz(func(t *testing.T, value, firstSecret, secondSecret string) {
		secrets := []string{firstSecret, secondSecret}
		original := slices.Clone(secrets)
		first := RedactSecrets(value, secrets)
		second := RedactSecrets(value, secrets)
		if first != second {
			t.Fatal("secret redaction is not deterministic")
		}
		if !slices.Equal(secrets, original) {
			t.Fatalf("secret redaction mutated its input: got %#v, want %#v", secrets, original)
		}
	})
}
