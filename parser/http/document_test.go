package http

import (
	nethttp "net/http"
	"testing"
)

func TestSplitDocument(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   []block
	}{
		{
			name:   "a separator ends a request",
			source: "GET https://example.com/a\n\n### next\nGET https://example.com/b\n",
			want: []block{
				{kind: blockRequest, text: "GET https://example.com/a\n\n", line: 1},
				{kind: blockComment, text: "### next", line: 3},
				{kind: blockRequest, text: "GET https://example.com/b\n", line: 4},
			},
		},
		{
			name:   "a naming comment ends a request",
			source: "GET https://example.com/a\n# @name second\nGET https://example.com/b\n",
			want: []block{
				{kind: blockRequest, text: "GET https://example.com/a\n", line: 1},
				{kind: blockComment, text: "# @name second", line: 2},
				{kind: blockRequest, text: "GET https://example.com/b\n", line: 3},
			},
		},
		{
			name:   "a blank line and a request line end a body",
			source: "POST https://example.com/a\n{\n\"k\": 1\n}\n\nGET https://example.com/b\n",
			want: []block{
				{kind: blockRequest, text: "POST https://example.com/a\n{\n\"k\": 1\n}\n\n", line: 1},
				{kind: blockRequest, text: "GET https://example.com/b\n", line: 6},
			},
		},
		{
			name:   "a plain comment stays inside a body",
			source: "POST https://example.com/a\n\nquery {\n  # keep me\n  id\n}\n",
			want: []block{
				{kind: blockRequest, text: "POST https://example.com/a\n\nquery {\n  # keep me\n  id\n}\n", line: 1},
			},
		},
		{
			name:   "a request line inside a body is content",
			source: "POST https://example.com/a\n\nfirst line\nGET https://example.com/b\n",
			want: []block{
				{kind: blockRequest, text: "POST https://example.com/a\n\nfirst line\nGET https://example.com/b\n", line: 1},
			},
		},
		{
			name:   "variables are their own blocks",
			source: "@host = \"example.com\"\nGET https://{{host}}/a\n",
			want: []block{
				{kind: blockVariable, text: "@host = \"example.com\"", line: 1},
				{kind: blockRequest, text: "GET https://{{host}}/a\n", line: 2},
			},
		},
		{
			name:   "an empty document holds no blocks",
			source: "\n\n  \n",
			want:   []block{},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := splitDocument(testCase.source)
			if len(got) != len(testCase.want) {
				t.Fatalf("got %d blocks, want %d: %+v", len(got), len(testCase.want), got)
			}
			for index, want := range testCase.want {
				if got[index] != want {
					t.Errorf("block %d: got %+v, want %+v", index, got[index], want)
				}
			}
		})
	}
}

func TestDetectBodyType(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		body        string
		want        string
	}{
		{name: "declared json", contentType: "application/json", body: "{}", want: BodyTypeJSON},
		{name: "declared xml", contentType: "application/xml; charset=utf-8", body: "<a/>", want: BodyTypeXML},
		{name: "declared graphql", contentType: "application/graphql", body: "query { id }", want: BodyTypeGraphQL},
		{name: "declared graphql wins over json", contentType: "application/graphql-response+json", body: "{}", want: BodyTypeGraphQL},
		{name: "a declared type the body contradicts", contentType: "application/json", body: "query { id }", want: BodyTypeJSON},
		{name: "an unknown declared type falls back to the body", contentType: "text/plain", body: "{\"a\": 1}", want: BodyTypeJSON},
		{name: "sniffed json object", body: "{\n  \"a\": 1\n}", want: BodyTypeJSON},
		{name: "sniffed json array", body: "[1, 2]", want: BodyTypeJSON},
		{name: "sniffed empty json object", body: "{}", want: BodyTypeJSON},
		{name: "sniffed xml", body: "<?xml version=\"1.0\"?>\n<a/>", want: BodyTypeXML},
		{name: "sniffed graphql selection set", body: "{\n  user { id }\n}", want: BodyTypeGraphQL},
		{name: "sniffed graphql operation", body: "query Example {\n  id\n}", want: BodyTypeGraphQL},
		{name: "sniffed graphql fragment", body: "fragment F on User { id }", want: BodyTypeGraphQL},
		{name: "plain text", body: "just some text", want: ""},
		{name: "no body", body: "", want: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			suite := HttpSuite{Header: nethttp.Header{}, Body: testCase.body}
			if testCase.contentType != "" {
				suite.Header.Set("Content-Type", testCase.contentType)
			}
			if got := detectBodyType(suite); got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestIsRecognizedRequest(t *testing.T) {
	cases := []struct {
		name   string
		method string
		uri    string
		want   bool
	}{
		{name: "a known method", method: "GET", uri: "example.com/users", want: true},
		{name: "an unusual method with a url", method: "PURGE", uri: "https://example.com/a", want: true},
		{name: "an unusual method with a path", method: "PURGE", uri: "/a", want: true},
		{name: "prose", method: "hello", uri: "world", want: false},
		{name: "a bare url", method: "", uri: "https://example.com", want: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			suite := HttpSuite{Method: testCase.method, Uri: testCase.uri}
			if got := suite.isRecognizedRequest(); got != testCase.want {
				t.Errorf("got %v, want %v", got, testCase.want)
			}
		})
	}
}
