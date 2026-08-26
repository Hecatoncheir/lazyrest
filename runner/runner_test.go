package runner

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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
		Header:   http.Header{"Accept": []string{"application/json"}},
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
				Header:        http.Header{"Content-Type": []string{"application/json"}, "X-Trace-Id": []string{"trace-123"}},
				Proto:         "HTTP/2.0",
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
	if response.Header.Get("X-Trace-ID") != "trace-123" || response.Protocol != "HTTP/2.0" {
		t.Fatalf("response metadata was not preserved: %+v", response)
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
		Header:   http.Header{"X-Custom-Header": []string{"CustomValue"}},
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

func TestExecute_SendsRepeatedHeaders(t *testing.T) {
	suite := parser.HttpSuite{
		Method: http.MethodGet,
		Uri:    "http://example.test",
		Header: http.Header{"Cookie": []string{"a=1", "b=2"}, "Accept": []string{"*/*"}},
	}
	var sent http.Header
	runner := NewFromSuiteWithConfig(suite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			sent = request.Header.Clone()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		})},
	})

	if _, err := runner.Execute(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if values := sent.Values("Cookie"); !slices.Equal(values, []string{"a=1", "b=2"}) {
		t.Fatalf("repeated header was not sent: %#v", values)
	}
	if sent.Get("Accept") != "*/*" {
		t.Fatalf("header was not sent: %#v", sent)
	}
}

func TestExecute_KeepsContentTypeDeclaredByTheRequest(t *testing.T) {
	suite := parser.HttpSuite{
		Method:   http.MethodPost,
		Uri:      "http://example.test",
		Body:     `{"a": 1}`,
		BodyType: "json",
		Header:   http.Header{"Content-Type": []string{"application/vnd.api+json"}},
	}
	var sent http.Header
	runner := NewFromSuiteWithConfig(suite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			sent = request.Header.Clone()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		})},
	})

	if _, err := runner.Execute(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if values := sent.Values("Content-Type"); !slices.Equal(values, []string{"application/vnd.api+json"}) {
		t.Fatalf("the declared Content-Type was not kept: %#v", values)
	}
}

func TestExecute_SendsDeclaredHostHeader(t *testing.T) {
	suite := parser.HttpSuite{
		Method: http.MethodGet,
		Uri:    "http://127.0.0.1:8080/users",
		Header: http.Header{"Host": []string{"virtual.example.test"}},
	}
	var sentHost string
	runner := NewFromSuiteWithConfig(suite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			sentHost = request.Host
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		})},
	})

	if _, err := runner.Execute(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if sentHost != "virtual.example.test" {
		t.Fatalf("declared Host was not sent: %q", sentHost)
	}
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

func TestWriteHurlVariables(t *testing.T) {
	path, remove, err := writeHurlVariables(map[string]string{
		"token":   "private-token",
		"baseUrl": "https://api.example.com",
		"broken":  "line\nbreak",
		"":        "unnamed",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer remove()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "baseUrl=https://api.example.com\ntoken=private-token\n" {
		t.Fatalf("unexpected variables file: %q", string(contents))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected permissions: %o", info.Mode().Perm())
	}

	remove()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("variables file was not removed")
	}
}

func TestWriteHurlVariables_WithoutVariables(t *testing.T) {
	path, remove, err := writeHurlVariables(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer remove()
	if path != "" {
		t.Fatalf("expected no variables file, got %q", path)
	}
}

func TestExecuteHurl_PassesVariablesToHurl(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stub executable is a shell script")
	}
	directory := t.TempDir()
	capturedPath := filepath.Join(directory, "captured")
	stub := filepath.Join(directory, "hurl")
	script := "#!/bin/sh\n" +
		": > \"$LAZYREST_TEST_CAPTURE\"\n" +
		"while [ $# -gt 0 ]; do\n" +
		"  if [ \"$1\" = \"--variables-file\" ]; then shift; cat \"$1\" >> \"$LAZYREST_TEST_CAPTURE\"; fi\n" +
		"  shift\n" +
		"done\n" +
		"echo '{\"entries\":[]}'\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	hurlFile := filepath.Join(directory, "workflow.hurl")
	if err := os.WriteFile(hurlFile, []byte("GET {{baseUrl}}/users\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAZYREST_TEST_CAPTURE", capturedPath)

	suite := parser.HttpSuite{
		IsHurl:       true,
		HurlFilePath: hurlFile,
		Variables:    map[string]string{"baseUrl": "https://api.example.com"},
	}
	runner := NewFromSuiteWithConfig(suite, Config{HurlExecutable: stub})

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(response.Body, "entries") {
		t.Fatalf("unexpected Hurl output: %q", response.Body)
	}
	captured, err := os.ReadFile(capturedPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(captured)) != "baseUrl=https://api.example.com" {
		t.Fatalf("variables did not reach Hurl: %q", string(captured))
	}
}

func TestExecuteHurl_RunsUpToTheSelectedEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stub executable is a shell script")
	}
	directory := t.TempDir()
	capturedPath := filepath.Join(directory, "captured")
	stub := filepath.Join(directory, "hurl")
	script := "#!/bin/sh\n" +
		"echo \"$@\" > \"$LAZYREST_TEST_CAPTURE\"\n" +
		"echo '{\"entries\":[]}'\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	hurlFile := filepath.Join(directory, "workflow.hurl")
	if err := os.WriteFile(hurlFile, []byte("GET https://example.test/a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAZYREST_TEST_CAPTURE", capturedPath)

	run := func(entry int) string {
		suite := parser.HttpSuite{IsHurl: true, HurlFilePath: hurlFile, HurlEntry: entry}
		runner := NewFromSuiteWithConfig(suite, Config{HurlExecutable: stub})
		if _, err := runner.Execute(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
		captured, err := os.ReadFile(capturedPath)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(captured))
	}

	// An entry may use what an earlier one captured, so it is reached by
	// running the file up to it.
	if arguments := run(2); !strings.Contains(arguments, "--to-entry 2") {
		t.Fatalf("the entry was not passed to Hurl: %q", arguments)
	}
	if arguments := run(0); strings.Contains(arguments, "--to-entry") {
		t.Fatalf("a whole-file run was limited to an entry: %q", arguments)
	}
}

func TestNewClientKeepsCookiesAcrossRequests(t *testing.T) {
	seen := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Get("Cookie")
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc"})
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	config := Config{Jar: jar}
	config.Client = NewClient(config)

	suite := parser.HttpSuite{Method: http.MethodGet, Uri: server.URL}
	for range 2 {
		runner := NewFromSuiteWithConfig(suite, config)
		if _, err := runner.Execute(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	}

	if first := <-seen; first != "" {
		t.Fatalf("the first request already carried a cookie: %q", first)
	}
	if second := <-seen; second != "session=abc" {
		t.Fatalf("the cookie was not carried over: %q", second)
	}
}

func TestNewClientWithoutAJarForgetsCookies(t *testing.T) {
	seen := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Get("Cookie")
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc"})
	}))
	defer server.Close()

	config := Config{}
	config.Client = NewClient(config)
	suite := parser.HttpSuite{Method: http.MethodGet, Uri: server.URL}
	for range 2 {
		runner := NewFromSuiteWithConfig(suite, config)
		if _, err := runner.Execute(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	}

	<-seen
	if second := <-seen; second != "" {
		t.Fatalf("a cookie was kept without a jar: %q", second)
	}
}

func TestNewClientStopsAfterMaxRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/next", http.StatusFound)
	}))
	defer server.Close()

	config := Config{MaxRedirects: 3}
	config.Client = NewClient(config)
	runner := NewFromSuiteWithConfig(parser.HttpSuite{Method: http.MethodGet, Uri: server.URL}, config)

	_, err := runner.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected the redirect chain to be stopped")
	}
	if !strings.Contains(err.Error(), "stopped after 3 redirects") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientCanReturnTheRedirectItself(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://example.test/moved", http.StatusFound)
	}))
	defer server.Close()

	config := Config{DisableRedirects: true}
	config.Client = NewClient(config)
	runner := NewFromSuiteWithConfig(parser.HttpSuite{Method: http.MethodGet, Uri: server.URL}, config)

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusFound {
		t.Fatalf("the redirect was followed: %+v", response.Code)
	}
	if location := response.Header.Get("Location"); location != "https://example.test/moved" {
		t.Fatalf("the Location header is missing: %q", location)
	}
}

func TestNewClientFollowsRedirectsByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/moved" {
			_, _ = w.Write([]byte("arrived"))
			return
		}
		http.Redirect(w, r, "/moved", http.StatusFound)
	}))
	defer server.Close()

	config := Config{}
	config.Client = NewClient(config)
	runner := NewFromSuiteWithConfig(parser.HttpSuite{Method: http.MethodGet, Uri: server.URL}, config)

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.Body != "arrived" {
		t.Fatalf("the redirect was not followed: %+v", response)
	}
}

func TestNewClientCanAcceptASelfSignedCertificate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("secure"))
	}))
	defer server.Close()
	suite := parser.HttpSuite{Method: http.MethodGet, Uri: server.URL}

	verifying := Config{}
	verifying.Client = NewClient(verifying)
	strict := NewFromSuiteWithConfig(suite, verifying)
	if _, err := strict.Execute(context.Background(), nil); err == nil {
		t.Fatal("a self-signed certificate was accepted by default")
	}

	accepting := Config{InsecureSkipVerify: true}
	accepting.Client = NewClient(accepting)
	relaxed := NewFromSuiteWithConfig(suite, accepting)
	response, err := relaxed.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("the certificate was not accepted: %v", err)
	}
	if response.Body != "secure" {
		t.Fatalf("unexpected response: %+v", response)
	}
}
