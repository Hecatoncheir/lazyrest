package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResponseFromHurlReportBuildsAnHTTPResponse(t *testing.T) {
	directory := t.TempDir()
	store := filepath.Join(directory, "store")
	if err := os.Mkdir(store, 0o700); err != nil {
		t.Fatal(err)
	}
	body := `{"ok":true}`
	if err := os.WriteFile(filepath.Join(store, "response.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	report := `[{
  "success": false,
  "entries": [{
    "asserts": [{"line": 7, "success": false, "message": "status mismatch"}],
    "calls": [{"response": {
      "body": "store/response.json",
      "headers": [{"name": "content-type", "value": "application/json"}, {"name": "set-cookie", "value": "a=1"}, {"name": "set-cookie", "value": "b=2"}],
      "http_version": "HTTP/2",
      "status": 201
    }}]
  }]
}]`
	if err := os.WriteFile(filepath.Join(directory, "report.json"), []byte(report), 0o600); err != nil {
		t.Fatal(err)
	}

	response, available, err := responseFromHurlReport(directory, 25*time.Millisecond, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if !available || response.Code != "201 Created" || response.StatusCode != 201 || response.Protocol != "HTTP/2" {
		t.Fatalf("unexpected Hurl response metadata: %+v", response)
	}
	if response.Body != body || response.ContentLength != len(body) || response.StoredLength != len(body) {
		t.Fatalf("unexpected Hurl response body: %+v", response)
	}
	if values := response.Header.Values("Set-Cookie"); len(values) != 2 {
		t.Fatalf("repeated response headers were lost: %#v", response.Header)
	}
	if len(response.AssertionErrors) != 1 || response.AssertionErrors[0] != "status mismatch" || response.IsSuccessful() {
		t.Fatalf("assertion failure was not preserved: %+v", response)
	}
}

func TestResponseFromHurlReportTruncatesTheStoredBody(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "body"), []byte("123456"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := `[{"success":true,"entries":[{"calls":[{"response":{"body":"body","status":200}}]}]}]`
	if err := os.WriteFile(filepath.Join(directory, "report.json"), []byte(report), 0o600); err != nil {
		t.Fatal(err)
	}

	response, _, err := responseFromHurlReport(directory, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	if response.Body != "1234" || response.ContentLength != 6 || response.StoredLength != 4 || !response.Truncated {
		t.Fatalf("unexpected truncated Hurl response: %+v", response)
	}
}

func TestResponseFromHurlReportRejectsEscapingBodyPath(t *testing.T) {
	directory := t.TempDir()
	report := `[{"success":true,"entries":[{"calls":[{"response":{"body":"../secret","status":200}}]}]}]`
	if err := os.WriteFile(filepath.Join(directory, "report.json"), []byte(report), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := responseFromHurlReport(directory, 0, 4)
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected an escaping path error, got %v", err)
	}
}
