package hurl

import (
	"os"
	"testing"
)

func TestGetSuitesFromFile(t *testing.T) {
	// Create a temporary hurl file
	// We add an extra header to ensure the body is swallowed by the last header node,
	// mimicking how the .http parser handles bodies in some tree-sitter configurations.
	content := `GET https://example.com/1
Header1: Value1
Header2: Value2

Body content 1

POST https://example.com/2
Content-Type: application/json
Extra-Header: Value

{"key": "value"}
`
	tmpFile, err := os.CreateTemp("", "test*.hurl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	parser, err := NewParser()
	if err != nil {
		t.Fatal(err)
	}

	suites, err := parser.GetSuitesFromFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(suites) != 2 {
		t.Errorf("expected 2 suites, got %d", len(suites))
	}

	// Check first suite
	s1 := suites[0]
	if s1.Method != "GET" || s1.Uri != "https://example.com/1" {
		t.Errorf("suite 1 mismatch: %+v", s1)
	}
	if s1.Header["Header1"] != "Value1" || s1.Header["Header2"] != "Value2" {
		t.Errorf("suite 1 headers mismatch: %+v", s1.Header)
	}
	if s1.Body != "Body content 1" {
		t.Errorf("suite 1 body mismatch: %q", s1.Body)
	}

	// Check second suite
	s2 := suites[1]
	if s2.Method != "POST" || s2.Uri != "https://example.com/2" {
		t.Errorf("suite 2 mismatch: %+v", s2)
	}
	if s2.Header["Content-Type"] != "application/json" {
		t.Errorf("suite 2 headers mismatch: %+v", s2.Header)
	}
	if s2.Body != `{"key": "value"}` {
		t.Errorf("suite 2 body mismatch: %q", s2.Body)
	}
}
