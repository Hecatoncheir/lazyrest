package http

import (
	"os"
	"path/filepath"
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
	defer parser.Reset()

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
	defer parser.Reset()

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
	defer parser.Reset()
	
	_, err := parser.GetSuitesFromFile("non_existent_file.http")
	if err == nil {
		t.Error("Expected file not found error, but got nil")
	}
}
