package http

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetSuitesFromFile_Success(t *testing.T) {
	// 1. Setup: Создаем временный файл с имитацией *.http содержимого.
	tempDir, err := os.MkdirTemp("", "test_parser_")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "test_requests.http")
	// Имитация тестового содержимого HTTP-файла с одним запросом и одним набором запросов.
	mockContent := `
#@suite("Test Suite")
GET http://example.com/api/resource
Content-Type: application/json

#@test("Single Test")
POST http://example.com/api/data
Content-Type: application/json

#@suite("Another Suite")
GET http://example.com/api/status
`
	if err := os.WriteFile(testFilePath, []byte(mockContent), 0644); err != nil {
		t.Fatalf("Не удалось записать тестовый файл: %v", err)
	}

	// 2. Execution: Использование подсистем парсинга
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Не удалось создать парсер: %v", err)
	}
	defer parser.Close()

	suites, err := parser.GetSuitesFromFile(testFilePath)
	if err != nil {
		t.Fatalf("Ошибка при получении наборов запросов: %v", err)
	}

	// 3. Assertions
	if len(suites) != 3 {
		t.Errorf("Ожидалось 3 набора запросов, получено %d", len(suites))
	}

	for i, s := range suites {
		if s.Method == "" || s.Uri == "" {
			t.Errorf("Набор запросов %d имеет пустые method или uri", i)
		}
	}
	if suites[0].Name != "Test Suite" || suites[1].Name != "Single Test" || suites[2].Name != "Another Suite" {
		t.Errorf("request names were not parsed: %+v", suites)
	}
}

func TestGetSuitesFromFile_BodyTypes(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "test_body_")
	defer os.RemoveAll(tempDir)
	testFilePath := filepath.Join(tempDir, "bodies.http")

	mockContent := `GET http://example.com/json
{
"key": "val"
}

POST http://example.com/xml
<?xml version="1.0"?>
<tag>
</tag>

GET http://example.com/graphql
query {
user { id }
}
`
	os.WriteFile(testFilePath, []byte(mockContent), 0644)

	parser, _ := NewParser()
	defer parser.Close()

	suites, err := parser.GetSuitesFromFile(testFilePath)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if len(suites) != 3 {
		t.Fatalf("Expected 3 suites, got %d", len(suites))
	}

	expectedTypes := []string{"json", "xml", "graphql"}
	for i, s := range suites {
		if s.BodyType != expectedTypes[i] {
			t.Errorf("Suite %d: expected body type %s, got %s", i, expectedTypes[i], s.BodyType)
		}
		if s.Body == "" {
			t.Errorf("Suite %d: expected non-empty body", i)
		}
	}
}

func TestGetSuitesFromFile_FileNotFound(t *testing.T) {
	parser, _ := NewParser()
	defer parser.Close()

	_, err := parser.GetSuitesFromFile("non_existent_file.http")
	if err == nil {
		t.Error("Expected file not found error, but got nil")
	}
}

func TestParseFile_VariablesAndDiagnostics(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "variables.http")
	content := `@host = "example.com"
@token = "secret-value"

# @name List users
GET https://{{host}}/users
Authorization: Bearer {{token}}
X-Missing: {{missing}}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	result, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Suites) != 1 {
		t.Fatalf("expected one request, got %d", len(result.Suites))
	}
	suite := result.Suites[0]
	if suite.Name != "List users" {
		t.Errorf("unexpected name: %q", suite.Name)
	}
	if suite.Uri != "https://example.com/users" {
		t.Errorf("unexpected resolved URI: %q", suite.Uri)
	}
	if suite.Header["Authorization"] != "Bearer secret-value" {
		t.Errorf("unexpected resolved header: %q", suite.Header["Authorization"])
	}
	foundMissing := false
	for _, diagnostic := range result.Diagnostics {
		if strings.Contains(diagnostic.Message, "undefined variable: missing") {
			foundMissing = true
		}
	}
	if !foundMissing {
		t.Errorf("expected missing-variable diagnostic, got %+v", result.Diagnostics)
	}
}
