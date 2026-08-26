package http

import (
	"slices"
	"testing"
)

func TestParseRequestText_KeepsRepeatedHeaders(t *testing.T) {
	parsed := parseRequestText("GET https://example.com/c\nCookie: a=1\nCookie: b=2\nAccept: application/json\n")

	if parsed.Method != "GET" || parsed.Uri != "https://example.com/c" {
		t.Fatalf("unexpected request line: %q %q", parsed.Method, parsed.Uri)
	}
	if values := parsed.Header.Values("Cookie"); !slices.Equal(values, []string{"a=1", "b=2"}) {
		t.Fatalf("repeated header was not kept: %#v", values)
	}
	if parsed.Header.Get("Accept") != "application/json" {
		t.Fatalf("header after a repeated one was lost: %#v", parsed.Header)
	}
	if parsed.Body != "" {
		t.Fatalf("a header was read as the body: %q", parsed.Body)
	}
}

func TestParseRequestText_KeepsHeadersAfterWildcardValue(t *testing.T) {
	parsed := parseRequestText("GET https://example.com/d\nAccept: */*\nX-After: yes\n")

	if parsed.Header.Get("Accept") != "*/*" || parsed.Header.Get("X-After") != "yes" {
		t.Fatalf("wildcard value swallowed the next header: %#v", parsed.Header)
	}
	if parsed.Body != "" {
		t.Fatalf("a header was read as the body: %q", parsed.Body)
	}
}

func TestParseRequestText_BodySeparation(t *testing.T) {
	cases := []struct {
		name       string
		text       string
		wantHeader string
		wantBody   string
	}{
		{
			name:     "blank line before the body",
			text:     "POST https://example.com\nContent-Type: application/json\n\n{\"a\": 1}\n",
			wantBody: `{"a": 1}`,
		},
		{
			name:     "body without a blank line",
			text:     "GET https://example.com\n{\n\"key\": \"val\"\n}\n",
			wantBody: "{\n\"key\": \"val\"\n}",
		},
		{
			name:       "colon inside the value",
			text:       "GET https://example.com\nHost: example.com:8080\n",
			wantHeader: "example.com:8080",
		},
		{
			name:       "folded continuation line",
			text:       "GET https://example.com\nHost: example.com\n  continued\n",
			wantHeader: "example.com continued",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			parsed := parseRequestText(testCase.text)
			if parsed.Body != testCase.wantBody {
				t.Errorf("unexpected body: %q, want %q", parsed.Body, testCase.wantBody)
			}
			if testCase.wantHeader != "" && parsed.Header.Get("Host") != testCase.wantHeader {
				t.Errorf("unexpected header: %q, want %q", parsed.Header.Get("Host"), testCase.wantHeader)
			}
		})
	}
}
