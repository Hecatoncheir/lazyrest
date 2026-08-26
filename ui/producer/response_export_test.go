package producer

import (
	"net/http"
	"strings"
	"testing"
	"time"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

func TestCurrentResponseExportsTheVisibleBodyWithoutMarkupOrSecrets(t *testing.T) {
	createdAt := time.Date(2026, time.August, 26, 14, 5, 12, 0, time.UTC)
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Name:         "List Users",
		SecretValues: []string{"secret-value"},
	}, runner.Response{
		Code:      "200 OK",
		Protocol:  "HTTP/2.0",
		Truncated: true,
		Header: http.Header{
			"Content-Type": {"application/json"},
			"Set-Cookie":   {"session=secret-value"},
		},
		Body: `{"name":"Ada","token":"secret-value"}`,
	}, nil, createdAt)
	widget := &Producer{
		history:         []HistoryEntry{entry},
		historyIndex:    0,
		resultAvailable: true,
		bodyViewMode:    BodyViewPretty,
	}

	exported, ok := widget.CurrentResponse()
	if !ok {
		t.Fatal("visible response was not available for export")
	}
	if !strings.Contains(exported.Body, "\n  \"name\": \"Ada\"") {
		t.Fatalf("pretty body was not exported: %q", exported.Body)
	}
	if strings.Contains(exported.Body, "secret-value") || strings.Contains(exported.Full, "secret-value") {
		t.Fatalf("export contains a secret: %+v", exported)
	}
	if strings.Contains(exported.Full, "[green]") || strings.Contains(exported.Full, "Response:") {
		t.Fatalf("export contains presentation markup: %q", exported.Full)
	}
	for _, want := range []string{"HTTP/2.0 200 OK", "Content-Type: application/json", "Set-Cookie: <redacted>"} {
		if !strings.Contains(exported.Full, want) {
			t.Errorf("full response is missing %q: %q", want, exported.Full)
		}
	}
	if exported.SuggestedFileName != "list-users-20260826-140512.json" {
		t.Fatalf("unexpected filename: %q", exported.SuggestedFileName)
	}
	if !exported.Truncated {
		t.Fatal("truncation state was lost")
	}

	widget.bodyViewMode = BodyViewRaw
	raw, ok := widget.CurrentResponse()
	if !ok || raw.Body != raw.RawBody || strings.Contains(raw.Body, "\n") {
		t.Fatalf("raw body was not exported unchanged: %+v", raw)
	}
}

func TestCurrentResponseRejectsAnUnavailableOrRunningResult(t *testing.T) {
	widget := &Producer{}
	if _, ok := widget.CurrentResponse(); ok {
		t.Fatal("empty producer exposed a response")
	}

	widget.history = []HistoryEntry{{Response: runner.Response{Code: "200 OK", Body: "ok"}}}
	widget.resultAvailable = true
	_, _ = widget.StartRun()
	if _, ok := widget.CurrentResponse(); ok {
		t.Fatal("producer exposed an earlier response while a new one was running")
	}
	widget.CancelActive()
}

func TestCurrentResponseFollowsTheSelectedHistoryEntry(t *testing.T) {
	widget := &Producer{
		history: []HistoryEntry{
			{Response: runner.Response{Code: "200 OK", Body: "first"}},
			{Response: runner.Response{Code: "200 OK", Body: "second"}},
		},
		resultAvailable: true,
		bodyViewMode:    BodyViewRaw,
	}
	for index, want := range []string{"first", "second"} {
		widget.historyIndex = index
		exported, ok := widget.CurrentResponse()
		if !ok || exported.Body != want {
			t.Fatalf("history index %d exported %+v, available=%v", index, exported, ok)
		}
	}
}
