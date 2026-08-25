package producer

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
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
		Header: map[string]string{"Authorization": "Bearer " + secret, "Accept": secret}, SecretValues: []string{secret},
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
