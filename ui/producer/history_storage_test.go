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
)

func TestHistoryPersistsAndRedactsSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lazyrest", "history.json")
	secret := "super-secret"
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Name: "Secret " + secret, Uri: "https://example.test?token=" + secret, Body: secret,
		Header: http.Header{"Authorization": {"Bearer " + secret}, "Accept": {secret}}, SecretValues: []string{secret},
	}, runner.Response{Body: secret, Header: http.Header{"Set-Cookie": {secret}, "X-Value": {secret}}}, errors.New("failed "+secret), time.Now())
	widget := &Producer{historyPath: path, history: []HistoryEntry{entry}}
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
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected history permissions: %o", info.Mode().Perm())
	}

	restored := &Producer{historyPath: path}
	if err := restored.loadHistory(); err != nil {
		t.Fatal(err)
	}
	if len(restored.history) != 1 || restored.history[0].Suite.Uri == "" {
		t.Fatalf("history was not restored: %+v", restored.history)
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

	widget := &Producer{historyPath: path}
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

	widget := &Producer{historyPath: path, history: []HistoryEntry{entry}}
	if err := widget.saveHistory(); err != nil {
		t.Fatal(err)
	}

	restored := &Producer{historyPath: path}
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

func TestHistoryBoundsPersistedBodiesWithoutTouchingMemory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	large := strings.Repeat("x", maxHistoryBodyBytes*2)
	entry := sanitizedHistoryEntry(parserhttp.HttpSuite{
		Method: "GET",
		Uri:    "https://example.test",
		Body:   large,
	}, runner.Response{Code: "200 OK", Body: large}, nil, time.Now())

	widget := &Producer{historyPath: path, history: []HistoryEntry{entry}}
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

	restored := &Producer{historyPath: path}
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

	widget := &Producer{historyPath: path, history: []HistoryEntry{entry}}
	widget.persistHistory()
	widget.WaitForHistory()

	restored := &Producer{historyPath: path}
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
