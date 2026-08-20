package http

import (
	"os"
	"path/filepath"
	"testing"
)

// NOTE: Due to the deep dependency on the external tree-sitter bindings and actual file reads,
// these tests require a complex setup (e.g., mocking the sitter library).
// Here, we focus on testing the public API flow: GetSuitesFromFile, simulating the necessary file read.

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

	// В реальном тесте здесь бы требовалась мокировка getTree,
	// но мы полагаемся на работоспособность всей цепочки.
	suites, err := parser.GetSuitesFromFile(testFilePath)
	if err != nil {
		t.Fatalf("Ошибка при получении наборов запросов: %v", err)
	}

	// 3. Assertions
	if len(suites) != 3 {
		t.Errorf("Ожидалось 3 набора запросов (one suite, one test, one another suite), получено %d", len(suites))
	}

	// Проверяем, что хотя бы одно поле заполнено для каждого набора запросов
	for i, s := range suites {
		if s.Method == "" || s.Uri == "" {
			t.Errorf("Набор запросов %d имеет пустые method или uri", i)
		}
	}
}

func TestGetSuitesFromFile_FileNotFound(t *testing.T) {
	parser, _ := NewParser()
	defer parser.Reset()
	
	// Ожидаем отказ из-за несуществующего файла
	_, err := parser.GetSuitesFromFile("non_existent_file.http")
	if err == nil {
		t.Error("Expected file not found error, but got nil")
	}
}
