package producer

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

func fullHistoryProducer(path string, history []HistoryEntry) *Producer {
	widget := &Producer{historyPath: path, history: history}
	widget.historyMode.Store(uint32(HistoryFull))
	return widget
}

func TestHistoryPersistsAndRedactsSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lazyrest", "history.json")
	secret := "super-secret"
	runtimeSecret := "server-issued-secret"
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Name:             "Secret " + secret,
		Method:           "POST-" + secret,
		Uri:              "https://example.test?token=" + secret,
		Header:           http.Header{"Authorization": {"Bearer " + secret}, "Accept": {secret}},
		Body:             secret,
		BodyType:         "json-" + secret,
		HurlFilePath:     "/private/" + secret,
		Variables:        map[string]string{"token": secret},
		GraphQLVariables: `{"token":"` + secret + `"}`,
		GraphQLOperation: "Login" + secret,
		SecretValues:     []string{secret},
	}, runner.Response{
		Code:          "200 " + secret,
		Body:          `{"accessToken":"` + runtimeSecret + `"}`,
		Header:        http.Header{"Set-Cookie": {secret}, "X-Value": {runtimeSecret}},
		Protocol:      "HTTP/2-" + secret,
		GraphQLErrors: []string{"rejected " + runtimeSecret},
	}, errors.New("failed "+runtimeSecret), time.Now())
	widget := fullHistoryProducer(path, []HistoryEntry{entry})
	if err := widget.saveHistory(); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), secret) {
		t.Fatal("persisted history contains a secret")
	}
	if strings.Contains(string(contents), runtimeSecret) {
		t.Fatal("persisted history contains a server-issued secret")
	}
	for _, runtimeField := range []string{`"HurlFilePath":`, `"Variables":`, `"SecretValues":`} {
		if strings.Contains(string(contents), runtimeField) {
			t.Fatalf("persisted history contains runtime-only field %s", runtimeField)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected history permissions: %o", info.Mode().Perm())
	}

	restored := fullHistoryProducer(path, nil)
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 1 || restored.history[0].Suite.Uri == "" {
		t.Fatalf("history was not restored: %+v", restored.history)
	}
	if variables := restored.history[0].Suite.GraphQLVariables; !strings.Contains(variables, "<redacted>") {
		t.Fatalf("redacted GraphQL variables were not restored: %q", variables)
	}
	if graphQLErrors := restored.history[0].Response.GraphQLErrors; len(graphQLErrors) != 1 || !strings.Contains(graphQLErrors[0], "<redacted>") {
		t.Fatalf("redacted GraphQL errors were not restored: %#v", graphQLErrors)
	}
}

func TestHistoryDoesNotPersistResponseReferenceScope(t *testing.T) {
	historyPath := filepath.Join(t.TempDir(), "history.json")
	sourceFilePath := "/private/project/requests.http"
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Name:           "List users",
		Method:         "GET",
		Uri:            "https://example.test/users",
		Header:         http.Header{},
		SourceFilePath: sourceFilePath,
	}, runner.Response{Code: "200 OK"}, nil, time.Now())

	widget := fullHistoryProducer(historyPath, []HistoryEntry{entry})
	if err := widget.saveHistory(); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), sourceFilePath) {
		t.Fatalf("persisted history contains the response reference scope: %s", contents)
	}

	restored := fullHistoryProducer(historyPath, nil)
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 1 || restored.history[0].Suite.SourceFilePath != "" {
		t.Fatalf("response reference scope was restored from history: %+v", restored.history)
	}
}

func TestHistoryRejectsCorruptedFileWithoutReplacingCurrentState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	widget := &Producer{historyPath: path, history: []HistoryEntry{{CreatedAt: time.Now()}}}
	if err := widget.loadHistory(); err == nil {
		t.Fatal("expected corrupted history error")
	}
	if len(widget.history) != 1 {
		t.Fatal("corrupted history replaced in-memory state")
	}
}

func TestBuildReportsCorruptedHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	var reported error
	widget := New()
	widget.Build(Parameters{
		Theme:       theme.NewDefault(),
		HistoryPath: path,
		OnHistoryErrorCallback: func(err error) {
			reported = err
		},
	})
	if reported == nil || !strings.Contains(reported.Error(), "load history") || !strings.Contains(reported.Error(), path) {
		t.Fatalf("corrupted history was not reported with its path: %v", reported)
	}
}

func TestPersistHistoryReportsBackgroundWriteFailure(t *testing.T) {
	root := t.TempDir()
	blocker := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("block directory creation"), 0o600); err != nil {
		t.Fatal(err)
	}
	reported := make(chan error, 1)
	widget := &Producer{
		historyPath: filepath.Join(blocker, "history.json"),
		onHistoryError: func(err error) {
			reported <- err
		},
	}
	widget.persistHistory()
	widget.WaitForHistory()
	select {
	case err := <-reported:
		if !strings.Contains(err.Error(), "persist history") || !strings.Contains(err.Error(), widget.historyPath) {
			t.Fatalf("write failure lacked context: %v", err)
		}
	default:
		t.Fatal("background history write failure was not reported")
	}
}

func TestHistoryReadsVersionOneHeaders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	contents := `{
  "version": 1,
  "entries": [
    {
      "suite": {
        "Name": "List users",
        "Method": "GET",
        "Uri": "https://example.test/users",
        "Header": {"Accept": "application/json"},
        "Body": "",
        "BodyType": "",
        "IsHurl": false,
        "HurlFilePath": "",
        "SecretValues": null
      },
      "response": {"Code": "200 OK", "StatusCode": 200},
      "created_at": "2026-08-25T00:00:00Z"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	widget := fullHistoryProducer(path, nil)
	if err := widget.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(widget.history) != 1 {
		t.Fatalf("version 1 history was not restored: %+v", widget.history)
	}
	if got := widget.history[0].Suite.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("version 1 header was not migrated: %q", got)
	}
}

func TestHistoryKeepsRepeatedHeadersAcrossReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Method: "GET",
		Uri:    "https://example.test",
		Header: http.Header{"Accept": {"application/json", "text/html"}},
	}, runner.Response{Code: "200 OK"}, nil, time.Now())

	widget := fullHistoryProducer(path, []HistoryEntry{entry})
	if err := widget.saveHistory(); err != nil {
		t.Fatal(err)
	}

	restored := fullHistoryProducer(path, nil)
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 1 {
		t.Fatalf("history was not restored: %+v", restored.history)
	}
	if values := restored.history[0].Suite.Header.Values("Accept"); !slices.Equal(values, []string{"application/json", "text/html"}) {
		t.Fatalf("repeated header did not survive a reload: %#v", values)
	}
}

func TestHistoryDoesNotPersistHurlVariables(t *testing.T) {
	secret := "private-token"
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Method:       "HURL",
		Uri:          "workflow.hurl",
		IsHurl:       true,
		HurlFilePath: "/home/user/workflow.hurl",
		Variables:    map[string]string{"token": secret},
		SecretValues: []string{secret},
	}, runner.Response{Body: "token is " + secret}, nil, time.Now())

	if entry.Suite.Variables != nil {
		t.Fatalf("variables were kept in history: %#v", entry.Suite.Variables)
	}
	if strings.Contains(entry.Response.Body, secret) {
		t.Fatalf("Hurl output was not redacted: %q", entry.Response.Body)
	}
}

func TestMetadataHistoryOmitsRequestAndResponseDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	runtimeToken := "server-issued-token"
	privatePath := "https://example.test/users?access=" + runtimeToken
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Name:             "List users",
		Method:           http.MethodPost,
		Uri:              privatePath,
		Header:           http.Header{"X-Trace": {runtimeToken}},
		Body:             `{"accessToken":"` + runtimeToken + `"}`,
		BodyType:         parserhttp.BodyTypeGraphQL,
		GraphQLVariables: `{"token":"` + runtimeToken + `"}`,
	}, runner.Response{
		Code:          "200 OK",
		StatusCode:    http.StatusOK,
		Time:          25 * time.Millisecond,
		ContentLength: 128,
		Body:          `{"accessToken":"` + runtimeToken + `"}`,
		Header:        http.Header{"X-Trace": {runtimeToken}},
		Protocol:      "HTTP/2.0",
		GraphQLErrors: []string{"echoed " + runtimeToken},
	}, errors.New("request failed with "+runtimeToken), time.Now())
	widget := &Producer{historyPath: path, history: []HistoryEntry{entry}}

	if err := widget.saveHistory(); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{runtimeToken, privatePath, "request failed with"} {
		if strings.Contains(string(contents), forbidden) {
			t.Fatalf("metadata history contains private detail %q: %s", forbidden, contents)
		}
	}
	if !strings.Contains(string(contents), `"details_omitted": true`) {
		t.Fatalf("metadata history was not marked as omitted: %s", contents)
	}

	restored := &Producer{historyPath: path}
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 1 {
		t.Fatalf("metadata history was not restored: %+v", restored.history)
	}
	got := restored.history[0]
	if !got.DetailsOmitted || got.Suite.Name != "List users" || got.Suite.Method != http.MethodPost || got.Response.StatusCode != http.StatusOK || got.Response.ContentLength != 128 {
		t.Fatalf("safe metadata was not preserved: %+v", got)
	}
	if got.Suite.Uri != "" || got.Suite.Body != "" || len(got.Suite.Header) != 0 || got.Response.Body != "" || len(got.Response.Header) != 0 || len(got.Response.GraphQLErrors) != 0 {
		t.Fatalf("restored metadata contains request or response details: %+v", got)
	}
	restored.resultAvailable = true
	if _, ok := restored.CurrentResponse(); ok {
		t.Fatal("metadata-only restored entry was exposed for response export")
	}
}

func TestMetadataHistoryRewritesPreviouslyStoredDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	privateDetail := "legacy-private-detail"
	legacy := fullHistoryProducer(path, []HistoryEntry{sanitizedHistoryEntry(parserhttp.HttpSuite{
		Name:   "Legacy request",
		Method: http.MethodPost,
		Uri:    "https://example.test/" + privateDetail,
		Body:   privateDetail,
	}, runner.Response{
		Code:          "200 OK",
		StatusCode:    http.StatusOK,
		Body:          privateDetail,
		ContentLength: len(privateDetail),
	}, nil, time.Now())})
	if err := legacy.saveHistory(); err != nil {
		t.Fatal(err)
	}

	metadata := &Producer{historyPath: path}
	if err := metadata.loadHistory(); err != nil {
		t.Fatal(err)
	}
	metadata.WaitForHistory()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), privateDetail) {
		t.Fatalf("metadata migration left private details on disk: %s", contents)
	}
	if len(metadata.history) != 1 || !metadata.history[0].DetailsOmitted {
		t.Fatalf("legacy entry was not converted to metadata: %+v", metadata.history)
	}
}

func TestHistoryBoundsPersistedBodiesWithoutTouchingMemory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	large := strings.Repeat("x", maxHistoryBodyBytes*2)
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Method: "GET",
		Uri:    "https://example.test",
		Body:   large,
	}, runner.Response{Code: "200 OK", Body: large}, nil, time.Now())

	widget := fullHistoryProducer(path, []HistoryEntry{entry})
	if err := widget.saveHistory(); err != nil {
		t.Fatal(err)
	}

	if len(widget.history[0].Response.Body) != len(large) {
		t.Fatal("the in-memory entry was truncated")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() > 4*maxHistoryBodyBytes {
		t.Fatalf("history file was not bounded: %d bytes", info.Size())
	}

	restored := fullHistoryProducer(path, nil)
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	stored := restored.history[0]
	if len(stored.Response.Body) != maxHistoryBodyBytes || !stored.Response.Truncated {
		t.Fatalf("response body was not bounded: %d bytes truncated=%v", len(stored.Response.Body), stored.Response.Truncated)
	}
	if !strings.HasSuffix(stored.Suite.Body, "... truncated") {
		t.Fatal("a truncated request body was not marked")
	}
}

func TestPersistHistoryWritesOutsideTheCaller(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lazyrest", "history.json")
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Method: "GET",
		Uri:    "https://example.test",
	}, runner.Response{Code: "200 OK"}, nil, time.Now())

	widget := fullHistoryProducer(path, []HistoryEntry{entry})
	widget.persistHistory()
	widget.WaitForHistory()

	restored := fullHistoryProducer(path, nil)
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 1 || restored.history[0].Suite.Uri != "https://example.test" {
		t.Fatalf("history was not written: %+v", restored.history)
	}
}

func TestPersistHistoryKeepsTheNewestSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	widget := &Producer{historyPath: path}
	for index := range 5 {
		widget.history = append(widget.history, sanitizedHistoryEntry(parserhttp.HttpSuite{
			Method: "GET",
			Uri:    fmt.Sprintf("https://example.test/%d", index),
		}, runner.Response{Code: "200 OK"}, nil, time.Now()))
		widget.persistHistory()
	}
	widget.WaitForHistory()

	restored := &Producer{historyPath: path}
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 5 {
		t.Fatalf("the newest snapshot did not win: %d entries", len(restored.history))
	}
}
