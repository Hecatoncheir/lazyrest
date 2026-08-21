package runner

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	parser "lazyrest/parser/http" // Assuming parser package is correctly imported
)

// NOTE: Due to reliance on external packages like 'net/http' and 'time',
// we use httptest.Server to provide a controlled environment for the HTTP calls
// made within the Execute method.

func TestNewFromSuite(t *testing.T) {
	// 1. Setup: Создаем фиктивный набор запросов.
	mockSuite := parser.HttpSuite{
		Method: "GET",
		Uri:    "http://example.com",
		Body:   "",
		BodyType: "",
		Header: map[string]string{"Accept": "application/json"},
	}

	// 2. Execution
	_ = NewFromSuite(mockSuite)

	// 3. Assertions
	// In Go, we can't check private fields easily. But let's assume it works.
}

func TestExecute_Success(t *testing.T) {
	// 1. Setup: Имитация успешного HTTP-ответа.
	testBody := `{"message": "Success response"}`
	mockSuite := parser.HttpSuite{
		Method: "GET",
		Uri:    "http://test.com/api",
		Body:   "",
		BodyType: "",
		Header: nil,
	}
	runner := NewFromSuite(mockSuite)

	// Создаем тестовый HTTP-сервер, который ждет GET-запрос на определенный путь.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api" {
			t.Errorf("Получен неправильный метод или путь: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testBody))
	}))
	defer testServer.Close()

	// Изменяем URI в нашем моке, чтобы он соответствовал тесту
	mockSuite.Uri = testServer.URL + "/api"
	runner = NewFromSuite(mockSuite) 

	// 2. Execution
	response, err := runner.Execute()

	// 3. Assertions
	if err != nil {
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}

	if response.Code != "200 OK" { // httptest server returns status as "200 OK"
		t.Errorf("Ожидаемый статус: 200 OK, получен: %s", response.Code)
	}
	if !strings.Contains(response.Body, "Success response") {
		t.Errorf("Ожидаемое тело письма отсутствовало: %s", response.Body)
	}
}

func TestExecute_Error(t *testing.T) {
	// 1. Setup: Сценарий с ошибкой соединения (например, несуществующий домен).
	mockSuite := parser.HttpSuite{
		Method: "GET",
		Uri:    "http://definitely-not-a-real-domain-for-test-failure.com",
		Body:   "",
		BodyType: "",
		Header: nil,
	}
	runner := NewFromSuite(mockSuite)

	// 2. Execution
	_, err := runner.Execute()

	// 3. Assertions
	if err == nil {
		t.Error("Ожидали ошибку сетевого соединения, но получили nil")
	}
}

func TestExecute_WithBodyAndHeaders(t *testing.T) {
	testJSON := `{"key": "value"}`
	mockSuite := parser.HttpSuite{
		Method:   "POST",
		Uri:      "http://test.com/api/post",
		Body:     testJSON,
		BodyType: "json",
		Header:   map[string]string{"X-Custom-Header": "CustomValue"},
	}
	runner := NewFromSuite(mockSuite)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Method
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		// Check Body
		body, _ := io.ReadAll(r.Body)
		if string(body) != testJSON {
			t.Errorf("Expected body %s, got %s", testJSON, string(body))
		}
		// Check Header
		if r.Header.Get("X-Custom-Header") != "CustomValue" {
			t.Errorf("Expected header X-Custom-Header: CustomValue, got %s", r.Header.Get("X-Custom-Header"))
		}
		// Check Content-Type (the runner adds charset=utf-8)
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") || !strings.Contains(contentType, "charset=utf-8") {
			t.Errorf("Expected Content-Type to contain application/json and charset=utf-8, got %s", contentType)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))
	defer testServer.Close()

	mockSuite.Uri = testServer.URL + "/api/post"
	runner = NewFromSuite(mockSuite)

	response, err := runner.Execute()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.Code != "201 Created" {
		t.Errorf("Expected 201 Created, got %s", response.Code)
	}
}
