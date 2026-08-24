package runner

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	parser "github.com/Hecatoncheir/lazyrest/parser/http"
)

func TestNewFromSuite(t *testing.T) {
	// 1. Setup: Создаем фиктивный набор запросов.
	mockSuite := parser.HttpSuite{
		Method:   "GET",
		Uri:      "http://example.com",
		Body:     "",
		BodyType: "",
		Header:   map[string]string{"Accept": "application/json"},
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
		Method:   "GET",
		Uri:      "http://test.com/api",
		Body:     "",
		BodyType: "",
		Header:   nil,
	}
	runner := NewFromSuiteWithConfig(mockSuite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/api" {
				t.Errorf("Получен неправильный метод или путь: %s %s", r.Method, r.URL.Path)
			}
			return &http.Response{
				StatusCode:    http.StatusOK,
				Body:          io.NopCloser(strings.NewReader(testBody)),
				ContentLength: int64(len(testBody)),
				Header:        make(http.Header),
			}, nil
		})},
	})

	// 2. Execution
	response, err := runner.Execute(context.Background(), nil)

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
	mockSuite := parser.HttpSuite{
		Method:   "GET",
		Uri:      "http://example.test",
		Body:     "",
		BodyType: "",
		Header:   nil,
	}
	runner := NewFromSuiteWithConfig(mockSuite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("connection failed")
		})},
	})

	// 2. Execution
	_, err := runner.Execute(context.Background(), nil)

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
	runner := NewFromSuiteWithConfig(mockSuite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
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

			return &http.Response{
				StatusCode:    http.StatusCreated,
				Body:          io.NopCloser(strings.NewReader("created")),
				ContentLength: int64(len("created")),
				Header:        make(http.Header),
			}, nil
		})},
	})

	response, err := runner.Execute(context.Background(), nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.Code != "201 Created" {
		t.Errorf("Expected 201 Created, got %s", response.Code)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestExecute_RespectsContextCancellation(t *testing.T) {
	suite := parser.HttpSuite{Method: http.MethodGet, Uri: "http://example.test"}
	runner := NewFromSuiteWithConfig(suite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			<-request.Context().Done()
			return nil, request.Context().Err()
		})},
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := runner.Execute(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestExecute_RespectsConfiguredTimeout(t *testing.T) {
	suite := parser.HttpSuite{Method: http.MethodGet, Uri: "http://example.test"}
	runner := NewFromSuiteWithConfig(suite, Config{
		Timeout: 10 * time.Millisecond,
		Client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			<-request.Context().Done()
			return nil, request.Context().Err()
		})},
	})

	_, err := runner.Execute(context.Background(), nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestExecute_TruncatesLargeResponses(t *testing.T) {
	suite := parser.HttpSuite{Method: http.MethodGet, Uri: "http://example.test"}
	runner := NewFromSuiteWithConfig(suite, Config{
		MaxResponseBytes: 4,
		Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Body:          io.NopCloser(strings.NewReader("123456")),
				ContentLength: 6,
				Header:        make(http.Header),
			}, nil
		})},
	})

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.Body != "1234" || !response.Truncated {
		t.Fatalf("expected truncated response, got %+v", response)
	}
}

func TestExecuteHurl_RequiresFilePath(t *testing.T) {
	runner := NewFromSuite(parser.HttpSuite{IsHurl: true})

	_, err := runner.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected empty Hurl path error")
	}
}

func TestExecuteHurl_UsesConfiguredExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	tempDir := t.TempDir()
	executable := filepath.Join(tempDir, "fake-hurl")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\nprintf '{\"ok\":true}'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	hurlFile := filepath.Join(tempDir, "session.hurl")
	if err := os.WriteFile(hurlFile, []byte("GET https://example.com\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runner := NewFromSuiteWithConfig(parser.HttpSuite{
		IsHurl:       true,
		HurlFilePath: hurlFile,
	}, Config{HurlExecutable: executable})

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != "OK" || response.Body != `{"ok":true}` {
		t.Fatalf("unexpected Hurl response: %+v", response)
	}
}
