package hurl

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/parser/http"
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

	if len(suites) != 2 {
		t.Fatalf("expected one suite per entry, got %d: %+v", len(suites), suites)
	}

	for index, suite := range suites {
		if !suite.IsHurl {
			t.Errorf("entry %d is not marked as Hurl", index)
		}
		if suite.HurlFilePath != tmpFile.Name() {
			t.Errorf("entry %d has path %q, want %q", index, suite.HurlFilePath, tmpFile.Name())
		}
		if suite.HurlEntry != index+1 {
			t.Errorf("entry %d is numbered %d", index, suite.HurlEntry)
		}
		if suite.Name == "" {
			t.Errorf("entry %d has no display name", index)
		}
	}
	if suites[0].Method != "GET" || suites[0].Uri != "https://example.com/1" {
		t.Errorf("unexpected first entry: %+v", suites[0])
	}
	if suites[1].Method != "POST" || suites[1].Uri != "https://example.com/2" {
		t.Errorf("unexpected second entry: %+v", suites[1])
	}
	if !strings.Contains(suites[0].Body, "Body content 1") {
		t.Errorf("the first entry lost its body: %q", suites[0].Body)
	}
	if strings.Contains(suites[0].Body, "example.com/2") {
		t.Errorf("the first entry swallowed the second: %q", suites[0].Body)
	}
}

func TestGetSuitesFromFile_KeepsTheWholeFileWhenNoEntryIsRecognized(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prose.hurl")
	if err := os.WriteFile(path, []byte("# nothing runnable here\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatal(err)
	}
	suites, err := parser.GetSuitesFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(suites) != 1 || suites[0].Method != "HURL" || suites[0].HurlEntry != 0 {
		t.Fatalf("expected the whole file as one suite, got %+v", suites)
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

func TestGetSuitesFromFileWithOptions_CarriesVariablesAndSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "workflow.hurl")
	if err := os.WriteFile(path, []byte("GET {{baseUrl}}/users\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatal(err)
	}

	suites, err := parser.GetSuitesFromFileWithOptions(path, http.ParseOptions{
		Variables: map[string]string{
			"host":    "api.example.com",
			"baseUrl": "https://{{host}}",
			"token":   "private-token",
		},
		SecretVariables: []string{"token"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(suites) != 1 {
		t.Fatalf("expected one runnable Hurl file, got %d", len(suites))
	}

	suite := suites[0]
	if suite.Variables["baseUrl"] != "https://api.example.com" {
		t.Errorf("nested variable was not resolved: %#v", suite.Variables)
	}
	if !slices.Contains(suite.SecretValues, "private-token") {
		t.Errorf("secret was not marked for redaction: %#v", suite.SecretValues)
	}
	if suite.Redact("token is private-token") != "token is <redacted>" {
		t.Errorf("Hurl output would not be redacted: %q", suite.Redact("token is private-token"))
	}
}
