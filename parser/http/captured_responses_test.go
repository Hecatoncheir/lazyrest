package http

import (
	"fmt"
	nethttp "net/http"
	"strings"
	"testing"
)

func TestResponseStoreReturnsSafeSortedCapturesAndClears(t *testing.T) {
	store := ResponseStore{}
	store.Record(HttpSuite{SourceFilePath: "z.http", Name: " second "}, ResponseValue{
		Status: "201 Created",
		Body:   "super-secret-body",
		Header: nethttp.Header{"Authorization": {"Bearer super-secret-token"}},
	})
	store.Record(HttpSuite{SourceFilePath: "a.http", Name: "first"}, ResponseValue{
		Status: "200 OK",
		Body:   "ok",
		Header: nethttp.Header{"Content-Type": {"text/plain"}, "X-Token": {"secret"}},
	})

	captures := store.Captures()
	if len(captures) != 2 || captures[0].SourceFilePath != "a.http" || captures[1].Name != "second" {
		t.Fatalf("captures are not sorted or trimmed: %+v", captures)
	}
	if captures[0].Status != "200 OK" || captures[0].BodyBytes != 2 || captures[0].HeaderCount != 2 {
		t.Fatalf("unexpected capture summary: %+v", captures[0])
	}
	if rendered := fmt.Sprintf("%+v", captures); rendered == "" || strings.Contains(rendered, "super-secret") || strings.Contains(rendered, "Bearer") || strings.Contains(rendered, "text/plain") {
		t.Fatalf("capture summaries exposed response values: %s", rendered)
	}
	if removed := store.Clear(); removed != 2 || len(store.Captures()) != 0 {
		t.Fatalf("clear removed %d responses, captures=%v", removed, store.Captures())
	}
}
