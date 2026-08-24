package hurl

import (
	"os"
	"testing"
)

func TestGetSuitesFromFile(t *testing.T) {
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

	if len(suites) != 1 {
		t.Fatalf("expected one runnable Hurl file, got %d", len(suites))
	}

	suite := suites[0]
	if !suite.IsHurl {
		t.Error("expected Hurl suite")
	}
	if suite.HurlFilePath != tmpFile.Name() {
		t.Errorf("expected path %q, got %q", tmpFile.Name(), suite.HurlFilePath)
	}
	if suite.Method != "HURL" || suite.Uri != tmpFile.Name() {
		t.Errorf("unexpected Hurl suite: %+v", suite)
	}
	if suite.Name == "" {
		t.Error("expected display name")
	}
}

func TestGetSuitesFromFile_FileNotFound(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := parser.GetSuitesFromFile("missing.hurl"); err == nil {
		t.Error("expected file-not-found error")
	}
}
